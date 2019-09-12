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

package builder

import (
	"encoding/json"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/resource"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	dockerv10 "github.com/openshift/api/image/docker10"

	"github.com/stretchr/testify/assert"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	imgfake "github.com/openshift/client-go/image/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	uri       = "https://github.com/kiegroup/kogito-examples"
	kogitoApp = &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "testns",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI: &uri,
				},
			},
		},
	}
)

func TestBuildResources_CreateAllWithoutImage(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.DeploymentConfig{}, &buildv1.BuildConfig{})
	s.AddKnownTypes(imgv1.SchemeGroupVersion, &imgv1.ImageStreamTag{}, &imgv1.ImageStream{})

	inv, err := BuildOrFetchObjects(&Context{
		FactoryContext: resource.FactoryContext{
			Client: &client.Client{ControlCli: fake.NewFakeClient(), ImageCli: imgfake.NewSimpleClientset().ImageV1()},
		},
		KogitoApp: kogitoApp,
	})

	assert.Nil(t, err)
	assert.NotNil(t, inv)
	assert.Nil(t, inv.DeploymentConfig)
	assert.False(t, inv.DeploymentConfigStatus.IsNew)
	assert.NotNil(t, inv.BuildConfigS2I)
	assert.NotNil(t, inv.BuildConfigService)
	assert.True(t, inv.BuildConfigS2IStatus.IsNew)
	assert.True(t, inv.BuildConfigServiceStatus.IsNew)
}

func TestBuildResources_CreateAllSuccess(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.DeploymentConfig{}, &buildv1.BuildConfig{}, &routev1.Route{})
	s.AddKnownTypes(imgv1.SchemeGroupVersion, &imgv1.ImageStreamTag{}, &imgv1.ImageStream{})
	dockerImageRaw, err := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
			},
		},
	})
	isTag := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app:latest",
			Namespace: "testns",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	inv, err := BuildOrFetchObjects(&Context{
		FactoryContext: resource.FactoryContext{
			Client: &client.Client{ControlCli: fake.NewFakeClient(), ImageCli: imgfake.NewSimpleClientset(&isTag).ImageV1()},
		},
		KogitoApp: kogitoApp,
	})

	assert.Nil(t, err)
	assert.NotNil(t, inv)

	assert.NotNil(t, inv.BuildConfigS2I)
	assert.True(t, inv.BuildConfigS2IStatus.IsNew)

	assert.NotNil(t, inv.BuildConfigService)
	assert.True(t, inv.BuildConfigServiceStatus.IsNew)

	assert.NotNil(t, inv.DeploymentConfig)
	assert.True(t, inv.DeploymentConfigStatus.IsNew)

	assert.NotNil(t, inv.Service)
	assert.True(t, inv.ServiceStatus.IsNew)

	assert.NotNil(t, inv.Route)
	assert.True(t, inv.RouteStatus.IsNew)

	assert.Len(t, inv.DeploymentConfig.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, inv.DeploymentConfig.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, int32(8080))
}
