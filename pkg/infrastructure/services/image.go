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
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	dockerImageKind      = "DockerImage"
	annotationKeyVersion = "version"
)

var imageStreamTagAnnotations = map[string]string{
	"iconClass":   "icon-jbpm",
	"description": "Runtime image for Kogito Service",
	"tags":        "kogito,services",
}

var imageStreamAnnotations = map[string]string{
	"openshift.io/provider-display-name": "KIE Group",
	"openshift.io/display-name":          "Kogito Service image",
}

// ImageHandler describes the handler structure to handle Kogito Services Images
type ImageHandler interface {
	GetImageStream() *imgv1.ImageStream
	HasImageStream() bool
}

// imageHandler defines the base structure for images in either OpenShift or Kubernetes clusters
type imageHandler struct {
	// imageStream is the created Image Stream for the service, should be nil on Kubernetes
	imageStream *imgv1.ImageStream
	// image is the CR structure attribute given by the user
	image *v1alpha1.Image
	// defaultImageName is the default image name for this service. Used to resolve the image from the Kogito Team registry when no custom image is given.
	defaultImageName string
	// imageStreamName name for the image stream that will handle image tags for the given instance
	imageStreamName string
	// namespace to fetch/create objects
	namespace string
	// client to handle API cluster calls
	client *client.Client
}

// GetImageStream gets the a reference for the ImageStream in this handler. Can be nil on non OpenShift clusters
func (i *imageHandler) GetImageStream() *imgv1.ImageStream {
	return i.imageStream
}

// HasImageStream verifies if an ImageStream has been created for this handler
func (i *imageHandler) HasImageStream() bool {
	return i.imageStream != nil
}

// resolveImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest" or "internal-registry/namespace/image:hash".
// Can be empty if on OpenShift and the ImageStream is not ready.
func (i *imageHandler) resolveImage() (string, error) {
	if i.client.IsOpenshift() {
		is := &imgv1.ImageStream{ObjectMeta: v1.ObjectMeta{Name: i.imageStreamName, Namespace: i.namespace}}
		if exists, err := kubernetes.ResourceC(i.client).Fetch(is); err != nil {
			return "", err
		} else if !exists {
			return "", nil
		}
		// the image is on an ImageStreamTag object
		for _, tag := range is.Spec.Tags {
			if tag.Name == i.resolveTag() {
				ist, err := openshift.ImageStreamC(i.client).FetchTag(
					types.NamespacedName{
						Name:      i.imageStreamName,
						Namespace: i.namespace,
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
		domain = infrastructure.DefaultImageRegistry
	}
	ns := i.image.Namespace
	if len(ns) == 0 {
		ns = infrastructure.DefaultImageNamespace
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
		return infrastructure.GetKogitoImageVersion()
	}
	return i.image.Tag
}

// createImageStream creates the ImageStream referencing the given namespace.
// Adds a docker image in the "From" reference based on the given image if `addFromReference` is set to `true`
func (i *imageHandler) createImageStream(namespace string, addFromReference, insecureImageRegistry bool) {
	if i.client.IsOpenshift() {
		imageStreamTagAnnotations[annotationKeyVersion] = i.resolveTag()
		i.imageStream = &imgv1.ImageStream{
			ObjectMeta: v1.ObjectMeta{Name: i.imageStreamName, Namespace: namespace, Annotations: imageStreamAnnotations},
			Spec: imgv1.ImageStreamSpec{
				LookupPolicy: imgv1.ImageLookupPolicy{Local: true},
				Tags: []imgv1.TagReference{
					{
						Name:            i.resolveTag(),
						Annotations:     imageStreamTagAnnotations,
						ReferencePolicy: imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy},
						ImportPolicy:    imgv1.TagImportPolicy{Insecure: insecureImageRegistry},
					},
				},
			},
		}
		if addFromReference {
			i.imageStream.Spec.Tags[0].From = &corev1.ObjectReference{
				Kind: dockerImageKind,
				Name: i.resolveRegistryImage(),
			}
		}
	}
}

// NewImageHandlerForBuiltServices creates a new handler for Kogito Services being built
func NewImageHandlerForBuiltServices(image *v1alpha1.Image, namespace string, cli *client.Client) ImageHandler {
	handler := &imageHandler{
		image:            image,
		imageStream:      nil,
		namespace:        namespace,
		imageStreamName:  image.Name,
		defaultImageName: image.Name,
		client:           cli,
	}
	if cli.IsOpenshift() {
		// creates the empty tag reference in the ImageStream since this handler is for services being built
		handler.createImageStream(namespace, false, false)
	}
	return handler
}

func newImageHandler(instance v1alpha1.KogitoService, definition ServiceDefinition, cli *client.Client) (*imageHandler, error) {
	if instance.GetSpec().GetImage() == nil {
		instance.GetSpec().SetImage(v1alpha1.Image{})
	}

	addDockerImageReference := !instance.GetSpec().GetImage().IsEmpty() || !definition.customService
	if len(instance.GetSpec().GetImage().Tag) == 0 && len(definition.DefaultImageTag) > 0 {
		instance.GetSpec().GetImage().Tag = definition.DefaultImageTag
	}
	if len(instance.GetSpec().GetImage().Name) == 0 {
		instance.GetSpec().GetImage().Name = definition.DefaultImageName
	}
	handler := &imageHandler{
		image:            instance.GetSpec().GetImage(),
		imageStream:      nil,
		defaultImageName: definition.DefaultImageName,
		imageStreamName:  definition.DefaultImageName,
		namespace:        instance.GetNamespace(),
		client:           cli,
	}
	if cli.IsOpenshift() {
		sharedImageStream, err := GetSharedDeployedImageStream(handler.imageStreamName, instance.GetNamespace(), cli)
		if err != nil {
			return nil, err
		}
		if sharedImageStream != nil {
			handler.imageStream = sharedImageStream
		} else {
			handler.createImageStream(instance.GetNamespace(), addDockerImageReference, instance.GetSpec().IsInsecureImageRegistry())
		}
	}
	return handler, nil
}

// GetSharedDeployedImageStream gets the deployed ImageStream shared among Kogito Custom Resources
func GetSharedDeployedImageStream(name, namespace string, cli *client.Client) (*imgv1.ImageStream, error) {
	deployedImageStream := &imgv1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: namespace},
	}
	if exists, err := kubernetes.ResourceC(cli).Fetch(deployedImageStream); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return deployedImageStream, nil
}

// AddSharedImageStreamToResources adds the shared ImageStream in the given resource map.
// Normally used during reconciliation phase to bring a not yet owned ImageStream to the deployed list.
func AddSharedImageStreamToResources(resources map[reflect.Type][]resource.KubernetesResource, name, ns string, cli *client.Client) error {
	if cli.IsOpenshift() {
		// is the image already there?
		for _, is := range resources[reflect.TypeOf(imgv1.ImageStream{})] {
			if is.GetName() == name &&
				is.GetNamespace() == ns {
				return nil
			}
		}
		// fetch the shared image
		sharedImageStream, err := GetSharedDeployedImageStream(name, ns, cli)
		if err != nil {
			return err
		}
		if sharedImageStream != nil {
			resources[reflect.TypeOf(imgv1.ImageStream{})] = append(resources[reflect.TypeOf(imgv1.ImageStream{})], sharedImageStream)
		}
	}
	return nil
}
