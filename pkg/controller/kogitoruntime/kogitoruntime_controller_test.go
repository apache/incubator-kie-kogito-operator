// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kogitoruntime

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestReconcileKogitoRuntime_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "example-quarkus", Namespace: t.Name()},
		Spec: v1alpha1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:      &replicas,
				ServiceLabels: map[string]string{"process": "example-quarkus"},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)

	r := ReconcileKogitoRuntime{client: cli, scheme: meta.GetRegisteredSchema()}

	// first reconciliation
	test.AssertReconcileMustNotRequeue(t, &r, instance)
	// second time
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)

	// svc discovery
	svc := &corev1.Service{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	exists, err := kubernetes.ResourceC(cli).Fetch(svc)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, svc.Labels["process"] == instance.Name)

	// sa, namespace env var, volume count and protobuf
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(deployment)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.True(t, framework.GetEnvVarFromContainer("NAMESPACE", &deployment.Spec.Template.Spec.Containers[0]) == instance.Namespace)
	assert.Equal(t, "kogito-service-viewer", deployment.Spec.Template.Spec.ServiceAccountName)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 2) // #1 for property, #2 for downward api
	// command to register protobuf
	assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].Lifecycle.PostStart.Exec.Command, podStartExecCommand)

	configMap := &corev1.ConfigMap{ObjectMeta: v1.ObjectMeta{Name: getProtoBufConfigMapName(instance.Name), Namespace: instance.Namespace}}
	exists, err = kubernetes.ResourceC(cli).Fetch(configMap)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, getProtoBufConfigMapName(instance.Name), configMap.Name)
}

// see https://issues.redhat.com/browse/KOGITO-2535
func TestReconcileKogitoRuntime_CustomImage(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1alpha1.KogitoRuntimeSpec{
			Runtime: v1alpha1.SpringBootRuntimeType,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &replicas,
				Image: v1alpha1.Image{
					Domain:    "quay.io",
					Name:      "process-springboot-example-default",
					Namespace: "custom",
					Tag:       "latest",
				},
			},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	test.AssertReconcileMustNotRequeue(t, &ReconcileKogitoRuntime{client: cli, scheme: meta.GetRegisteredSchema()}, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)

	// image stream
	is := imagev1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(&is)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, is.Spec.Tags, 1)
	assert.Equal(t, "latest", is.Spec.Tags[0].Name)
	assert.Equal(t, "quay.io/custom/process-springboot-example-default:latest", is.Spec.Tags[0].From.Name)
}

func TestReconcileKogitoRuntime_CustomConfigMap(t *testing.T) {
	replicas := int32(1)
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{Namespace: t.Name(), Name: "mysuper-cm"},
	}
	instance := &v1alpha1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1alpha1.KogitoRuntimeSpec{
			Runtime: v1alpha1.SpringBootRuntimeType,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:            &replicas,
				PropertiesConfigMap: "mysuper-cm",
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance, cm}, nil, nil)
	// we take the ownership of the custom cm
	test.AssertReconcileMustRequeue(t, &ReconcileKogitoRuntime{client: cli, scheme: meta.GetRegisteredSchema()}, instance)
	// we requeue..
	test.AssertReconcileMustNotRequeue(t, &ReconcileKogitoRuntime{client: cli, scheme: meta.GetRegisteredSchema()}, instance)
	_, err := kubernetes.ResourceC(cli).Fetch(cm)
	assert.NoError(t, err)
	assert.NotEmpty(t, cm.OwnerReferences)
}
