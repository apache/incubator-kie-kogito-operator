package constants

import "github.com/kiegroup/submarine-cloud-operator/pkg/apis/app/v1alpha1"

const (
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "1.0"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
)

// RuntimeImageDefaults ...
var RuntimeImageDefaults = map[v1alpha1.RuntimeType][]v1alpha1.Image{
	v1alpha1.QuarkusRuntimeType: {
		{
			ImageStreamName:      "kaas-quarkus-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
			BuilderImage:         true,
		},
		{
			ImageStreamName:      "kaas-quarkus-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	v1alpha1.SpringbootRuntimeType: {
		{
			ImageStreamName:      "kaas-springboot-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
			BuilderImage:         true,
		},
		{
			ImageStreamName:      "kaas-springboot-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
}
