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
	"errors"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"strconv"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BuildS2INameSuffix is the suffix added to the build s2i builds for the Kogio Service Runtime
	BuildS2INameSuffix           = "-builder"
	nativeBuildEnvVarKey         = "NATIVE"
	buildS2IlimitCPUEnvVarKey    = "LIMIT_CPU"
	buildS2IlimitMemoryEnvVarKey = "LIMIT_MEMORY"
)

var (
	// DefaultBuildS2IJVMCPULimit is the default CPU limit for JVM s2i builds
	DefaultBuildS2IJVMCPULimit = v1alpha1.ResourceMap{Resource: v1alpha1.ResourceCPU, Value: "2"}
	// DefaultBuildS2IJVMMemoryLimit is the default Memory limit for JVM s2i builds
	DefaultBuildS2IJVMMemoryLimit = v1alpha1.ResourceMap{Resource: v1alpha1.ResourceMemory, Value: "2Gi"}
	// DefaultBuildS2IJVMLimits is the default resource limits for JVM s2i builds
	DefaultBuildS2IJVMLimits = []v1alpha1.ResourceMap{DefaultBuildS2IJVMCPULimit, DefaultBuildS2IJVMMemoryLimit}
	// DefaultBuildS2INativeCPULimit is the default CPU limit for Native s2i builds
	DefaultBuildS2INativeCPULimit = v1alpha1.ResourceMap{Resource: v1alpha1.ResourceCPU, Value: "2"}
	// DefaultBuildS2INativeMemoryLimit is the default Memory limit for Native s2i builds
	DefaultBuildS2INativeMemoryLimit = v1alpha1.ResourceMap{Resource: v1alpha1.ResourceMemory, Value: "10Gi"}
	// DefaultBuildS2INativeLimits is the default resource limits for Native s2i builds
	DefaultBuildS2INativeLimits = []v1alpha1.ResourceMap{DefaultBuildS2INativeCPULimit, DefaultBuildS2INativeMemoryLimit}
)

// newBuildConfigS2I creates a new build configuration for source to image (s2i) builds
func newBuildConfigS2I(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig, err error) {
	if kogitoApp.Spec.Build == nil || kogitoApp.Spec.Build.GitSource == nil {
		return buildConfig, errors.New("GitSource in the Kogito App Spec is required to create new build configurations")
	}

	image := resolveS2IImage(kogitoApp)

	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", kogitoApp.Name, BuildS2INameSuffix),
			Namespace: kogitoApp.Namespace,
			Labels: map[string]string{
				LabelKeyBuildType: string(BuildTypeS2I),
			},
		},
	}

	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{}
	buildConfig.Spec.Resources = shared.FromResourcesToResourcesRequirements(kogitoApp.Spec.Build.Resources)
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}
	setBCS2ISource(kogitoApp, &buildConfig)
	setBCS2IStrategy(kogitoApp, &buildConfig, &image)
	meta.SetGroupVersionKind(&buildConfig.TypeMeta, meta.KindBuildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)

	return buildConfig, nil
}

func resolveS2IImage(kogitoApp *v1alpha1.KogitoApp) v1alpha1.ImageStream {
	return ensureImageBuild(kogitoApp.Spec.Build.ImageS2I, BuildImageStreams[BuildTypeS2I][kogitoApp.Spec.Runtime])
}

func setBCS2ISource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.Type = buildv1.BuildSourceGit
	// Remove the trailing slash, as it will be removed by openshift
	buildConfig.Spec.Source.ContextDir = strings.TrimSuffix(kogitoApp.Spec.Build.GitSource.ContextDir, "/")
	buildConfig.Spec.Source.Git = &buildv1.GitBuildSource{
		URI: *kogitoApp.Spec.Build.GitSource.URI,
		Ref: kogitoApp.Spec.Build.GitSource.Reference,
	}
}

func setBCS2IStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, image *v1alpha1.ImageStream) {
	envs := shared.FromEnvToEnvVar(kogitoApp.Spec.Build.Env)
	if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType {
		envs = framework.EnvOverride(envs, corev1.EnvVar{Name: nativeBuildEnvVarKey, Value: strconv.FormatBool(kogitoApp.Spec.Build.Native)})
	}
	limitCPU, limitMemory := getBCS2ILimitsAsIntString(buildConfig)
	envs = framework.EnvOverride(envs, corev1.EnvVar{Name: buildS2IlimitCPUEnvVarKey, Value: limitCPU})
	envs = framework.EnvOverride(envs, corev1.EnvVar{Name: buildS2IlimitMemoryEnvVarKey, Value: limitMemory})

	imageName, imageNamespace := parseImage(image)

	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      imageName,
			Namespace: imageNamespace,
			Kind:      kindImageStreamTag,
		},
		Env:         envs,
		Incremental: &kogitoApp.Spec.Build.Incremental,
	}
}

func getBCS2ILimitsAsIntString(buildConfig *buildv1.BuildConfig) (limitCPU, limitMemory string) {
	limitCPU = ""
	limitMemory = ""
	if &buildConfig.Spec.Resources == nil || buildConfig.Spec.Resources.Limits == nil {
		return "", ""
	}

	limitMemoryInt, possible := buildConfig.Spec.Resources.Limits.Memory().AsInt64()
	if !possible {
		limitMemoryInt = buildConfig.Spec.Resources.Limits.Memory().ToDec().AsDec().UnscaledBig().Int64()
	}

	if limitMemoryInt > 0 {
		limitMemory = strconv.FormatInt(limitMemoryInt, 10)
	}

	limitCPU = buildConfig.Spec.Resources.Limits.Cpu().String()
	return limitCPU, limitMemory
}
