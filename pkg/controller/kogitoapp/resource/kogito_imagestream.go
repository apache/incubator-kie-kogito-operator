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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	v1 "k8s.io/api/core/v1"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	imgv1 "github.com/openshift/api/image/v1"
)

// KogitoQuarkusUbi8Image quarkus runtime builder image
const KogitoQuarkusUbi8Image = "kogito-quarkus-ubi8"

// KogitoQuarkusJVMUbi8Image quarkus jvm runtime builder image
const KogitoQuarkusJVMUbi8Image = "kogito-quarkus-jvm-ubi8"

// KogitoQuarkusUbi8s2iImage quarkus s2i builder image
const KogitoQuarkusUbi8s2iImage = "kogito-quarkus-ubi8-s2i"

// KogitoSpringbootUbi8Image springboot runtime builder image
const KogitoSpringbootUbi8Image = "kogito-springboot-ubi8"

// KogitoSpringbootUbi8s2iImage springboot s2i builder image
const KogitoSpringbootUbi8s2iImage = "kogito-springboot-ubi8-s2i"

var (
	// ImageStreamNameList holds all kogito images, if one of them does not exists, it will be created
	ImageStreamNameList = [...]string{
		KogitoQuarkusUbi8Image,
		KogitoQuarkusJVMUbi8Image,
		KogitoQuarkusUbi8s2iImage,
		KogitoSpringbootUbi8Image,
		KogitoSpringbootUbi8s2iImage,
	}
)

// BuildImageStreams are the image streams needed to perform the initial builds
var BuildImageStreams = map[buildType]map[v1alpha1.RuntimeType]string{
	BuildTypeS2I: {
		v1alpha1.QuarkusRuntimeType:    KogitoQuarkusUbi8s2iImage,
		v1alpha1.SpringbootRuntimeType: KogitoSpringbootUbi8s2iImage,
	},
	BuildTypeRuntime: {
		v1alpha1.QuarkusRuntimeType:    KogitoQuarkusUbi8Image,
		v1alpha1.SpringbootRuntimeType: KogitoSpringbootUbi8Image,
	},
	BuildTypeRuntimeJvm: {
		v1alpha1.QuarkusRuntimeType: KogitoQuarkusJVMUbi8Image,
	},
}

// resolves the custom image stream name
func resolveCustomImageStreamName(tagImageName string) string {
	return fmt.Sprintf("custom-%s", tagImageName)
}

// newImageStream creates a new ImageStreamTag on the OpenShift cluster using the tag reference name.
// tagRefName refers to a full tag name like kogito-app:latest. If no tag is passed (e.g. kogito-app), "latest" will be used for the tag
func newImageStream(kogitoApp *v1alpha1.KogitoApp, tagRefName string) *imgv1.ImageStream {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, openshift.ImageTagLatest)
	}

	is := &imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      result[0],
			Namespace: kogitoApp.Namespace,
		},
		Spec: imgv1.ImageStreamSpec{
			LookupPolicy: imgv1.ImageLookupPolicy{
				Local: true,
			},
			Tags: []imgv1.TagReference{
				{
					Name: result[1],
					ReferencePolicy: imgv1.TagReferencePolicy{
						Type: imgv1.LocalTagReferencePolicy,
					},
				},
			},
		},
	}

	addDefaultMeta(&is.ObjectMeta, kogitoApp)
	meta.SetGroupVersionKind(&is.TypeMeta, meta.KindImageStream)

	return is
}

// CreateCustomKogitoImageStream creates a ImageStreamList based on the given image tag. Breaks down the image tag to extract image name and version number.
func CreateCustomKogitoImageStream(targetNamespace string, targetImageTag string) (imageList imgv1.ImageStreamList) {
	_, _, imageName, tagVersion := framework.SplitImageTag(targetImageTag)
	imageName = resolveCustomImageStreamName(imageName)
	imageStream := imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: targetNamespace,
		},
		Spec: imgv1.ImageStreamSpec{
			Tags: []imgv1.TagReference{
				{
					Name: tagVersion,
					ReferencePolicy: imgv1.TagReferencePolicy{
						Type: imgv1.LocalTagReferencePolicy,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: targetImageTag,
					},
				},
			},
		},
	}
	imageList = imgv1.ImageStreamList{}
	imageList.Items = append(imageList.Items, imageStream)
	return imageList
}

