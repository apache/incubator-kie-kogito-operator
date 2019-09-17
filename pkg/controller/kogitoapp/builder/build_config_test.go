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
	v1 "k8s.io/api/core/v1"
	"testing"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_BuidConfig_NonNativeBuild(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
				Native: false,
				// we'll try to trick the build
				Env: []v1alpha1.Env{{Name: nativeBuildEnvVarKey, Value: "true"}},
			},
		},
	}
	bcS2I, _ := NewBuildConfigS2I(kogitoApp)
	bcService, _ := NewBuildConfigService(kogitoApp, &bcS2I)

	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "false"})
	assert.NotContains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "true"})
	assert.Contains(t, bcService.Spec.Strategy.SourceStrategy.From.Name, BuildImageStreams[BuildTypeRuntimeJvm][v1alpha1.QuarkusRuntimeType].ImageStreamName)
}

func Test_BuildConfig_WithCustomImage(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
				ImageS2I: v1alpha1.Image{
					ImageStreamTag:       "latest",
					ImageStreamNamespace: "openshift",
				},
				ImageRuntime: v1alpha1.Image{
					ImageStreamName:      "my-image",
					ImageStreamNamespace: "openshift",
				},
				Native: true,
			},
		},
	}
	bcS2I, err := NewBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcService, err := NewBuildConfigService(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcService)

	assert.Equal(t, "my-image:"+ImageStreamTag, bcService.Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, "kogito-quarkus-ubi8-s2i:latest", bcS2I.Spec.Strategy.SourceStrategy.From.Name)
}

func Test_buildConfigResource_New(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
				Native: true,
			},
		},
	}
	bcS2I, err := NewBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcService, err := NewBuildConfigService(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcService)

	assert.Contains(t, bcS2I.Spec.Output.To.Name, BuildS2INameSuffix)
	assert.NotContains(t, bcService.Spec.Output.To.Name, BuildS2INameSuffix)
	assert.Len(t, bcService.Spec.Triggers, 1)
	assert.Len(t, bcS2I.Spec.Triggers, 0)
	assert.Equal(t, bcService.Spec.Source.Images[0].From, *bcS2I.Spec.Output.To)
}
