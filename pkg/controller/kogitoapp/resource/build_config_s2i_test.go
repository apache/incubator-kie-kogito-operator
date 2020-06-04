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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
)

func Test_getBCS2ILimitsAsIntString(t *testing.T) {
	type args struct {
		buildConfig *buildv1.BuildConfig
	}
	var tests = []struct {
		name            string
		args            args
		wantLimitCPU    string
		wantLimitMemory string
	}{
		{"With Limits", args{buildConfig: &buildv1.BuildConfig{
			Spec: buildv1.BuildConfigSpec{
				CommonSpec: buildv1.CommonSpec{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    resource.MustParse("1000m"),
							v1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
		}}, "1", "536870912"},
		{"With Limits half cpu", args{buildConfig: &buildv1.BuildConfig{
			Spec: buildv1.BuildConfigSpec{
				CommonSpec: buildv1.CommonSpec{
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			},
		}}, "500m", "536870912"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLimitCPU, gotLimitMemory := getBCS2ILimitsAsIntString(tt.args.buildConfig)
			if gotLimitCPU != tt.wantLimitCPU {
				t.Errorf("getBCS2ILimitsAsIntString() gotLimitCPU = %v, want %v", gotLimitCPU, tt.wantLimitCPU)
			}
			if gotLimitMemory != tt.wantLimitMemory {
				t.Errorf("getBCS2ILimitsAsIntString() gotLimitMemory = %v, want %v", gotLimitMemory, tt.wantLimitMemory)
			}
		})
	}
}

func TestNewBuildConfigS2I(t *testing.T) {
	uri := "http://example.git"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("500m"),
						v1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("250m"),
						v1.ResourceMemory: resource.MustParse("64Mi"),
					},
				},
				Incremental: true,
				GitSource: v1alpha1.GitSource{
					URI: uri,
				},
				Native: true,
			},
		},
	}

	bc, err := newBuildConfigS2I(kogitoApp)
	assert.NoError(t, err)
	assert.Equal(t, bc.Namespace, "test")
	assert.Equal(t, bc.Name, "test-builder")
	assert.Contains(t, bc.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{
		Name:  buildS2IlimitCPUEnvVarKey,
		Value: "500m",
	})
}

func TestNewBuildConfigS2IFromFile(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("500m"),
						v1.ResourceMemory: resource.MustParse("128Mi"),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("250m"),
						v1.ResourceMemory: resource.MustParse("64Mi"),
					},
				},
				Incremental: true,
				Native:      true,
			},
		},
	}

	bc, err := newBuildConfigS2IFromFile(kogitoApp)
	assert.NoError(t, err)
	assert.Equal(t, bc.Namespace, "test")
	assert.Equal(t, bc.Name, "test-builder")
	assert.Contains(t, bc.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{
		Name:  buildS2IlimitCPUEnvVarKey,
		Value: "500m",
	})
}

func Test_setBCS2IStrategy_withCustomArtifactDetails(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				Artifact: v1alpha1.Artifact{
					GroupID:    "com.mycompany",
					ArtifactID: "testproject",
					Version:    "2.0-SNAPSHOT",
				},
			},
		},
	}

	buildConfig := &buildv1.BuildConfig{}

	s2iBaseImage := corev1.ObjectReference{}

	setBCS2IStrategy(kogitoApp, buildConfig, s2iBaseImage, true)

	envs := buildConfig.Spec.Strategy.SourceStrategy.Env
	{
		contains, envVarValue := getBuildEnvVariable(mavenGroupIDEnvVar, envs)
		assert.True(t, contains)
		assert.Equal(t, "com.mycompany", envVarValue)
	}

	{
		contains, envVarValue := getBuildEnvVariable(mavenArtifactIDEnvVar, envs)
		assert.True(t, contains)
		assert.Equal(t, "testproject", envVarValue)
	}

	{
		contains, envVarValue := getBuildEnvVariable(mavenArtifactVersionEnvVar, envs)
		assert.True(t, contains)
		assert.Equal(t, "2.0-SNAPSHOT", envVarValue)
	}
}

func Test_setBCS2IStrategy_withDefaultArtifactDetails(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build:   &v1alpha1.KogitoAppBuildObject{},
		},
	}

	buildConfig := &buildv1.BuildConfig{}

	s2iBaseImage := corev1.ObjectReference{}

	setBCS2IStrategy(kogitoApp, buildConfig, s2iBaseImage, true)

	envs := buildConfig.Spec.Strategy.SourceStrategy.Env
	containsGroupID, _ := getBuildEnvVariable(mavenGroupIDEnvVar, envs)
	assert.False(t, containsGroupID)

	containsArtifactID, _ := getBuildEnvVariable(mavenArtifactIDEnvVar, envs)
	assert.False(t, containsArtifactID)

	containsVersion, _ := getBuildEnvVariable(mavenArtifactVersionEnvVar, envs)
	assert.False(t, containsVersion)
}

func getBuildEnvVariable(envVarName string, envs []corev1.EnvVar) (contains bool, envVarValue string) {
	for _, buildEnv := range envs {
		if buildEnv.Name == envVarName {
			return true, buildEnv.Value
		}
	}
	return false, ""
}

func Test_setBCS2IStrategy_mavenDownloadOutputEnable(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				EnableMavenDownloadOutput: true,
			},
		},
	}

	buildConfig := &buildv1.BuildConfig{}

	s2iBaseImage := corev1.ObjectReference{}

	setBCS2IStrategy(kogitoApp, buildConfig, s2iBaseImage, false)

	envs := buildConfig.Spec.Strategy.SourceStrategy.Env
	for _, buildEnv := range envs {
		if buildEnv.Name == mavenDownloadOutputEnvVar {
			assert.Equal(t, "true", buildEnv.Value)
		}
	}
}

func Test_setBCS2IStrategy_mavenDownloadOutputDisable(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build: &v1alpha1.KogitoAppBuildObject{
				EnableMavenDownloadOutput: false,
			},
		},
	}

	buildConfig := &buildv1.BuildConfig{}

	s2iBaseImage := corev1.ObjectReference{}

	setBCS2IStrategy(kogitoApp, buildConfig, s2iBaseImage, false)

	envs := buildConfig.Spec.Strategy.SourceStrategy.Env
	for _, buildEnv := range envs {
		if buildEnv.Name == mavenDownloadOutputEnvVar {
			assert.Fail(t, "Env variable "+mavenDownloadOutputEnvVar+" should not set.")
		}
	}
}
