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

package connector

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/internal"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path"
	"testing"
)

func TestMountProtoBufConfigMapOnDataIndex(t *testing.T) {
	fileName1 := "mydomain.proto"
	fileName2 := "mydomain2.proto"
	instance := test.CreateFakeDataIndex(t.Name())
	instance.SetUID(types.UID(uuid.New().String()))
	dc := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					UID: instance.UID,
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "test",
						},
					},
				},
			},
		},
	}

	runtimeService := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
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
			Annotations: map[string]string{
				infrastructure.FromFileKey:  "true",
				infrastructure.MountPathKey: path.Join(DefaultProtobufMountPath, runtimeService.GetName()),
				infrastructure.FileModeKey:  fmt.Sprint(framework.ModeForProtoBufConfigMapVolume),
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

	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	supportingServiceHandler := internal.NewKogitoSupportingServiceHandler(context)
	protoBufHandler := NewProtoBufHandler(context, supportingServiceHandler)
	err := protoBufHandler.MountProtoBufConfigMapOnDataIndex(runtimeService)
	assert.NoError(t, err)
	supportingServiceManager := manager.NewKogitoSupportingServiceManager(context, supportingServiceHandler)
	deployment, err := supportingServiceManager.FetchKogitoSupportingServiceDeployment(t.Name(), api.DataIndex)
	assert.NoError(t, err)

	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
}
