package definitions

import (
	"errors"
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nameSuffix = "-builder"
)

// NewBuildConfigS2I creates a new build configuration for source to image (s2i) builds
func NewBuildConfigS2I(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig, err error) {
	if kogitoApp.Spec.Build == nil || kogitoApp.Spec.Build.GitSource == nil {
		return buildConfig, errors.New("GitSource in the Kogito App Spec is required to create new build configurations")
	}
	image := BuildImageStreams[BuildTypeS2I][kogitoApp.Spec.Runtime]
	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", kogitoApp.Spec.Name, nameSuffix),
			Namespace: kogitoApp.Namespace,
			Labels: map[string]string{
				LabelKeyBuildType: string(BuildTypeS2I),
			},
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", buildConfig.Name, tagLatest)}
	setBCS2ISource(kogitoApp, &buildConfig)
	setBCS2IStrategy(kogitoApp, &buildConfig, &image)
	SetGroupVersionKind(&buildConfig.TypeMeta, KindBuildConfig)
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
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", image.ImageStreamName, image.ImageStreamTag),
			Namespace: image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
		Env:         shared.FromEnvToEnvVar(kogitoApp.Spec.Build.Env),
		Incremental: &kogitoApp.Spec.Build.Incremental,
	}
}
