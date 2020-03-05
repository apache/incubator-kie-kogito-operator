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

package test

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/version"
	v1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetImageStreams creates and gets an ImageStream and its ImageStreamTag for mocking purposes
func GetImageStreams(imageName, namespace, ownerName string) (*v1.ImageStream, *v1.ImageStreamTag) {
	image := fmt.Sprintf("quay.io/kiegroup/%s:%s", imageName, version.Version)
	is := &v1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:            imageName,
			Namespace:       namespace,
			OwnerReferences: []metav1.OwnerReference{{Name: ownerName}},
		},
		Spec: v1.ImageStreamSpec{
			LookupPolicy: v1.ImageLookupPolicy{Local: true},
			Tags: []v1.TagReference{
				{
					Name: version.Version,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: image,
					},
					ReferencePolicy: v1.TagReferencePolicy{Type: v1.LocalTagReferencePolicy},
				},
			},
		},
	}
	tag := &v1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s:%s", imageName, version.Version), Namespace: namespace},
		Image: v1.Image{
			DockerImageReference: image,
		},
	}
	return is, tag
}
