// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package converter

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

// FromImageFlagToImage converts given ImageFlags into Image
func FromImageFlagToImage(imageFlags *flag.ImageFlags) v1alpha1.Image {
	return FromImageTagToImage(imageFlags.Image, "")
}

// FromImageTagToImage converts given image tag into Image
func FromImageTagToImage(imageTag, imageVersion string) v1alpha1.Image {
	image := framework.ConvertImageTagToImage(imageTag)

	// Apply image version only if image is empty
	// It is expected that user will provide image version in input image tag, only when user didn't provide image tag
	// and separately provides image version then default image domain/repo is picked at runtime and tag is used as user
	// provided image-version
	if image.IsEmpty() {
		image.Tag = imageVersion
	}
	return image
}
