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
	"fmt"
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/version"
)

// BuildType which build can we perform? Supported are s2i and service
type BuildType string

const (
	// BuildTypeS2I source to image build type will take a source code and transform it into an executable service
	BuildTypeS2I BuildType = "s2i"
	// BuildTypeRuntime will create a image with a Kogito Service available
	BuildTypeRuntime BuildType = "runtime"
	// BuildTypeRuntimeJvm will create a image with JRE installed to run the Kogito Service
	BuildTypeRuntimeJvm BuildType = "runtime-jvm"
	// LabelKeyBuildType is the label key to identify the build type
	LabelKeyBuildType = "buildtype"
)

const (
	kindImageStreamTag = "ImageStreamTag"
	tagLatest          = "latest"
	// ImageStreamNamespace default namespace for the ImageStreams
	ImageStreamNamespace = "openshift"
)

var (
	// ImageStreamTag default tag version for the ImageStreams
	ImageStreamTag = version.Version
)

// BuildImageStreams are the image streams needed to perform the initial builds
var BuildImageStreams = map[BuildType]map[v1alpha1.RuntimeType]v1alpha1.Image{
	BuildTypeS2I: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-ubi8-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-ubi8-s2i",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	BuildTypeRuntime: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-ubi8",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
		v1alpha1.SpringbootRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-springboot-ubi8",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
	BuildTypeRuntimeJvm: {
		v1alpha1.QuarkusRuntimeType: v1alpha1.Image{
			ImageStreamName:      "kogito-quarkus-jvm-ubi8",
			ImageStreamNamespace: ImageStreamNamespace,
			ImageStreamTag:       ImageStreamTag,
		},
	},
}

// ensureImageBuild will check the build image parameters for emptiness and fill then with default values
func ensureImageBuild(image v1alpha1.Image, defaultImage v1alpha1.Image) v1alpha1.Image {
	if &image != nil {
		if len(image.ImageStreamTag) == 0 {
			image.ImageStreamTag = defaultImage.ImageStreamTag
		}
		if len(image.ImageStreamName) == 0 {
			image.ImageStreamName = defaultImage.ImageStreamName
		}
		if len(image.ImageStreamNamespace) == 0 {
			image.ImageStreamNamespace = defaultImage.ImageStreamNamespace
		}
		return image
	}

	return defaultImage
}

func parseImage(image *v1alpha1.Image) (string, string) {
	return fmt.Sprintf("%s:%s", image.ImageStreamName, image.ImageStreamTag), image.ImageStreamNamespace
}
