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
				Env: []v1alpha1.Env{{Name: nativeBuildEnvVarKey, Value: "true"}},
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
				Native: false,
				Resources: v1alpha1.Resources{
					Limits: DefaultBuildS2IJVMLimits,
				},
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)

	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "false"})
	assert.NotContains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: nativeBuildEnvVarKey, Value: "true"})
	assert.Contains(t, bcRuntime.Spec.Strategy.SourceStrategy.From.Name, BuildImageStreams[BuildTypeRuntimeJvm][v1alpha1.QuarkusRuntimeType].ImageStreamName)
	assert.Equal(t, resource.MustParse(DefaultBuildS2IJVMCPULimit.Value), *bcS2I.Spec.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse(DefaultBuildS2IJVMMemoryLimit.Value), *bcS2I.Spec.Resources.Limits.Memory())
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
				ImageS2I: v1alpha1.ImageStream{
					ImageStreamTag:       "latest",
					ImageStreamNamespace: "openshift",
				},
				ImageRuntime: v1alpha1.ImageStream{
					ImageStreamName:      "my-image",
					ImageStreamNamespace: "openshift",
				},
				Native: true,
			},
		},
	}
	bcS2I, err := newBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcRuntime, err := newBuildConfigRuntime(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcRuntime)

	assert.Equal(t, "my-image:"+ImageStreamTag, bcRuntime.Spec.Strategy.SourceStrategy.From.Name)
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
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
				Native: true,
				Resources: v1alpha1.Resources{
					Limits: DefaultBuildS2INativeLimits,
				},
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
	assert.NotContains(t, bcRuntime.Spec.Output.To.Name, BuildS2INameSuffix)
	assert.Len(t, bcRuntime.Spec.Triggers, 1)
	assert.Len(t, bcS2I.Spec.Triggers, 0)
	assert.Equal(t, bcRuntime.Spec.Source.Images[0].From, *bcS2I.Spec.Output.To)
	assert.Equal(t, resource.MustParse(DefaultBuildS2INativeCPULimit.Value), *bcS2I.Spec.Resources.Limits.Cpu())
	assert.Equal(t, resource.MustParse(DefaultBuildS2INativeMemoryLimit.Value), *bcS2I.Spec.Resources.Limits.Memory())
	assert.Contains(t, bcS2I.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: buildS2IlimitMemoryEnvVarKey, Value: bcS2I.Spec.Resources.Limits.Memory().ToDec().AsDec().UnscaledBig().String()})
}

func Test_parseImage(t *testing.T) {
	type args struct {
		image *v1alpha1.ImageStream
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{"testParseImage",
			args{
				image: &v1alpha1.ImageStream{
					ImageStreamName:      "testImage",
					ImageStreamTag:       "vTest",
					ImageStreamNamespace: "testNamespace",
				},
			},
			"testImage:vTest",
			"testNamespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseImage(tt.args.image)
			if got != tt.want {
				t.Errorf("parseImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseImage() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