// CreateKogitoImageStream returns the ImageStreamList with all Kogito needed images.
// targetNamespace namespace where to set the ImageStream
// targetImageTag full image name for the image. Empty to use Operator default
// runtimeType Runtime type for this image stream. Ignored if targetImageTag is not empty
func CreateKogitoImageStream(kogitoApp *v1alpha1.KogitoApp, targetVersion string) imgv1.ImageStreamList {
	var kogitoImageStreamList imgv1.ImageStreamList

	for _, imageName := range ImageStreamNameList {
		create := false
		tagAnnotations := make(map[string]string)
		tagAnnotations["iconClass"] = "icon-jbpm"
		tagAnnotations["version"] = targetVersion

		imageStreamAnnotations := make(map[string]string)
		imageStreamAnnotations["openshift.io/provider-display-name"] = "Kie Group."

		switch imageName {
		case KogitoQuarkusUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on Quarkus native image"
			tagAnnotations["description"] = "Runtime image for Kogito based on Quarkus native image"
			tagAnnotations["tags"] = "runtime,kogito,quarkus"
			tagAnnotations["supports"] = "quarkus"
			if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType && kogitoApp.Spec.Build.Native {
				create = true
			}

		case KogitoQuarkusJVMUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on Quarkus JVM image"
			tagAnnotations["description"] = "Runtime image for Kogito based on Quarkus JVM image"
			tagAnnotations["tags"] = "runtime,kogito,quarkus,jvm"
			tagAnnotations["supports"] = "quarkus"
			if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType && !kogitoApp.Spec.Build.Native {
				create = true
			}

		case KogitoQuarkusUbi8s2iImage:
			imageStreamAnnotations["openshift.io/display-name"] = "Platform for building Kogito based on Quarkus"
			tagAnnotations["description"] = "Platform for building Kogito based on Quarkus"
			tagAnnotations["tags"] = "builder,kogito,quarkus"
			tagAnnotations["supports"] = "quarkus"
			if kogitoApp.Spec.Runtime == v1alpha1.QuarkusRuntimeType {
				create = true
			}

		case KogitoSpringbootUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on SpringBoot"
			tagAnnotations["description"] = "Runtime image for Kogito based on SpringBoot"
			tagAnnotations["tags"] = "runtime,kogito,springboot"
			tagAnnotations["supports"] = "springboot"
			if kogitoApp.Spec.Runtime == v1alpha1.SpringbootRuntimeType {
				create = true
			}

		case KogitoSpringbootUbi8s2iImage:
			imageStreamAnnotations["openshift.io/display-name"] = "Platform for building Kogito based on SpringBoot"
			tagAnnotations["description"] = "Platform for building Kogito based on SpringBoot"
			tagAnnotations["tags"] = "builder,kogito,springboot"
			tagAnnotations["supports"] = "springboot"
			if kogitoApp.Spec.Runtime == v1alpha1.SpringbootRuntimeType {
				create = true
			}
		}

		// if no build type is provided, add all imagestream
		if create || kogitoApp.Spec.Runtime == "" {
			currentImageStream := imgv1.ImageStream{
				ObjectMeta: metav1.ObjectMeta{
					Name:        imageName,
					Namespace:   kogitoApp.Namespace,
					Annotations: imageStreamAnnotations,
				},
				Spec: imgv1.ImageStreamSpec{
					Tags: []imgv1.TagReference{
						{
							Name:        targetVersion,
							Annotations: tagAnnotations,
							ReferencePolicy: imgv1.TagReferencePolicy{
								Type: imgv1.LocalTagReferencePolicy,
							},
							From: &v1.ObjectReference{
								Kind: "DockerImage",
								Name: fmt.Sprintf("quay.io/kiegroup/%s:%s", imageName, targetVersion),
							},
						},
					},
				},
			}
			kogitoImageStreamList.Items = append(kogitoImageStreamList.Items, currentImageStream)
		}
	}

	return kogitoImageStreamList
}

// GetImageStreamTagFromStream gets an ImageStreamTag reference from an ImageStream using tagName (e.g. 1.0.0) as an index
func GetImageStreamTagFromStream(tagName string, imageStream *imgv1.ImageStream) (imgStreamTag *imgv1.ImageStreamTag) {
	if imageStream == nil {
		return
	}

	for _, tag := range imageStream.Spec.Tags {
		if tag.Name == tagName {
			imgStreamTag = &imgv1.ImageStreamTag{
				ObjectMeta: metav1.ObjectMeta{Namespace: imageStream.Namespace, Name: fmt.Sprintf("%s:%s", imageStream.Name, tagName), Annotations: tag.Annotations},
				Tag:        &tag,
			}

		}
	}
	return
}
