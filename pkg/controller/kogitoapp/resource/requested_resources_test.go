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
	"encoding/json"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func createKogitoApp() *v1alpha1.KogitoApp {
	uri := "https://github.com/kiegroup/kogito-examples"
	return &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "testns",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: v1alpha1.GitSource{
					URI: uri,
				},
			},
		},
	}
}

func TestBuildResources_CreateAllWithoutImage(t *testing.T) {
	kogitoApp := createKogitoApp()
	client := test.CreateFakeClient(nil, nil, nil)
	resources, err := GetRequestedResources(kogitoApp, client)

	assert.Nil(t, err)
	assert.NotNil(t, resources)
	assert.Nil(t, resources.DeploymentConfig)
	assert.NotNil(t, resources.BuildConfigS2I)
	assert.NotNil(t, resources.BuildConfigRuntime)
	assert.NotNil(t, resources.ImageStreamS2I)
	assert.NotNil(t, resources.ImageStreamRuntime)
}

func TestBuildResources_CreateAllSuccess(t *testing.T) {
	kogitoApp := createKogitoApp()

	dockerImageRaw, err := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	})
	assert.Nil(t, err)
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
	client := test.CreateFakeClient([]runtime.Object{&isTag}, []runtime.Object{&isTag}, nil)
	log.Errorf("kogitoapp", kogitoApp.GetName())

	resources, err := GetRequestedResources(kogitoApp, client)

	assert.Nil(t, err)
	assert.NotNil(t, resources)

	assert.NotNil(t, resources.BuildConfigS2I)

	assert.NotNil(t, resources.BuildConfigRuntime)

	assert.NotNil(t, resources.ImageStreamS2I)

	assert.NotNil(t, resources.ImageStreamRuntime)

	assert.NotNil(t, resources.DeploymentConfig)

	assert.NotNil(t, resources.Service)

	assert.NotNil(t, resources.Route)

	assert.NotNil(t, resources.ServiceMonitor)

	assert.NotNil(t, resources.AppPropCM)

	assert.True(t, len(resources.AppPropContentHash) > 0)

	assert.Len(t, resources.DeploymentConfig.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, resources.DeploymentConfig.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, int32(8080))
}

func TestBuildResources_CreateAllWithoutAppPropConfigMap(t *testing.T) {
	kogitoApp := createKogitoApp()
	client := test.CreateFakeClient([]runtime.Object{
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app-properties",
				Namespace: "testns",
			},
			Data: map[string]string{
				"application.properties": "",
			},
		},
	},
		nil, nil)
	resources, err := GetRequestedResources(kogitoApp, client)

	assert.Nil(t, err)
	assert.NotNil(t, resources)
	assert.NotNil(t, resources.AppPropContentHash)
	assert.Nil(t, resources.AppPropCM)
}

func Test_buildConfigS2IBuilderWithGitUrl(t *testing.T) {
	blankBuildSpec := &builderChain{
		KogitoApp: &v1alpha1.KogitoApp{
			Spec: v1alpha1.KogitoAppSpec{
				Build: &v1alpha1.KogitoAppBuildObject{ImageS2ITag: tagLatest},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: t.Name()},
		},
		Resources: &KogitoAppResources{},
	}
	gitURLGiven := &builderChain{
		KogitoApp: &v1alpha1.KogitoApp{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: t.Name()},
			Spec:       v1alpha1.KogitoAppSpec{Build: &v1alpha1.KogitoAppBuildObject{GitSource: v1alpha1.GitSource{URI: "http://anything.com"}}}},
		Resources: &KogitoAppResources{},
	}

	chain := buildConfigS2IBuilder(blankBuildSpec)
	assert.NotNil(t, chain)
	assert.NotNil(t, chain.Resources.BuildConfigS2I)

	chain = buildConfigS2IBuilder(gitURLGiven)
	assert.NotNil(t, chain)
	assert.NotNil(t, chain.Resources.BuildConfigS2I)
}

func Test_buildConfigS2IBuilderWithoutGitUrl(t *testing.T) {
	blankBuildSpec := &builderChain{
		KogitoApp: &v1alpha1.KogitoApp{
			Spec: v1alpha1.KogitoAppSpec{
				Build: &v1alpha1.KogitoAppBuildObject{ImageS2ITag: tagLatest},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test", Namespace: t.Name()},
		},
		Resources: &KogitoAppResources{},
	}
	gitURLGiven := &builderChain{
		KogitoApp: &v1alpha1.KogitoApp{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: t.Name()},
			Spec:       v1alpha1.KogitoAppSpec{Build: &v1alpha1.KogitoAppBuildObject{}}},
		Resources: &KogitoAppResources{},
	}

	chain := buildConfigS2IBuilder(blankBuildSpec)
	assert.NotNil(t, chain)
	assert.NotNil(t, chain.Resources.BuildConfigS2I)

	chain = buildConfigS2IBuilder(gitURLGiven)
	assert.NotNil(t, chain)
	assert.NotNil(t, chain.Resources.BuildConfigS2I)
}
