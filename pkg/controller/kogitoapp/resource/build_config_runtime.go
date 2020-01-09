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

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	destinationDir   = "."
	runnerSourcePath = "/home/kogito/bin"
)

// newBuildConfigRuntime creates a new build configuration for Kogito services builds
func newBuildConfigRuntime(kogitoApp *v1alpha1.KogitoApp, fromBuild *buildv1.BuildConfig) (buildConfig buildv1.BuildConfig, err error) {
	if fromBuild == nil {
		err = errors.New("Impossible to create a runner build configuration without a s2i build definition")
		return buildConfig, err
	}

	image, buildType := resolveRuntimeImage(kogitoApp)

	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Name,
			Namespace: kogitoApp.Namespace,
			Labels: map[string]string{
				LabelKeyBuildType: string(buildType),
			},
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", kogitoApp.Name, tagLatest)}
	setBCRuntimeSource(kogitoApp, &buildConfig, fromBuild)
	setBCRuntimeStrategy(kogitoApp, &buildConfig, &image)
	setBCRuntimeTriggers(&buildConfig, fromBuild)
	meta.SetGroupVersionKind(&buildConfig.TypeMeta, meta.KindBuildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig, err
}

func resolveRuntimeImage(kogitoApp *v1alpha1.KogitoApp) (v1alpha1.ImageStream, BuildType) {
	buildType := BuildTypeRuntime
	if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType && !kogitoApp.Spec.Build.Native {
		buildType = BuildTypeRuntimeJvm
	}

	image := ensureImageBuild(kogitoApp.Spec.Build.ImageRuntime, BuildImageStreams[buildType][kogitoApp.Spec.Runtime])

	return image, buildType
}

func setBCRuntimeSource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.Type = buildv1.BuildSourceImage
	buildConfig.Spec.Source.Images = []buildv1.ImageSource{
		{
			From: *fromBuildConfig.Spec.Output.To,
			Paths: []buildv1.ImageSourcePath{
				{
					DestinationDir: destinationDir,
					SourcePath:     runnerSourcePath,
				},
			},
		},
	}
}

func setBCRuntimeStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, image *v1alpha1.ImageStream) {
	imageName, imageNamespace := parseImage(image)
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      imageName,
			Namespace: imageNamespace,
			Kind:      kindImageStreamTag,
		},
	}
}

func setBCRuntimeTriggers(buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
		{
			Type:        buildv1.ImageChangeBuildTriggerType,
			ImageChange: &buildv1.ImageChangeTrigger{From: fromBuildConfig.Spec.Output.To},
		},
	}
}
