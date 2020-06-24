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

package build

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
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

	customKogitoImagePrefix = "custom-"
	labelKeyVersion         = "version"
)

var (
	// kogitoImages maps the default Kogito Images on a matrix of RuntimeType and its purpose
	kogitoImages = map[v1alpha1.RuntimeType]map[bool]string{
		v1alpha1.QuarkusRuntimeType: {
			true:  KogitoQuarkusUbi8s2iImage,
			false: KogitoQuarkusJVMUbi8Image,
		},
		v1alpha1.SpringbootRuntimeType: {
			true:  KogitoSpringBootUbi8s2iImage,
			false: KogitoSpringBootUbi8Image,
		},
	}

	// imageStreamDefaultAnnotations lists the default annotations for ImageStreams
	imageStreamDefaultAnnotations = map[string]map[string]string{
		KogitoQuarkusUbi8Image: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Runtime image for Kogito based on Quarkus native image",
		},
		KogitoQuarkusJVMUbi8Image: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Runtime image for Kogito based on Quarkus JVM image",
		},
		KogitoQuarkusUbi8s2iImage: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Platform for building Kogito based on Quarkus",
		},
		KogitoSpringBootUbi8Image: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Runtime image for Kogito based on SpringBoot",
		},
		KogitoSpringBootUbi8s2iImage: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Platform for building Kogito based on SpringBoot",
		},
	}

	//tagDefaultAnnotations lists the default annotations for ImageStreamTags
	tagDefaultAnnotations = map[string]map[string]string{
		KogitoQuarkusUbi8Image: {
			"iconClass":   "icon-jbpm",
			"description": "Runtime image for Kogito based on Quarkus native image",
			"tags":        "runtime,kogito,quarkus",
			"supports":    "quarkus",
		},
		KogitoQuarkusJVMUbi8Image: {
			"iconClass":   "icon-jbpm",
			"description": "Runtime image for Kogito based on Quarkus JVM image",
			"tags":        "runtime,kogito,quarkus,jvm",
			"supports":    "quarkus",
		},
		KogitoQuarkusUbi8s2iImage: {
			"iconClass":   "icon-jbpm",
			"description": "Platform for building Kogito based on Quarkus",
			"tags":        "builder,kogito,quarkus",
			"supports":    "quarkus",
		},
		KogitoSpringBootUbi8Image: {
			"iconClass":   "icon-jbpm",
			"description": "Runtime image for Kogito based on SpringBoot",
			"tags":        "runtime,kogito,springboot",
			"supports":    "springboot",
		},
		KogitoSpringBootUbi8s2iImage: {
			"iconClass":   "icon-jbpm",
			"description": "Platform for building Kogito based on SpringBoot",
			"tags":        "builder,kogito,springboot",
			"supports":    "springboot",
		},
	}
)

// resolveKogitoImageNameTag resolves the ImageStreamTag to be used in the given build, e.g. kogito-quarkus-ubi8-s2i:0.11
func resolveKogitoImageNameTag(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	return strings.Join([]string{
		resolveKogitoImageName(build, isBuilder),
		resolveKogitoImageTag(build, isBuilder),
	}, ":")
}

// resolveKogitoImageTag resolves the ImageTag to be used in the given build, e.g. 0.11
func resolveKogitoImageTag(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	image := &build.Spec.RuntimeImage
	if isBuilder {
		image = &build.Spec.BuildImage
	}
	if len(image.Tag) > 0 {
		return image.Tag
	}
	return infrastructure.GetKogitoImageVersion()
}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageName(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	image := &build.Spec.RuntimeImage
	if isBuilder {
		image = &build.Spec.BuildImage
	}
	if len(image.Name) > 0 {
		return image.Name
	}
	imageName := kogitoImages[build.Spec.Runtime][isBuilder]
	if build.Spec.Native && !isBuilder {
		imageName = KogitoQuarkusUbi8Image
	}
	return imageName
}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageStreamName(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	imageName := resolveKogitoImageName(build, isBuilder)
	image := &build.Spec.RuntimeImage
	if isBuilder {
		image = &build.Spec.BuildImage
	}
	if len(image.Name) > 0 { // custom image
		return strings.Join([]string{customKogitoImagePrefix, imageName}, "")
	}
	return imageName
}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageStreamTagName(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	imageStream := resolveKogitoImageStreamName(build, isBuilder)
	imageTag := resolveKogitoImageTag(build, isBuilder)
	return strings.Join([]string{imageStream, imageTag}, ":")
}

