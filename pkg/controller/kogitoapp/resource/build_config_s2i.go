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
	"strconv"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"

	buildv1 "github.com/openshift/api/build/v1"

	corev1 "k8s.io/api/core/v1"
)

const (
	// BuildS2INameSuffix is the suffix added to the build s2i builds for the Kogito Service Runtime
	BuildS2INameSuffix           = "-builder"
	nativeBuildEnvVarKey         = "NATIVE"
	buildS2IlimitCPUEnvVarKey    = "LIMIT_CPU"
	buildS2IlimitMemoryEnvVarKey = "LIMIT_MEMORY"
	mavenMirrorURLEnvVar         = "MAVEN_MIRROR_URL"
	mavenGroupIDEnvVar           = "PROJECT_GROUP_ID"
	mavenArtifactIDEnvVar        = "PROJECT_ARTIFACT_ID"
	mavenArtifactVersionEnvVar   = "PROJECT_VERSION"
	mavenDownloadOutputEnvVar    = "MAVEN_DOWNLOAD_OUTPUT"
)

// newBuildConfigS2I creates a new build configuration for source to image (s2i) builds
func newBuildConfigS2I(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig, err error) {
	if kogitoApp.Spec.IsGitURIEmpty() {
		return buildConfig, errors.New("GitSource in the Kogito App Spec is required to create new build configurations")
	}

	buildConfig = defaultBuildConfigS2I(kogitoApp, false)
	buildConfig.Labels = map[string]string{
		LabelKeyBuildType:    string(BuildTypeS2I),
		LabelKeyBuildVariant: string(BuildVariantSource),
	}

	setBCS2ISource(kogitoApp, &buildConfig)
	return buildConfig, nil
}

// newBuildConfigS2IFromFile creates a new build configuration for source to image (s2i) builds having Kogito assets as input
func newBuildConfigS2IFromFile(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig, err error) {
	buildConfig = defaultBuildConfigS2I(kogitoApp, true)
	buildConfig.Labels = map[string]string{
		LabelKeyBuildType:    string(BuildTypeS2I),
		LabelKeyBuildVariant: string(BuildVariantBinary),
	}
	return buildConfig, nil
}

// defaultBuildConfigS2I returns the default configuration used in a build from source or build from file
func defaultBuildConfigS2I(kogitoApp *v1alpha1.KogitoApp, buildFromAsset bool) (buildConfig buildv1.BuildConfig) {
	buildConfig.Namespace = kogitoApp.Namespace
	buildConfig.Name = fmt.Sprintf("%s%s", kogitoApp.Name, BuildS2INameSuffix)

	s2iBaseImage := corev1.ObjectReference{
		Kind:      kindImageStreamTag,
		Namespace: kogitoApp.Namespace,
		Name:      resolveImageStreamTagNameForBuilds(kogitoApp, kogitoApp.Spec.Build.ImageS2ITag, BuildTypeS2I),
	}
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
		{
			Type:        buildv1.ImageChangeBuildTriggerType,
			ImageChange: &buildv1.ImageChangeTrigger{From: &s2iBaseImage},
		},
	}

	buildConfig.Spec.Resources = kogitoApp.Spec.Build.Resources
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}

	setBCS2IStrategy(kogitoApp, &buildConfig, s2iBaseImage, buildFromAsset)

	meta.SetGroupVersionKind(&buildConfig.TypeMeta, meta.KindBuildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)

	return buildConfig
}

func setBCS2ISource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.Type = buildv1.BuildSourceGit
	// Remove the trailing slash, as it will be removed by openshift
	buildConfig.Spec.Source.ContextDir = strings.TrimSuffix(kogitoApp.Spec.Build.GitSource.ContextDir, "/")
	buildConfig.Spec.Source.Git = &buildv1.GitBuildSource{
		URI: kogitoApp.Spec.Build.GitSource.URI,
		Ref: kogitoApp.Spec.Build.GitSource.Reference,
	}
}

func setBCS2IStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, s2iBaseImage corev1.ObjectReference, buildFromAsset bool) {
	envs := kogitoApp.Spec.Build.Envs
	if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType {
		envs = framework.EnvOverride(envs, corev1.EnvVar{Name: nativeBuildEnvVarKey, Value: strconv.FormatBool(kogitoApp.Spec.Build.Native)})
	}

	limitCPU, limitMemory := getBCS2ILimitsAsIntString(buildConfig)
	envs = framework.EnvOverride(envs, corev1.EnvVar{Name: buildS2IlimitCPUEnvVarKey, Value: limitCPU})
	envs = framework.EnvOverride(envs, corev1.EnvVar{Name: buildS2IlimitMemoryEnvVarKey, Value: limitMemory})

	if len(kogitoApp.Spec.Build.MavenMirrorURL) > 0 {
		log.Infof("Setting maven mirror url to %s", kogitoApp.Spec.Build.MavenMirrorURL)
		envs = framework.EnvOverride(envs, corev1.EnvVar{Name: mavenMirrorURLEnvVar, Value: kogitoApp.Spec.Build.MavenMirrorURL})
	}

	if kogitoApp.Spec.Build.EnableMavenDownloadOutput {
		log.Infof("Enable logging for transfer progress of downloading/uploading maven dependencies")
		envs = framework.EnvOverride(envs, corev1.EnvVar{Name: mavenDownloadOutputEnvVar, Value: strconv.FormatBool(kogitoApp.Spec.Build.EnableMavenDownloadOutput)})
	}

	// if user has provided a file, binary build should be used instead.
	if buildFromAsset {

		if len(kogitoApp.Spec.Build.Artifact.GroupID) > 0 {
			log.Debugf("Setting final generated artifact group id %s", kogitoApp.Spec.Build.Artifact.GroupID)
			envs = framework.EnvOverride(envs, corev1.EnvVar{Name: mavenGroupIDEnvVar, Value: kogitoApp.Spec.Build.Artifact.GroupID})
		}

		if len(kogitoApp.Spec.Build.Artifact.ArtifactID) > 0 {
			log.Debugf("Setting final generated artifact id %s", kogitoApp.Spec.Build.Artifact.ArtifactID)
			envs = framework.EnvOverride(envs, corev1.EnvVar{Name: mavenArtifactIDEnvVar, Value: kogitoApp.Spec.Build.Artifact.ArtifactID})
		}

		if len(kogitoApp.Spec.Build.Artifact.Version) > 0 {
			log.Debugf("Setting final generated artifact version %s", kogitoApp.Spec.Build.Artifact.Version)
			envs = framework.EnvOverride(envs, corev1.EnvVar{Name: mavenArtifactVersionEnvVar, Value: kogitoApp.Spec.Build.Artifact.Version})
		}

		buildConfig.Spec.Source.Type = buildv1.BuildSourceBinary
		// The comparator hits reconciliation if this are not set to empty values. TODO: fix on the operator-utils project
		buildConfig.Spec.Source.Binary = &buildv1.BinaryBuildSource{AsFile: ""}
		buildConfig.Spec.RunPolicy = buildv1.BuildRunPolicySerial
		// set it to an empty state, in this case we don't want OCP triggering the build for us.
		buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{}
	}

	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From:        s2iBaseImage,
		Env:         envs,
		Incremental: &kogitoApp.Spec.Build.Incremental,
	}
}

func getBCS2ILimitsAsIntString(buildConfig *buildv1.BuildConfig) (limitCPU, limitMemory string) {
	limitCPU = ""
	limitMemory = ""
	if buildConfig == nil || buildConfig.Spec.Resources.Limits == nil {
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
