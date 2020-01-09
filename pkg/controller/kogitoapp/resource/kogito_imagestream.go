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
	"github.com/kiegroup/kogito-cloud-operator/version"
	v1 "k8s.io/api/core/v1"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	imgv1 "github.com/openshift/api/image/v1"
)

const (
	// KogitoQuarkusUbi8Image quarkus runtime builder image
	KogitoQuarkusUbi8Image = "kogito-quarkus-ubi8"
	// KogitoQuarkusJVMUbi8Image quarkus jvm runtime builder image
	KogitoQuarkusJVMUbi8Image = "kogito-quarkus-jvm-ubi8"
	// KogitoQuarkusUbi8s2iImage quarkus s2i builder image
	KogitoQuarkusUbi8s2iImage = "kogito-quarkus-ubi8-s2i"
	// KogitoSpringbootUbi8Image springboot runtime builder image
	KogitoSpringbootUbi8Image = "kogito-springboot-ubi8"
	// KogitoSpringbootUbi8s2iImage springboot s2i builder image
	KogitoSpringbootUbi8s2iImage = "kogito-springboot-ubi8-s2i"
	// KogitoDataIndexImage data index image
	KogitoDataIndexImage = "kogito-data-index"
)

var (
	// ImageStreamTag default tag version for the ImageStreams
	ImageStreamTag = version.Version
	// ImageStreamNameList holds all kogito images, if one of them does not exists, it will be created
	ImageStreamNameList = [...]string{
		KogitoQuarkusUbi8Image,
		KogitoQuarkusJVMUbi8Image,
		KogitoQuarkusUbi8s2iImage,
		KogitoSpringbootUbi8Image,
		KogitoSpringbootUbi8s2iImage,
		KogitoDataIndexImage,
	}
)

// newImageStreamTag creates a new ImageStreamTag on the OpenShift cluster using the tag reference name.
// tagRefName refers to a full tag name like kogito-app:latest. If no tag is passed (e.g. kogito-app), "latest" will be used for the tag
func newImageStreamTag(kogitoApp *v1alpha1.KogitoApp, tagRefName string) *imgv1.ImageStream {
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

// KogitoImageStream returns the ImageStreamList with all Kogito needed images.
// It will be used to create the Kogito Images on the cluster in case it is missing.
func KogitoImageStream(targetNamespace string, targetTag string, buildType v1alpha1.RuntimeType, native bool) imgv1.ImageStreamList {

	var KogitoImageStreamList imgv1.ImageStreamList

	for _, imageName := range ImageStreamNameList {
		create := false
		tagAnnotations := make(map[string]string)
		tagAnnotations["iconClass"] = "icon-jbpm"
		tagAnnotations["version"] = targetTag

		imageStreamAnnotations := make(map[string]string)
		imageStreamAnnotations["openshift.io/provider-display-name"] = "Kie Group."

		switch imageName {
		case KogitoQuarkusUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on Quarkus native image"
			tagAnnotations["description"] = "Runtime image for Kogito based on Quarkus native image"
			tagAnnotations["tags"] = "runtime,kogito,quarkus"
			tagAnnotations["supports"] = "quarkus"
			if buildType == v1alpha1.QuarkusRuntimeType && native {
				create = true
			}

		case KogitoQuarkusJVMUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on Quarkus JVM image"
			tagAnnotations["description"] = "Runtime image for Kogito based on Quarkus JVM image"
			tagAnnotations["tags"] = "runtime,kogito,quarkus,jvm"
			tagAnnotations["supports"] = "quarkus"
			if buildType == v1alpha1.QuarkusRuntimeType && !native {
				create = true
			}

		case KogitoQuarkusUbi8s2iImage:
			imageStreamAnnotations["openshift.io/display-name"] = "Platform for building Kogito based on Quarkus"
			tagAnnotations["description"] = "Platform for building Kogito based on Quarkus"
			tagAnnotations["tags"] = "builder,kogito,quarkus"
			tagAnnotations["supports"] = "quarkus"
			if buildType == v1alpha1.QuarkusRuntimeType {
				create = true
			}

		case KogitoSpringbootUbi8Image:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for Kogito based on SpringBoot"
			tagAnnotations["description"] = "Runtime image for Kogito based on SpringBoot"
			tagAnnotations["tags"] = "runtime,kogito,springboot"
			tagAnnotations["supports"] = "springboot"
			if buildType == v1alpha1.SpringbootRuntimeType {
				create = true
			}

		case KogitoSpringbootUbi8s2iImage:
			imageStreamAnnotations["openshift.io/display-name"] = "Platform for building Kogito based on SpringBoot"
			tagAnnotations["description"] = "Platform for building Kogito based on SpringBoot"
			tagAnnotations["tags"] = "builder,kogito,springboot"
			tagAnnotations["supports"] = "springboot"
			if buildType == v1alpha1.SpringbootRuntimeType {
				create = true
			}

		case KogitoDataIndexImage:
			imageStreamAnnotations["openshift.io/display-name"] = "Runtime image for the Kogito Data Index Service"
			tagAnnotations["description"] = "Runtime image for the Kogito Data Index Service"
			tagAnnotations["tags"] = "kogito,data-index"
			if buildType == "" {
				create = true
			}
		}

		// if no build type is provided, add all imagestream
		if create || buildType == "" {
			currentImageStream := imgv1.ImageStream{
				ObjectMeta: metav1.ObjectMeta{
					Name:        imageName,
					Namespace:   targetNamespace,
					Annotations: imageStreamAnnotations,
				},
				Spec: imgv1.ImageStreamSpec{
					Tags: []imgv1.TagReference{
						{
							Name:        targetTag,
							Annotations: tagAnnotations,
							ReferencePolicy: imgv1.TagReferencePolicy{
								Type: imgv1.LocalTagReferencePolicy,
							},
							From: &v1.ObjectReference{
								Kind: "DockerImage",
								Name: fmt.Sprintf("quay.io/kiegroup/%s:%s", imageName, targetTag),
							},
						},
					},
				},
			}
			KogitoImageStreamList.Items = append(KogitoImageStreamList.Items, currentImageStream)
		}
	}

	return KogitoImageStreamList
}
