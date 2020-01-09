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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imgv1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestCreateRequiredResources_OnOpenShift(t *testing.T) {
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: "job-service", Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			Replicas: int32(2),
			Image:    v1alpha1.Image{},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	resources, err := CreateRequiredResources(instance, cli)
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Len(t, resources, 4)

	deployment := framework.GetResource(reflect.TypeOf(appsv1.Deployment{}), instance.Name, resources).(*appsv1.Deployment)
	assert.NotNil(t, deployment)
	assert.Len(t, deployment.Annotations, 1)
	assert.Contains(t, deployment.Annotations[annotationKeyImageTriggers], instance.Name)

	imageStream := framework.GetResource(reflect.TypeOf(imgv1.ImageStream{}), defaultImageName, resources).(*imgv1.ImageStream)
	assert.NotNil(t, imageStream)
	assert.Contains(t, deployment.Annotations[annotationKeyImageTriggers], imageStream.Name)
}

func TestCreateRequiredResources_OnKubernetes(t *testing.T) {
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: "job-service", Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			Replicas: int32(2),
			Image:    v1alpha1.Image{},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)

	resources, err := CreateRequiredResources(instance, cli)
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Len(t, resources, 2)

	deployment := framework.GetResource(reflect.TypeOf(appsv1.Deployment{}), instance.Name, resources).(*appsv1.Deployment)
	assert.NotNil(t, deployment)
	assert.Empty(t, deployment.Annotations)
}
