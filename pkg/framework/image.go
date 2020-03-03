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

package framework

import (
	"regexp"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

const (
	dockerTagRegx = `(?P<domain>(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]/)?(?P<namespace>.+/)?(?P<image>[^:]+)(?P<tag>:.+)?`
)

var (
	// DockerTagRegxCompiled is the compiled regex to verify docker tag names
	DockerTagRegxCompiled = *regexp.MustCompile(dockerTagRegx)
)

// ConvertImageTagToImage converts a plain string into an Image structure. For example, see https://regex101.com/r/1YX9rh/1.
func ConvertImageTagToImage(imageTag string) v1alpha1.Image {
	domain, ns, name, tag := SplitImageTag(imageTag)
	image := v1alpha1.Image{
		Domain:    domain,
		Namespace: ns,
		Name:      name,
		Tag:       tag,
	}

	return image
}

// ConvertImageToImageTag converts an Image into a plain string (domain/namespace/name:tag).
func ConvertImageToImageTag(image v1alpha1.Image) string {
	imageTag := ""
	if len(image.Domain) > 0 {
		imageTag += image.Domain + "/"
	}
	if len(image.Namespace) > 0 {
		imageTag += image.Namespace + "/"
	}
	imageTag += image.Name
	if len(image.Tag) > 0 {
		imageTag += ":" + image.Tag
	}
	return imageTag
}

// splitImageTag
func splitImageTag(imageTag string) (domain, namespace, name, tag string) {
	domain = ""
	namespace = ""
	name = ""
	tag = ""
	if len(imageTag) > 0 {
		if strings.HasPrefix(imageTag, ":") {
			tag = strings.Split(imageTag, ":")[1]
			return
		}

		imageMatch := DockerTagRegxCompiled.FindStringSubmatch(imageTag)
		if len(imageMatch[1]) > 0 {
			domain = strings.Split(imageMatch[1], "/")[0]
		}
		if len(imageMatch[2]) > 0 {
			namespace = strings.Split(imageMatch[2], "/")[0]
		}
		name = imageMatch[3]
		tag = strings.ReplaceAll(imageMatch[4], ":", "")
	}
	return
}

// SplitImageTag breaks into parts a given tag name, adds "latest" to the tag name if it's empty. For example, see https://regex101.com/r/1YX9rh/1.
func SplitImageTag(imageTag string) (domain, namespace, name, tag string) {
	if len(imageTag) == 0 {
		return
	}
	domain, namespace, name, tag = splitImageTag(imageTag)
	if len(tag) == 0 {
		tag = "latest"
	}
	return
}
