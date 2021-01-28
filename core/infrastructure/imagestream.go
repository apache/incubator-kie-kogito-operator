package infrastructure

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
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

// ImageStreamHandler ...
type ImageStreamHandler interface {
	FetchImageStream(key types.NamespacedName) (*imgv1.ImageStream, error)
	MustFetchImageStream(key types.NamespacedName) (*imgv1.ImageStream, error)
	CreateImageStream(name, namespace, imageName, tag string, addFromReference, insecureImageRegistry bool) *imgv1.ImageStream
}

type imageStreamHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewImageStreamHandler ...
func NewImageStreamHandler(client *client.Client, log logger.Logger) ImageStreamHandler {
	return &imageStreamHandler{
		client: client,
		log:    log,
	}
}

// GetSharedDeployedImageStream gets the deployed ImageStream shared among Kogito Custom Resources
func (i *imageStreamHandler) FetchImageStream(key types.NamespacedName) (*imgv1.ImageStream, error) {
	imageStream := &imgv1.ImageStream{}
	if exists, err := kubernetes.ResourceC(i.client).FetchWithKey(key, imageStream); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	} else {
		i.log.Debug("Successfully fetch deployed kogito infra reference")
		return imageStream, nil
	}
}

// GetSharedDeployedImageStream gets the deployed ImageStream shared among Kogito Custom Resources
func (i *imageStreamHandler) MustFetchImageStream(key types.NamespacedName) (*imgv1.ImageStream, error) {
	if imageStream, err := i.FetchImageStream(key); err != nil {
		return nil, err
	} else if imageStream == nil {
		return nil, fmt.Errorf("image stream with name %s not found in namespace %s", key.Name, key.Namespace)
	} else {
		i.log.Debug("Successfully fetch deployed kogito infra reference")
		return imageStream, nil
	}
}

// createImageStream creates the ImageStream referencing the given namespace.
// Adds a docker image in the "From" reference based on the given image if `addFromReference` is set to `true`
func (i *imageStreamHandler) CreateImageStream(name, namespace, imageName, tag string, addFromReference, insecureImageRegistry bool) *imgv1.ImageStream {
	if i.client.IsOpenshift() {
		imageStreamTagAnnotations[annotationKeyVersion] = tag
		imageStream := &imgv1.ImageStream{
			ObjectMeta: v1.ObjectMeta{Name: name, Namespace: namespace, Annotations: imageStreamAnnotations},
			Spec: imgv1.ImageStreamSpec{
				LookupPolicy: imgv1.ImageLookupPolicy{Local: true},
				Tags: []imgv1.TagReference{
					{
						Name:            tag,
						Annotations:     imageStreamTagAnnotations,
						ReferencePolicy: imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy},
						ImportPolicy:    imgv1.TagImportPolicy{Insecure: insecureImageRegistry},
					},
				},
			},
		}
		if addFromReference {
			imageStream.Spec.Tags[0].From = &corev1.ObjectReference{
				Kind: dockerImageKind,
				Name: imageName,
			}
		}
		return imageStream
	}
	return nil
}
