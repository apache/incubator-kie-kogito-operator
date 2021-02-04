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
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	customKogitoImagePrefix = "custom-"
	labelKeyVersion         = "version"
)

var (
	// imageStreamDefaultAnnotations lists the default annotations for ImageStreams
	imageStreamDefaultAnnotations = map[string]map[string]string{
		infrastructure.KogitoRuntimeNative: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Runtime image for Kogito based on Quarkus native image",
		},
		infrastructure.KogitoRuntimeJVM: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Runtime image for Kogito based on Quarkus or Spring Boot JVM image",
		},
		infrastructure.KogitoBuilderImage: {
			"openshift.io/provider-display-name": "KIE Group",
			"openshift.io/display-name":          "Platform for building Kogito based on Quarkus or Spring Boot",
		},
	}

	//tagDefaultAnnotations lists the default annotations for ImageStreamTags
	tagDefaultAnnotations = map[string]map[string]string{
		infrastructure.KogitoRuntimeNative: {
			"iconClass":   "icon-jbpm",
			"description": "Runtime image for Kogito based on Quarkus native image",
			"tags":        "runtime,kogito,quarkus",
			"supports":    "quarkus",
		},
		infrastructure.KogitoRuntimeJVM: {
			"iconClass":   "icon-jbpm",
			"description": "Runtime image for Kogito based on Quarkus or Spring Boot JVM image",
			"tags":        "runtime,kogito,quarkus,springboot,jvm",
			"supports":    "quarkus,springboot",
		},
		infrastructure.KogitoBuilderImage: {
			"iconClass":   "icon-jbpm",
			"description": "Platform for building Kogito based on Quarkus or Spring Boot",
			"tags":        "builder,kogito,quarkus,springboot",
			"supports":    "quarkus,springboot",
		},
	}
)

// resolveKogitoImageNameTag resolves the ImageStreamTag to be used in the given build, e.g. kogito-quarkus-ubi8-s2i:0.11
func resolveKogitoImageNameTag(build *v1beta1.KogitoBuild, isBuilder bool) string {
	return strings.Join([]string{
		resolveKogitoImageName(build, isBuilder),
		resolveKogitoImageTag(build, isBuilder),
	}, ":")
}

// resolveKogitoImageTag resolves the ImageTag to be used in the given build, e.g. 0.11
func resolveKogitoImageTag(build *v1beta1.KogitoBuild, isBuilder bool) string {
	image := framework.ConvertImageTagToImage(build.Spec.RuntimeImage)
	if isBuilder {
		image = framework.ConvertImageTagToImage(build.Spec.BuildImage)
	}
	if len(image.Tag) > 0 {
		return image.Tag
	}
	return infrastructure.GetKogitoImageVersion()
}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageName(build *v1beta1.KogitoBuild, isBuilder bool) string {
	if isBuilder {
		image := framework.ConvertImageTagToImage(build.Spec.BuildImage)
		if len(image.Name) > 0 {
			return image.Name
		}
		return infrastructure.KogitoBuilderImage
	}
	image := framework.ConvertImageTagToImage(build.Spec.RuntimeImage)
	if len(image.Name) > 0 {
		return image.Name
	}
	if build.Spec.Native {
		return infrastructure.KogitoRuntimeNative
	}
	return infrastructure.KogitoRuntimeJVM

}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageStreamName(build *v1beta1.KogitoBuild, isBuilder bool) string {
	imageName := resolveKogitoImageName(build, isBuilder)
	image := framework.ConvertImageTagToImage(build.Spec.RuntimeImage)
	if isBuilder {
		image = framework.ConvertImageTagToImage(build.Spec.BuildImage)
	}
	if len(image.Name) > 0 { // custom image
		return strings.Join([]string{customKogitoImagePrefix, imageName}, "")
	}
	return imageName
}

// resolveKogitoImageName resolves the ImageName to be used in the given build, e.g. kogito-quarkus-ubi8-s2i
func resolveKogitoImageStreamTagName(build *v1beta1.KogitoBuild, isBuilder bool) string {
	imageStream := resolveKogitoImageStreamName(build, isBuilder)
	imageTag := resolveKogitoImageTag(build, isBuilder)
	return strings.Join([]string{imageStream, imageTag}, ":")
}

// resolveImageRegistry resolves the registry/namespace name to be used in the given build, e.g. quay.io/kiegroup
func resolveKogitoImageRegistryNamespace(build *v1beta1.KogitoBuild, isBuilder bool) string {
	namespace := infrastructure.DefaultImageNamespace
	registry := infrastructure.DefaultImageRegistry
	image := framework.ConvertImageTagToImage(build.Spec.RuntimeImage)
	if isBuilder {
		image = framework.ConvertImageTagToImage(build.Spec.BuildImage)
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
func newKogitoImageStreamForBuilders(build *v1beta1.KogitoBuild) imgv1.ImageStream {
	return newKogitoImageStream(build, true)
}

// newKogitoImageStreamForRuntime same as newKogitoImageStream(build, false)
func newKogitoImageStreamForRuntime(build *v1beta1.KogitoBuild) imgv1.ImageStream {
	return newKogitoImageStream(build, false)
}

// newKogitoImageStream creates a new OpenShift ImageStream based on the given build and the image purpose
func newKogitoImageStream(build *v1beta1.KogitoBuild, isBuilder bool) imgv1.ImageStream {
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
			// double check since the object could've been created in another thread
			if errors.IsAlreadyExists(err) {
				exists = true
				created = false
			} else {
				return created, err
			}
		} else {
			created = true
		}
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
func CreateRequiredKogitoImageStreams(build *v1beta1.KogitoBuild, client *client.Client) (created bool, err error) {
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
