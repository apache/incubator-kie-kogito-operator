// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestInjectDataIndexURLIntoKogitoRuntime(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uuid.New().String()),
		},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoRuntime.Name, Namespace: kogitoRuntime.Namespace, OwnerReferences: []metav1.OwnerReference{{UID: kogitoRuntime.UID}}},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "test"}}},
			},
		},
	}
	dataIndexes := &v1beta1.KogitoSupportingServiceList{
		Items: []v1beta1.KogitoSupportingService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultDataIndexName,
					Namespace: ns,
				},
				Spec: v1beta1.KogitoSupportingServiceSpec{
					ServiceType: v1beta1.DataIndex,
				},
				Status: v1beta1.KogitoSupportingServiceStatus{
					KogitoServiceStatus: v1beta1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoRuntime, dataIndexes, dc).Build()

	err := InjectDataIndexURLIntoKogitoRuntimeServices(cli, ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(cli).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, dc.Spec.Template.Spec.Containers[0].Env, v1.EnvVar{Name: dataIndexHTTPRouteEnv, Value: expectedRoute})
}

func TestMountProtoBufConfigMapsOnDeployment(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs",
			Labels:    map[string]string{ConfigMapProtoBufEnabledLabelKey: "true"},
		},
		Data: map[string]string{
			fileName1: "This is a protobuf file",
			fileName2: "This is another file",
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(cm).OnOpenShift().Build()

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: DefaultDataIndexName},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "my-container",
						},
					},
				},
			},
		},
	}
	err := MountProtoBufConfigMapsOnDeployment(cli, deployment)
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

func TestMountProtoBufConfigMapOnDataIndex(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      DefaultDataIndexName,
			UID:       types.UID(uuid.New().String()),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.DataIndex,
		},
	}
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace, OwnerReferences: []metav1.OwnerReference{{UID: instance.UID}}},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "test"}}},
			},
		},
	}

	cm1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs1",
			},
		},
		Data: map[string]string{fileName1: "This is a protobuf file"},
	}
	cm2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs2",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs2",
			},
		},
		Data: map[string]string{fileName2: "This is a protobuf file"},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, dc, cm1, cm2).OnOpenShift().Build()

	runtimeService := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
		},
	}
	err := MountProtoBufConfigMapOnDataIndex(cli, runtimeService)
	assert.NoError(t, err)
	deployment, err := getSupportingServiceDeployment(t.Name(), cli, v1beta1.DataIndex)
	assert.NoError(t, err)

	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
}
