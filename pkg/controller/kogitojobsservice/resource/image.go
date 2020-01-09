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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	imgv1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultImageDomain    = "quay.io"
	defaultImageNamespace = "kiegroup"
	defaultImageName      = "kogito-jobs-service"
	defaultImageTag       = "latest"
	// DefaultImageTagName is the default full tag name for the Jobs Service image
	DefaultImageTagName = defaultImageDomain + "/" + defaultImageNamespace + "/" + defaultImageName + ":" + defaultImageTag

	dockerImageKind = "DockerImage"

	annotationKeyVersion = "version"
)

var imageStreamTagAnnotations = map[string]string{
	"iconClass":   "icon-jbpm",
	"description": "Runtime image for Kogito Jobs Service",
	"tags":        "kogito,jobs-service",
}

var imageStreamAnnotations = map[string]string{
	"openshift.io/provider-display-name": "Kie Group.",
	"openshift.io/display-name":          "Runtime image for Kogito based on Quarkus native image",
}

// ImageResolver helps resolving image structure
type ImageResolver interface {
	// ResolveImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest"
	ResolveImage() string
}

// NewImageResolver creates a new ImageResolver
func NewImageResolver(instance *v1alpha1.KogitoJobsService) ImageResolver {
	return &imageHandler{
		image: &instance.Spec.Image,
	}
}

// imageHandler defines the base structure for images in either OpenShift or Kubernetes clusters
type imageHandler struct {
	// imageStream is the created Image Stream for the service, should be nil on Kubernetes
	imageStream *imgv1.ImageStream
	// image is the CR structure attribute given by the user
	image *v1alpha1.Image
}

func (i *imageHandler) hasImageStream() bool {
	return i.imageStream != nil
}

func (i *imageHandler) ResolveImage() string {
	domain := i.image.Domain
	if len(domain) == 0 {
		domain = defaultImageDomain
	}
	ns := i.image.Namespace
	if len(ns) == 0 {
		ns = defaultImageNamespace
	}

	return fmt.Sprintf("%s/%s/%s", domain, ns, i.resolveImageNameTag())
}

// resolves like "kogito-jobs-service:latest"
func (i *imageHandler) resolveImageNameTag() string {
	name := i.image.Name
	if len(name) == 0 {
		name = defaultImageName
	}
	return fmt.Sprintf("%s:%s", name, i.resolveTag())
}

// resolves like "latest"
func (i *imageHandler) resolveTag() string {
	if len(i.image.Tag) == 0 {
		return defaultImageTag
	}
	return i.image.Tag
}

func newImageHandler(instance *v1alpha1.KogitoJobsService, cli *client.Client) *imageHandler {
	if &instance.Spec.Image == nil {
		instance.Spec.Image = v1alpha1.Image{}
	}
	handler := &imageHandler{
		image:       &instance.Spec.Image,
		imageStream: nil,
	}

	if !cli.IsOpenshift() {
		return handler
	}

	imageStreamTagAnnotations[annotationKeyVersion] = handler.resolveTag()
	handler.imageStream = &imgv1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: defaultImageName, Namespace: instance.Namespace, Annotations: imageStreamAnnotations},
		Spec: imgv1.ImageStreamSpec{
			LookupPolicy: imgv1.ImageLookupPolicy{Local: true},
			Tags: []imgv1.TagReference{
				{
					Name:        handler.resolveTag(),
					Annotations: imageStreamTagAnnotations,
					From: &corev1.ObjectReference{
						Kind: dockerImageKind,
						Name: handler.ResolveImage(),
					},
					ReferencePolicy: imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy},
				},
			},
		},
	}

	return handler
}
