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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	imgv1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dockerImageKind = "DockerImage"
)

var imageStreamAnnotations = map[string]string{
	"openshift.io/provider-display-name": "Kie Group.",
	"openshift.io/display-name":          "Kogito Data Index service image",
}

var imageStreamTagAnnotations = map[string]string{
	"iconClass":   "icon-jbpm",
	"description": "Runtime image for Kogito Data Index",
	"tags":        "kogito,data-index",
}

func newImage(instance *v1alpha1.KogitoDataIndex) *imgv1.ImageStream {
	image := instance.Spec.Image
	if len(image) == 0 {
		image = DefaultDataIndexImage
	}
	_, _, _, tag := framework.SplitImageTag(image)

	return &imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{Name: DefaultDataIndexName, Namespace: instance.Namespace, Annotations: imageStreamAnnotations},
		Spec: imgv1.ImageStreamSpec{
			LookupPolicy: imgv1.ImageLookupPolicy{Local: true},
			Tags: []imgv1.TagReference{
				{
					Name:        tag,
					Annotations: imageStreamTagAnnotations,
					From: &corev1.ObjectReference{
						Kind: dockerImageKind,
						Name: image,
					},
					ReferencePolicy: imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy},
				},
			},
		},
	}
}
