package constants

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

const (
	// ImageStreamTag default tag name for the ImageStreams
	ImageStreamTag = "0.2.0"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
)

// RuntimeImageDefaults ...
var RuntimeImageDefaults = map[v1alpha1.RuntimeType][]v1alpha1.Image{
	v1alpha1.QuarkusRuntimeType: {
		{
			ImageStreamName:      "kogito-quarkus-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
			BuilderImage:         true,
		},
		{
			ImageStreamName:      "kogito-quarkus-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	v1alpha1.SpringbootRuntimeType: {
		{
			ImageStreamName:      "kogito-springboot-centos-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
			BuilderImage:         true,
		},
		{
			ImageStreamName:      "kogito-springboot-centos",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
}
