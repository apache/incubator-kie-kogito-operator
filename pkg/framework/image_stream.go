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

import "fmt"

const (
	annotationKeyImageTriggers         = "image.openshift.io/triggers"
	annotationValueImageTriggersFormat = "[{\"from\":{\"kind\":\"ImageStreamTag\",\"name\":\"%s\"},\"fieldPath\":\"spec.template.spec.containers[?(@.name==\\\"%s\\\")].image\"}]"
)

// ResolveImageStreamTriggerAnnotation creates a key and value combination for the ImageStream trigger to be linked with a Kubernetes Deployment
// this way, a Deployment resource can be attached to a ImageStream, like the DeploymentConfigs are.
// See: https://docs.openshift.com/container-platform/3.11/dev_guide/managing_images.html#image-stream-kubernetes-resources
// imageNameTag should be set in the format image-name:version
func ResolveImageStreamTriggerAnnotation(imageNameTag, containerName string) (key, value string) {
	key = annotationKeyImageTriggers
	value = fmt.Sprintf(annotationValueImageTriggersFormat, imageNameTag, containerName)
	return
}
