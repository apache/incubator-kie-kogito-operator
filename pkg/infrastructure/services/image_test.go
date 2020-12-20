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

package services

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_imageHandler_resolveImageOnOpenShiftWithImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultJobsServiceImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	imageHandler, err := newImageHandler(instance, ServiceDefinition{DefaultImageName: infrastructure.DefaultJobsServiceImageName}, cli)
	assert.NoError(t, err)
	image, err := imageHandler.resolveImage()
	assert.NoError(t, err)
	// since we have imagestream and tag, we should see them here
	assert.Contains(t, image, infrastructure.DefaultJobsServiceImageName)
}

func Test_imageHandler_resolveImageOnOpenShiftNoImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)
	imageHandler, err := newImageHandler(instance, ServiceDefinition{DefaultImageName: infrastructure.DefaultJobsServiceImageName}, cli)
	assert.NoError(t, err)
	image, err := imageHandler.resolveImage()
	assert.NoError(t, err)
	// on OpenShift and no ImageStream? Bye!
	assert.Empty(t, image)
}

func Test_imageHandler_resolveImageOnKubernetes(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)
	imageHandler, err := newImageHandler(instance, ServiceDefinition{DefaultImageName: infrastructure.DefaultJobsServiceImageName}, cli)
	assert.NoError(t, err)
	image, err := imageHandler.resolveImage()
	assert.NoError(t, err)
	// we should always have an image available on Kubernetes
	assert.Contains(t, image, infrastructure.DefaultJobsServiceImageName)
}

func Test_imageHandler_newImageHandlerInsecureImageRegistry(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas:              &replicas,
				InsecureImageRegistry: true,
			},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)
	imageHandler, err := newImageHandler(instance, ServiceDefinition{DefaultImageName: infrastructure.DefaultJobsServiceImageName}, cli)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(imageHandler.imageStream.Spec.Tags))
	assert.True(t, imageHandler.imageStream.Spec.Tags[0].ImportPolicy.Insecure)
}
