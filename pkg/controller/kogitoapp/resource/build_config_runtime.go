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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
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
		err = errors.New("Impossible to create a runner build configuration without a s2i build definition ")
		return buildConfig, err
	}

	buildConfig = getBCRuntime(kogitoApp, kogitoApp.Name, BuildVariantSource)
	// source is the image produced by the s2i BuildConfig
	buildConfig.Spec.Source.Type = buildv1.BuildSourceImage
	buildConfig.Spec.Source.Images = []buildv1.ImageSource{
		{
			From: *fromBuild.Spec.Output.To,
			Paths: []buildv1.ImageSourcePath{
				{
					DestinationDir: destinationDir,
					SourcePath:     runnerSourcePath,
				},
			},
		},
	}
	// triggers once we have a change in the s2i produced image
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
		{
			Type:        buildv1.ImageChangeBuildTriggerType,
			ImageChange: &buildv1.ImageChangeTrigger{From: fromBuild.Spec.Output.To},
		},
	}

	return buildConfig, nil
}

func getBCRuntime(kogitoApp *v1alpha1.KogitoApp, bcName string, variant buildVariant) buildv1.BuildConfig {
	buildType := BuildTypeRuntime
	if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType && !kogitoApp.Spec.Build.Native {
		buildType = BuildTypeRuntimeJvm
	}
	buildConfig := buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bcName,
			Namespace: kogitoApp.Namespace,
			Labels: map[string]string{
				LabelKeyBuildType:    string(buildType),
				LabelKeyBuildVariant: string(variant),
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", kogitoApp.Name, tagLatest)},
				},
				Strategy: buildv1.BuildStrategy{
					Type: buildv1.SourceBuildStrategyType,
					SourceStrategy: &buildv1.SourceBuildStrategy{
						From: corev1.ObjectReference{
							Kind:      kindImageStreamTag,
							Namespace: kogitoApp.Namespace,
							Name:      resolveImageStreamTagNameForBuilds(kogitoApp, kogitoApp.Spec.Build.ImageRuntimeTag, buildType),
						},
					},
				},
			},
			RunPolicy: buildv1.BuildRunPolicySerial,
		},
	}
	meta.SetGroupVersionKind(&buildConfig.TypeMeta, meta.KindBuildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig
}
