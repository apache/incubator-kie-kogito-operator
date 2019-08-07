package definitions

import (
	"errors"
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	destinationDir   = "."
	runnerSourcePath = "/home/kogito/bin"
)

func newBCRunner(kogitoApp *v1alpha1.KogitoApp, fromBuild *buildv1.BuildConfig, image v1alpha1.Image) (buildConfig buildv1.BuildConfig, err error) {
	if fromBuild == nil {
		err = errors.New("Impossible to create a runner build configuration without a s2i build definition")
		return buildConfig, err
	}
	// headers and base information
	buildConfig = buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Spec.Name,
			Namespace: kogitoApp.Namespace,
		},
	}
	buildConfig.Spec.Output.To = &corev1.ObjectReference{Kind: kindImageStreamTag, Name: fmt.Sprintf("%s:%s", kogitoApp.Spec.Name, tagLatest)}
	setBCRunnerSource(kogitoApp, &buildConfig, fromBuild)
	setBCRunnerStrategy(kogitoApp, &buildConfig, &image)
	setBCRunnerTriggers(&buildConfig, fromBuild)
	setGroupVersionKind(&buildConfig.TypeMeta, BuildConfigKind)
	addDefaultMeta(&buildConfig.ObjectMeta, kogitoApp)
	return buildConfig, err
}

func setBCRunnerSource(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
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

func setBCRunnerStrategy(kogitoApp *v1alpha1.KogitoApp, buildConfig *buildv1.BuildConfig, image *v1alpha1.Image) {
	buildConfig.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	buildConfig.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		From: corev1.ObjectReference{
			Name:      fmt.Sprintf("%s:%s", image.ImageStreamName, image.ImageStreamTag),
			Namespace: image.ImageStreamNamespace,
			Kind:      kindImageStreamTag,
		},
	}
}

func setBCRunnerTriggers(buildConfig *buildv1.BuildConfig, fromBuildConfig *buildv1.BuildConfig) {
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{
		{
			Type:        buildv1.ImageChangeBuildTriggerType,
			ImageChange: &buildv1.ImageChangeTrigger{From: fromBuildConfig.Spec.Output.To},
		},
	}
}
