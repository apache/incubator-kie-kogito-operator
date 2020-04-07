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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
)

// buildType which build can we perform? Supported are s2i and service
type buildType string
type buildVariant string

const (
	// BuildTypeS2I source to image build type will take a source code and transform it into an executable service
	BuildTypeS2I buildType = "s2i"
	// BuildTypeRuntime will create a image with a Kogito Service available
	BuildTypeRuntime buildType = "runtime"
	// BuildTypeRuntimeJvm will create a image with JRE installed to run the Kogito Service
	BuildTypeRuntimeJvm buildType = "runtime-jvm"
	// LabelKeyBuildType is the label key to identify the build type
	LabelKeyBuildType = "buildtype"
	// BuildVariantSource ...
	BuildVariantSource buildVariant = "source"
	// BuildVariantBinary ...
	BuildVariantBinary buildVariant = "binary"
	// LabelKeyBuildVariant describes the build variant
	LabelKeyBuildVariant = "buildvariant"
)

const (
	kindImageStreamTag = "ImageStreamTag"
	tagLatest          = "latest"
)

func resolveImageStreamTagNameForBuilds(kogitoApp *v1alpha1.KogitoApp, imageTag string, buildType buildType) (imageName string) {
	imageVersion := kogitoApp.Spec.Build.ImageVersion
	if len(imageTag) > 0 {
		_, _, imageName, imageVersion = framework.SplitImageTag(imageTag)
		imageName = resolveCustomImageStreamName(imageName)
	} else {
		imageName = BuildImageStreams[buildType][kogitoApp.Spec.Runtime]
		if len(imageVersion) == 0 {
			imageVersion = infrastructure.GetRuntimeImageVersion()
		}
	}
	imageName = fmt.Sprintf("%s:%s", imageName, imageVersion)

	return
}
