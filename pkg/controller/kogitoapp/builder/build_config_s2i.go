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
	"errors"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"strconv"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nameSuffix           = "-builder"
	nativeBuildEnvVarKey = "NATIVE"
)

// NewBuildConfigS2I creates a new build configuration for source to image (s2i) builds
func NewBuildConfigS2I(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig, err error) {
	if kogitoApp.Spec.Build == nil || kogitoApp.Spec.Build.GitSource == nil {
		return buildConfig, errors.New("GitSource in the Kogito App Spec is required to create new build configurations")
	}

	image := ensureImageBuild(kogitoApp.Spec.Build.ImageS2I, BuildImageStreams[BuildTypeS2I][kogitoApp.Spec.Runtime])

	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", kogitoApp.Name, nameSuffix),
			Namespace: kogitoApp.Namespace,
			Labels: map[string]string{
				LabelKeyBuildType: string(BuildTypeS2I),
			},
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}
	setBCS2ISource(kogitoApp, &buildConfig)
	setBCS2IStrategy(kogitoApp, &buildConfig, &image)
	meta.SetGroupVersionKind(&buildConfig.TypeMeta, meta.KindBuildConfig)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig, nil
}

func setBCS2ISource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Source.ContextDir = kogitoApp.Spec.Build.GitSource.ContextDir
	buildConfig.Spec.Source.Git = &buildv1.GitBuildSource{
		URI: *kogitoApp.Spec.Build.GitSource.URI,
		Ref: kogitoApp.Spec.Build.GitSource.Reference,
	}
}

func setBCS2IStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, image *v1alpha1.Image) {
	envs := shared.FromEnvToEnvVar(kogitoApp.Spec.Build.Env)
	if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType {
		envs = util.EnvOverride(envs, corev1.EnvVar{Name: nativeBuildEnvVarKey, Value: strconv.FormatBool(kogitoApp.Spec.Build.Native)})
	}

	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", image.ImageStreamName, image.ImageStreamTag),
			Namespace: image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
		Env:         envs,
		Incremental: &kogitoApp.Spec.Build.Incremental,
	}
}