// resolveImageRegistry resolves the registry/namespace name to be used in the given build, e.g. quay.io/kiegroup
func resolveKogitoImageRegistryNamespace(build *v1alpha1.KogitoBuild, isBuilder bool) string {
	namespace := infrastructure.DefaultImageNamespace
	registry := infrastructure.DefaultImageRegistry
	image := &build.Spec.RuntimeImage
	if isBuilder {
		image = &build.Spec.BuildImage
	}
	if len(image.Domain) > 0 {
		registry = image.Domain
	}
	if len(image.Namespace) > 0 {
		namespace = image.Namespace
	}
	return strings.Join([]string{registry, namespace}, "/")
}

// newKogitoImageStreamForBuilders same as newKogitoImageStream(build, true)
func newKogitoImageStreamForBuilders(build *v1alpha1.KogitoBuild) imgv1.ImageStream {
	return newKogitoImageStream(build, true)
}

// newKogitoImageStreamForRuntime same as newKogitoImageStream(build, false)
func newKogitoImageStreamForRuntime(build *v1alpha1.KogitoBuild) imgv1.ImageStream {
	return newKogitoImageStream(build, false)
}

// newKogitoImageStream creates a new OpenShift ImageStream based on the given build and the image purpose
func newKogitoImageStream(build *v1alpha1.KogitoBuild, isBuilder bool) imgv1.ImageStream {
	imageStreamName := resolveKogitoImageStreamName(build, isBuilder)
	imageTag := resolveKogitoImageTag(build, isBuilder)
	imageRegistry := resolveKogitoImageRegistryNamespace(build, isBuilder)
	tagAnnotations := tagDefaultAnnotations[imageStreamName]
	if tagAnnotations == nil { //custom image streams won't have a default tag ;)
		tagAnnotations = map[string]string{}
	}
	tagAnnotations[labelKeyVersion] = imageTag
	return imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:        imageStreamName,
			Namespace:   build.Namespace,
			Annotations: imageStreamDefaultAnnotations[imageStreamName],
		},
		Spec: imgv1.ImageStreamSpec{
			Tags: []imgv1.TagReference{
				{
					Name:        imageTag,
					Annotations: tagAnnotations,
					ReferencePolicy: imgv1.TagReferencePolicy{
						Type: imgv1.LocalTagReferencePolicy,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: fmt.Sprintf("%s/%s:%s",
							imageRegistry, resolveKogitoImageName(build, isBuilder), imageTag),
					},
				},
			},
		},
	}
}

// See CreateRequiredKogitoImageStreams
func createRequiredKogitoImageStreamTag(requiredStream imgv1.ImageStream, client *client.Client) (created bool, err error) {
	created = false
	deployedStream := &imgv1.ImageStream{ObjectMeta: metav1.ObjectMeta{Name: requiredStream.Name, Namespace: requiredStream.Namespace}}
	exists, err := kubernetes.ResourceC(client).Fetch(deployedStream)
	tagExists := false
	if err != nil {
		return created, err
	}
	if exists {
		for _, tag := range deployedStream.Spec.Tags {
			if tag.Name == requiredStream.Spec.Tags[0].Name {
				tagExists = true
				break
			}
		}
	}
	// nor tag nor image stream exists, we can safely create a new one for us
	if !tagExists && !exists {
		if err := kubernetes.ResourceC(client).Create(&requiredStream); err != nil {
			return created, err
		}
		created = true
	}
	// the required tag is not there, let's just add the required tag and move on
	if !tagExists && exists {
		deployedStream.Spec.Tags = append(deployedStream.Spec.Tags, requiredStream.Spec.Tags...)
		if err := kubernetes.ResourceC(client).Update(deployedStream); err != nil {
			return created, err
		}
		created = true
	}

	return created, nil
}

// CreateRequiredKogitoImageStreams creates the ImageStreams required by the BuildConfigs to build a custom Kogito Service.
// These images should not be controlled by a given KogitoBuild instance, but reused across all of them.
// This function checks the existence of any of the required ImageStreams by the given instance, if no ImageStream found, creates.
// If the ImageStream exists, but not the tag, a new tag for that same ImageStream is created.
// This way would be possible to handle different builds with different Kogito versions in the same namespace.
// Returns a flag indicating if one of them were created in the cluster or not.
func CreateRequiredKogitoImageStreams(build *v1alpha1.KogitoBuild, client *client.Client) (created bool, err error) {
	buildersCreated := false
	runtimeCreated := false
	if buildersCreated, err = createRequiredKogitoImageStreamTag(newKogitoImageStreamForBuilders(build), client); err != nil {
		return false, err
	}
	if runtimeCreated, err = createRequiredKogitoImageStreamTag(newKogitoImageStreamForRuntime(build), client); err != nil {
		return false, err
	}
	return buildersCreated || runtimeCreated, nil
}
