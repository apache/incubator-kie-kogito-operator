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
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
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
				// we'll try to trick the build
				Envs: []v1.EnvVar{{Name: nativeBuildEnvVarKey, Value: "true"}},
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				GitSource: v1alpha1.GitSource{
					URI:        uri,
					ContextDir: "process-quarkus-example",
				},
				Native:         false,
				MavenMirrorURL: "https://localhost.nexus:8080/public",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)

	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "false"})
	assert.NotContains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "true"})
	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: mavenMirrorURLEnvVar, Value: "https://localhost.nexus:8080/public"})
	assert.Contains(t, bcRuntime.Spec.Strategy.SourceStrategy.From.Name, BuildImageStreams[BuildTypeRuntimeJvm][v1alpha1.QuarkusRuntimeType])
	assert.Equal(t, resource.MustParse("2"), *bcS2I.Spec.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("2Gi"), *bcS2I.Spec.Resources.Limits.Memory())
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
				GitSource: v1alpha1.GitSource{
					URI:        uri,
					ContextDir: "process-quarkus-example",
				},
				ImageVersion:    "latest",
				ImageRuntimeTag: "quay.io/namespace/my-image:0.2",
				Native:          true,
			},
		},
	}
	bcS2I, err := newBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcRuntime, err := newBuildConfigRuntime(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcRuntime)

	assert.Equal(t, "custom-my-image:0.2", bcRuntime.Spec.Strategy.SourceStrategy.From.Name)
	assert.Equal(t, fmt.Sprintf("%s:%s", KogitoQuarkusUbi8s2iImage, "latest"), bcS2I.Spec.Strategy.SourceStrategy.From.Name)
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
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("2"),
						v1.ResourceMemory: resource.MustParse("10Gi"),
					},
				},
				GitSource: v1alpha1.GitSource{
					URI:        uri,
					ContextDir: "process-quarkus-example",
				},
				Native:         true,
				MavenMirrorURL: "https://localhost.nexus:8080/public",
			},
		},
	}
	bcS2I, err := newBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcRuntime, err := newBuildConfigRuntime(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcRuntime)

	assert.Contains(t, bcS2I.Spec.Output.To.Name, BuildS2INameSuffix)
	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: mavenMirrorURLEnvVar, Value: "https://localhost.nexus:8080/public"})
	assert.NotContains(t, bcRuntime.Spec.Output.To.Name, BuildS2INameSuffix)
	assert.Len(t, bcRuntime.Spec.Triggers, 1)
	assert.Len(t, bcS2I.Spec.Triggers, 1)
	assert.Equal(t, bcRuntime.Spec.Source.Images[0].From, *bcS2I.Spec.Output.To)
	assert.Equal(t, resource.MustParse("2"), *bcS2I.Spec.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse("10Gi"), *bcS2I.Spec.Resources.Limits.Memory())
	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: buildS2IlimitMemoryEnvVarKey, Value: bcS2I.Spec.Resources.Limits.Memory().ToDec().AsDec().UnscaledBig().String()})
}
