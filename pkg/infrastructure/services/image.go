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

package services

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	imgv1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dockerImageKind       = "DockerImage"
	annotationKeyVersion  = "version"
	defaultImageDomain    = "quay.io"
	defaultImageNamespace = "kiegroup"
)

var imageStreamTagAnnotations = map[string]string{
	"iconClass":   "icon-jbpm",
	"description": "Runtime image for Kogito Service",
	"tags":        "kogito,services",
}

var imageStreamAnnotations = map[string]string{
	"openshift.io/provider-display-name": "Kie Group.",
	"openshift.io/display-name":          "Kogito Service image",
}

// imageHandler defines the base structure for images in either OpenShift or Kubernetes clusters
type imageHandler struct {
	// imageStream is the created Image Stream for the service, should be nil on Kubernetes
	imageStream *imgv1.ImageStream
	// image is the CR structure attribute given by the user
	image *v1alpha1.Image
	// defaultImageName is the default image name for this service
	defaultImageName string
	// client to handle API cluster calls
	client *client.Client
}

func (i *imageHandler) hasImageStream() bool {
	return i.imageStream != nil
}

// resolveImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest" or "internal-registry/namespace/image:hash".
// Can be empty if on OpenShift and the ImageStream is not ready.
func (i *imageHandler) resolveImage() (string, error) {
	if i.client.IsOpenshift() {
		is := &imgv1.ImageStream{ObjectMeta: v1.ObjectMeta{Name: i.imageStream.Name, Namespace: i.imageStream.Namespace}}
		if exists, err := kubernetes.ResourceC(i.client).Fetch(is); err != nil {
			return "", err
		} else if !exists {
			return "", nil
		}
		// the image is on an ImageStreamTag object
		for _, tag := range is.Spec.Tags {
			if tag.From != nil && tag.From.Name == i.resolveRegistryImage() {
				ist, err := openshift.ImageStreamC(i.client).FetchTag(
					types.NamespacedName{
						Name:      i.defaultImageName,
						Namespace: i.imageStream.Namespace,
					}, i.resolveTag())
				if err != nil {
					return "", err
				} else if ist == nil {
					return "", nil
				}
				return ist.Image.DockerImageReference, nil
			}
		}
		return "", nil
	}

	return i.resolveRegistryImage(), nil
}

// resolveRegistryImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest", as informed by user.
func (i *imageHandler) resolveRegistryImage() string {
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
		name = i.defaultImageName
	}
	return fmt.Sprintf("%s:%s", name, i.resolveTag())
}

// resolves like "latest", 0.8.0, and so on
func (i *imageHandler) resolveTag() string {
	if len(i.image.Tag) == 0 {
		return infrastructure.GetRuntimeImageVersion()
	}
	return i.image.Tag
}

func newImageHandler(instance v1alpha1.KogitoService, defaultImageName string, defaultImageTag string, cli *client.Client) *imageHandler {
	if instance.GetSpec().GetImage() == nil {
		instance.GetSpec().SetImage(v1alpha1.Image{})
	}
	if len(instance.GetSpec().GetImage().Tag) == 0 && len(defaultImageTag) > 0 {
		instance.GetSpec().GetImage().Tag = defaultImageTag
	}
	handler := &imageHandler{
		image:            instance.GetSpec().GetImage(),
		imageStream:      nil,
		defaultImageName: defaultImageName,
		client:           cli,
	}

	if cli.IsOpenshift() {
		imageStreamTagAnnotations[annotationKeyVersion] = handler.resolveTag()
		handler.imageStream = &imgv1.ImageStream{
			ObjectMeta: v1.ObjectMeta{Name: defaultImageName, Namespace: instance.GetNamespace(), Annotations: imageStreamAnnotations},
			Spec: imgv1.ImageStreamSpec{
				LookupPolicy: imgv1.ImageLookupPolicy{Local: true},
				Tags: []imgv1.TagReference{
					{
						Name:        handler.resolveTag(),
						Annotations: imageStreamTagAnnotations,
						From: &corev1.ObjectReference{
							Kind: dockerImageKind,
							Name: handler.resolveRegistryImage(),
						},
						ReferencePolicy: imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy},
					},
				},
			},
		}
	}

	return handler
}
