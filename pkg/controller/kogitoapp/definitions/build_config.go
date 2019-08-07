package definitions

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
)

const (
	kindImageStreamTag = "ImageStreamTag"
	tagLatest          = "latest"
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "0.2.0"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
	// S2IBuildType source to image build type will take a source code and transform it into an executable service
	S2IBuildType BuildType = "s2i"
	// RunnerBuildType will create a image with a Kogito Service available
	RunnerBuildType BuildType = "runner"
)

// BuildImageStreams are the image streams needed to perform the initial builds
var BuildImageStreams = map[BuildType]map[v1alpha1.RuntimeType]v1alpha1.Image{
	S2IBuildType: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	RunnerBuildType: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
}

// BuildType which build can we perform? Supported are s2i and runner
type BuildType string

// BuildConfigComposition is the composition of the build configuration for the Kogito App
type BuildConfigComposition struct {
	BuildS2I    buildv1.BuildConfig
	BuildRunner buildv1.BuildConfig
	AsMap       map[BuildType]*buildv1.BuildConfig
}

// NewBuildConfig creates the BuildConfig resource structure for the KogitoApp CRD
func NewBuildConfig(kogitoApp *v1alpha1.KogitoApp) (buildConfig BuildConfigComposition, err error) {
	buildConfig = BuildConfigComposition{}

	if buildConfig.BuildS2I, err = newBCS2I(kogitoApp, BuildImageStreams[S2IBuildType][kogitoApp.Spec.Runtime]); err != nil {
		return buildConfig, err
	}
	if buildConfig.BuildRunner, err = newBCRunner(kogitoApp, &buildConfig.BuildS2I, BuildImageStreams[RunnerBuildType][kogitoApp.Spec.Runtime]); err != nil {
		return buildConfig, err
	}

	// transform the builds to a map to facilitate the redesign on controller side.
	// we should remove it after having inventory package to handle the objects
	buildConfig.AsMap = map[BuildType]*buildv1.BuildConfig{
		S2IBuildType:    &buildConfig.BuildS2I,
		RunnerBuildType: &buildConfig.BuildRunner,
	}
	return buildConfig, err
}
