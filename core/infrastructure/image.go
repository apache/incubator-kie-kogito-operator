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

package infrastructure

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/version"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	imgv1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

const (
	annotationKeyImageTriggers         = "image.openshift.io/triggers"
	annotationValueImageTriggersFormat = "[{\"from\":{\"kind\":\"ImageStreamTag\",\"name\":\"%s\"},\"fieldPath\":\"spec.template.spec.containers[?(@.name==\\\"%s\\\")].image\"}]"
	// LatestTag the default name for latest image tag
	LatestTag        = "latest"
	versionSeparator = "."
)

const (

	// DefaultImageRegistry the default services image repository
	DefaultImageRegistry = "quay.io"
	// DefaultImageNamespace the default services image namespace
	DefaultImageNamespace = "kiegroup"
	// KogitoQuarkusUbi8Image Quarkus runtime builder image
	KogitoQuarkusUbi8Image = "kogito-quarkus-ubi8"
	// KogitoQuarkusJVMUbi8Image Quarkus jvm runtime builder image
	KogitoQuarkusJVMUbi8Image = "kogito-quarkus-jvm-ubi8"
	// KogitoQuarkusUbi8s2iImage Quarkus s2i builder image
	KogitoQuarkusUbi8s2iImage = "kogito-quarkus-ubi8-s2i"
	// KogitoSpringBootUbi8Image SpringBoot runtime builder image
	KogitoSpringBootUbi8Image = "kogito-springboot-ubi8"
	// KogitoSpringBootUbi8s2iImage SpringBoot s2i builder image
	KogitoSpringBootUbi8s2iImage = "kogito-springboot-ubi8-s2i"
)

var (
	// KogitoImages maps the default Kogito Images on a matrix of RuntimeType and its purpose
	KogitoImages = map[api.RuntimeType]map[bool]string{
		api.QuarkusRuntimeType: {
			true:  KogitoQuarkusUbi8s2iImage,
			false: KogitoQuarkusJVMUbi8Image,
		},
		api.SpringBootRuntimeType: {
			true:  KogitoSpringBootUbi8s2iImage,
			false: KogitoSpringBootUbi8Image,
		},
	}
)

// ImageHandler describes the handler structure to handle Kogito Services Images
type ImageHandler interface {
	ResolveImage() (string, error)
	ResolveImageNameTag() string
	ResolveImageStreamTriggerAnnotation(containerName string) (key, value string)
	CreateImageStreamIfNotExists() (*imgv1.ImageStream, error)
}

// imageHandler defines the base structure for images in either OpenShift or Kubernetes clusters
type imageHandler struct {
	// image is the CR structure attribute given by the user
	image *api.Image
	// defaultImageName is the default image name for this service. Used to resolve the image from the Kogito Team registry when no custom image is given.
	defaultImageName string
	// imageStreamName name for the image stream that will handle image tags for the given instance
	imageStreamName string
	// namespace to fetch/create objects
	namespace             string
	addFromReference      bool
	insecureImageRegistry bool
	// client to handle API cluster calls
	client *client.Client
	log    logger.Logger
}

// NewImageHandler ...
func NewImageHandler(image *api.Image, defaultImageName, imageStreamName, namespace string, addFromReference, insecureImageRegistry bool, client *client.Client, log logger.Logger) ImageHandler {
	return &imageHandler{
		image:                 image,
		defaultImageName:      defaultImageName,
		imageStreamName:       imageStreamName,
		namespace:             namespace,
		addFromReference:      addFromReference,
		insecureImageRegistry: insecureImageRegistry,
		client:                client,
		log:                   log,
	}
}

func (i *imageHandler) CreateImageStreamIfNotExists() (*imgv1.ImageStream, error) {
	if i.client.IsOpenshift() {
		imageStreamHandler := NewImageStreamHandler(i.client, i.log)
		imageStream, err := imageStreamHandler.FetchImageStream(types.NamespacedName{Name: i.imageStreamName, Namespace: i.namespace})
		if err != nil {
			return nil, err
		}
		if imageStream == nil {
			imageStream = imageStreamHandler.CreateImageStream(i.imageStreamName, i.namespace, i.resolveRegistryImage(), i.resolveTag(), i.addFromReference, i.insecureImageRegistry)
		}
		return imageStream, nil
	}
	return nil, nil
}

// resolveImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest" or "internal-registry/namespace/image:hash".
// Can be empty if on OpenShift and the ImageStream is not ready.
func (i *imageHandler) ResolveImage() (string, error) {
	if i.client.IsOpenshift() {
		// the image is on an ImageStreamTag object
		ist, err := openshift.ImageStreamC(i.client).FetchTag(types.NamespacedName{Name: i.imageStreamName, Namespace: i.namespace}, i.resolveTag())
		if err != nil {
			return "", err
		} else if ist == nil {
			return "", nil
		}
		return ist.Image.DockerImageReference, nil
	}
	return i.resolveRegistryImage(), nil
}

// resolveRegistryImage resolves images like "quay.io/kiegroup/kogito-jobs-service:latest", as informed by user.
func (i *imageHandler) resolveRegistryImage() string {
	domain := i.image.Domain
	if len(domain) == 0 {
		domain = DefaultImageRegistry
	}
	ns := i.image.Namespace
	if len(ns) == 0 {
		ns = DefaultImageNamespace
	}
	return fmt.Sprintf("%s/%s/%s", domain, ns, i.ResolveImageNameTag())
}

// resolves like "kogito-jobs-service:latest"
func (i *imageHandler) ResolveImageNameTag() string {
	name := i.image.Name
	if len(name) == 0 {
		name = i.defaultImageName
	}
	return fmt.Sprintf("%s:%s", name, i.resolveTag())
}

// resolves like "latest", 0.8.0, and so on
func (i *imageHandler) resolveTag() string {
	if len(i.image.Tag) == 0 {
		return GetKogitoImageVersion()
	}
	return i.image.Tag
}

// ResolveImageStreamTriggerAnnotation creates a key and value combination for the ImageStream trigger to be linked with a Kubernetes Deployment
// this way, a Deployment resource can be attached to a ImageStream, like the DeploymentConfigs are.
// See: https://docs.openshift.com/container-platform/3.11/dev_guide/managing_images.html#image-stream-kubernetes-resources
// imageNameTag should be set in the format image-name:version
func (i *imageHandler) ResolveImageStreamTriggerAnnotation(containerName string) (key, value string) {
	imageNameTag := i.ResolveImageNameTag()
	key = annotationKeyImageTriggers
	value = fmt.Sprintf(annotationValueImageTriggersFormat, imageNameTag, containerName)
	return
}

// GetKogitoImageVersion gets the Kogito Runtime latest micro version based on the Operator current version
// E.g. Operator version is 0.9.0, the latest image version is 0.9.x-latest
func GetKogitoImageVersion() string {
	return getKogitoImageVersion(version.Version)
}

// unit test friendly unexported function
// in this case we are considering only micro updates, that's 0.9.0 -> 0.9, thus for 1.0.0 => 1.0
// in the future this should be managed with carefully if we desire a behavior like 1.0.0 => 1, that's minor upgrades
func getKogitoImageVersion(v string) string {
	if len(v) == 0 {
		return LatestTag
	}

	versionPrefix := strings.Split(v, versionSeparator)
	length := len(versionPrefix)
	if length > 0 {
		lastIndex := 2   // micro updates
		if length <= 2 { // guard against unusual cases
			lastIndex = length
		}
		return strings.Join(versionPrefix[:lastIndex], versionSeparator)
	}
	return LatestTag
}
