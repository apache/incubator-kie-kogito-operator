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

package shared

import (
	"regexp"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
)

const (
	dockerTagRegx = `(?P<namespace>.+/)?(?P<image>[^:]+)(?P<tag>:.+)?`
)

var (
	// DockerTagRegxCompiled is the compiled regex to verify docker tag names
	DockerTagRegxCompiled = *regexp.MustCompile(dockerTagRegx)
)

// FromStringToImage converts a plain string into an image. For example, see https://regex101.com/r/jl7MPD/2.
func FromStringToImage(imagetag string) v1alpha1.Image {
	image := v1alpha1.Image{}
	if len(imagetag) > 0 {
		if strings.HasPrefix(imagetag, ":") {
			image.ImageStreamTag = strings.Split(imagetag, ":")[1]
			return image
		}

		imageMatch := DockerTagRegxCompiled.FindStringSubmatch(imagetag)
		if len(imageMatch[1]) > 0 {
			imageNamespace := strings.Split(imageMatch[1], "/")
			image.ImageStreamNamespace = imageNamespace[len(imageNamespace)-2]
		}
		image.ImageStreamName = imageMatch[2]
		image.ImageStreamTag = strings.ReplaceAll(imageMatch[3], ":", "")
	}
	return image
}

// FromStringArrayToControllerEnvs converts a string array in the format of key=value pairs to the required type for the KogitoApp controller
func FromStringArrayToControllerEnvs(strings []string) []v1alpha1.Env {
	if strings == nil {
		return nil
	}
	var envs []v1alpha1.Env
	mapstr := util.FromStringsKeyPairToMap(strings)
	for k, v := range mapstr {
		envs = append(envs, v1alpha1.Env{Name: k, Value: v})
	}
	return envs
}

// FromStringArrayToControllerResourceMap ...
func FromStringArrayToControllerResourceMap(strings []string) []v1alpha1.ResourceMap {
	if strings == nil {
		return nil
	}
	var res []v1alpha1.ResourceMap
	mapstr := util.FromStringsKeyPairToMap(strings)
	for k, v := range mapstr {
		res = append(res, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceKind(k), Value: v})
	}
	return res
}

// ExtractResource reads a string array in the format memory=512M, cpu=1 and returns the value for a given kind
func ExtractResource(kind v1alpha1.ResourceKind, resources []string) string {
	for _, res := range resources {
		resKV := strings.Split(res, "=")
		if string(kind) == resKV[0] {
			return resKV[1]
		}
	}

	return ""
}
