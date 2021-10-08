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

package app

import (
	"context"
	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta2 "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcileKogitoRuntime_Reconcile(t *testing.T) {
	replicas := int32(1)

	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	kogitoInfinispan := test.CreateFakeKogitoInfinispan(t.Name())

	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "example-quarkus", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas:      &replicas,
				ServiceLabels: map[string]string{"process": "example-quarkus"},
				Infra: []string{
					kogitoKafka.GetName(),
					kogitoInfinispan.GetName(),
				},
			},
		},
	}

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, kogitoKafka, kogitoInfinispan).Build()
	r := NewKogitoRuntimeReconciler(cli, meta.GetRegisteredSchema())

	// first reconciliation
	test.AssertReconcileMustNotRequeue(t, r, instance)
	// second time
	test.AssertReconcileMustNotRequeue(t, r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, *instance.Status.Conditions, 2)

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
	// command to register protobuf does not exist anymore
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].Lifecycle)
}

// see https://issues.redhat.com/browse/KOGITO-2535
func TestReconcileKogitoRuntime_CustomImage(t *testing.T) {
	replicas := int32(1)
	ns := t.Name()
	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: ns, UID: test.GenerateUID()},
		Spec: v1beta1.KogitoRuntimeSpec{
			Runtime: api.SpringBootRuntimeType,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &replicas,
				Image:    "quay.io/kiegroup/process-springboot-example-default:latest",
			},
		},
	}
	is, tag := test.CreateFakeImageStreams("process-springboot-example-default", ns, "latest")
	err := framework.AddOwnerReference(instance, meta.GetRegisteredSchema(), is)
	assert.NoError(t, err)

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, is).AddImageObjects(tag).OnOpenShift().Build()

	r := NewKogitoRuntimeReconciler(cli, meta.GetRegisteredSchema())
	test.AssertReconcileMustNotRequeue(t, r, instance)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, *instance.Status.Conditions, 2)

	// image stream
	is = &imagev1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example-default", Namespace: instance.Namespace},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(is)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, is.Spec.Tags, 1)
	assert.Equal(t, "latest", is.Spec.Tags[0].Name)
	assert.Equal(t, "quay.io/kiegroup/process-springboot-example-default:latest", is.Spec.Tags[0].From.Name)
}

// see https://issues.redhat.com/browse/KOGITO-2535
func TestReconcileKogitoRuntime_InvalidCustomImage(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			Runtime: api.SpringBootRuntimeType,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &replicas,
				Image:    "quay.io/custom/process-springboot-example-default-invalid:latest",
			},
		},
	}
	imageStream := &imagev1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example-default-invalid", Namespace: t.Name()},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: "latest",
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: "quay.io/custom/process-springboot-example-default-invalid:latest",
					},
				},
			},
		},
		Status: imagev1.ImageStreamStatus{
			Tags: []imagev1.NamedTagEventList{
				{
					Tag: "latest",
					Conditions: []imagev1.TagEventCondition{
						{
							Type:    imagev1.ImportSuccess,
							Status:  corev1.ConditionFalse,
							Reason:  "UnAuthorized",
							Message: "you may not have access to the container image quay.io/custom/process-springboot-example-default-invalid:latest",
						},
					},
				},
			},
		},
	}
	err := framework.AddOwnerReference(instance, meta.GetRegisteredSchema(), imageStream)
	assert.NoError(t, err)
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, imageStream).OnOpenShift().Build()
	r := NewKogitoRuntimeReconciler(cli, meta.GetRegisteredSchema())

	_, err = r.Reconcile(context.TODO(), reconcile.Request{NamespacedName: types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}})
	assert.Error(t, err)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, *instance.Status.Conditions, 3)
	failedCondition := meta2.FindStatusCondition(*instance.Status.Conditions, string(api.FailedConditionType))
	assert.Equal(t, v1.ConditionTrue, failedCondition.Status)
	assert.Equal(t, "you may not have access to the container image quay.io/custom/process-springboot-example-default-invalid:latest", failedCondition.Message)
}

func TestReconcileKogitoRuntime_CustomConfigMap(t *testing.T) {
	replicas := int32(1)
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "mysuper-cm",
		},
		Data: map[string]string{
			"application.properties": "key1=value1",
		},
	}
	instance := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{Name: "process-springboot-example", Namespace: t.Name()},
		Spec: v1beta1.KogitoRuntimeSpec{
			Runtime: api.SpringBootRuntimeType,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas:            &replicas,
				PropertiesConfigMap: "mysuper-cm",
			},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, cm).Build()
	r := NewKogitoRuntimeReconciler(cli, meta.GetRegisteredSchema())
	test.AssertReconcileMustNotRequeue(t, r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(cm)
	assert.NoError(t, err)
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	exists, err := kubernetes.ResourceC(cli).Fetch(deployment)
	assert.NoError(t, err)
	assert.True(t, exists)

	configMapMounted := false
	container := &deployment.Spec.Template.Spec.Containers[0]
	for _, volumeMount := range container.VolumeMounts {
		if volumeMount.Name == "mysuper-cm" {
			configMapMounted = true
		}
	}
	assert.True(t, configMapMounted)
}
