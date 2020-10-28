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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
	"testing"
)

func TestKogitoSupportingServiceDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	kogitoInfinispan := test.CreateFakeKogitoInfinispan(t.Name())

	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		// We don't need to specify that we need Infinispan, it will figure out that alone :)
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Infra: []string{
					kogitoKafka.Name,
					kogitoInfinispan.Name,
				},
			},
		},
	}

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, kogitoKafka, kogitoInfinispan).OnOpenShift().Build()
	r := &ReconcileKogitoDataIndex{
		cli,
		meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	// basic checks
	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
}

func TestReconcileKogitoSupportingServiceDataIndex_mountProtoBufConfigMaps(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{
			fileName1: "This is a protobuf file",
			fileName2: "This is another file",
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(cm).OnOpenShift().Build()
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultDataIndexName},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}
	err := mountProtoBufConfigMaps(deployment, cli)
	assert.NoError(t, err)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes[0].Name, cm.Name)
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 2)

	// we need to have them ordered to be able to do the appropriate comparision.
	sort.Slice(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, func(i, j int) bool {
		return deployment.Spec.Template.Spec.Containers[0].VolumeMounts[i].SubPath < deployment.Spec.Template.Spec.Containers[0].VolumeMounts[j].SubPath
	})
	assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name, cm.Name)
	assert.Equal(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name, cm.Name)
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].SubPath, fileName1)
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath, fileName1)
}

func TestReconcileKogitoSupportingServiceDataIndex_MultipleProtoBufCMs(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: infrastructure.DefaultDataIndexName},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
	}
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{fileName1: "This is a protobuf file"},
	}
	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs2",
			Labels:    map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{fileName2: "This is a protobuf file"},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, cm1, cm2).OnOpenShift().Build()
	r := &ReconcileKogitoDataIndex{
		cli,
		meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	result, err := r.Reconcile(req)
	assert.NotNil(t, result)
	assert.False(t, result.Requeue)
	assert.NoError(t, err)
}
