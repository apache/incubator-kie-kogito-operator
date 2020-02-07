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
	"github.com/kiegroup/kogito-cloud-operator/version"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	imgv1 "github.com/openshift/api/image/v1"
)

func TestKogitoImageStreamGeneration(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{Namespace: "test"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: "",
			Build:   &v1alpha1.KogitoAppBuildObject{Native: false},
		},
	}
	itemsTest := CreateKogitoImageStream(kogitoApp, version.Version)
	assert.Equal(t, 5, len(itemsTest.Items))

	for _, item := range itemsTest.Items {
		assert.Equal(t, "test", item.Namespace)
		assert.Equal(t, "Kie Group.", item.Annotations["openshift.io/provider-display-name"])
		assert.Equal(t, "icon-jbpm", item.Spec.Tags[0].Annotations["iconClass"])
		assert.Equal(t, version.Version, item.Spec.Tags[0].Annotations["version"])

		assert.Equal(t, version.Version, item.Spec.Tags[0].Name)

		switch item.Name {
		case KogitoQuarkusUbi8Image:
			assert.Equal(t, "Runtime image for Kogito based on Quarkus native image", item.Annotations["openshift.io/display-name"])
			assert.Equal(t, "Runtime image for Kogito based on Quarkus native image", item.Spec.Tags[0].Annotations["description"])
			assert.Equal(t, "runtime,kogito,quarkus", item.Spec.Tags[0].Annotations["tags"])
			assert.Equal(t, "quarkus", item.Spec.Tags[0].Annotations["supports"])
			assert.Equal(t, imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy}, item.Spec.Tags[0].ReferencePolicy)
			assert.Equal(t, "DockerImage", item.Spec.Tags[0].From.Kind)
			assert.Equal(t, fmt.Sprintf("quay.io/kiegroup/%s:%s", item.Name, version.Version), item.Spec.Tags[0].From.Name)

		case KogitoQuarkusJVMUbi8Image:
			assert.Equal(t, "Runtime image for Kogito based on Quarkus JVM image", item.Annotations["openshift.io/display-name"])
			assert.Equal(t, "Runtime image for Kogito based on Quarkus JVM image", item.Spec.Tags[0].Annotations["description"])
			assert.Equal(t, "runtime,kogito,quarkus,jvm", item.Spec.Tags[0].Annotations["tags"])
			assert.Equal(t, "quarkus", item.Spec.Tags[0].Annotations["supports"])
			assert.Equal(t, imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy}, item.Spec.Tags[0].ReferencePolicy)
			assert.Equal(t, "DockerImage", item.Spec.Tags[0].From.Kind)
			assert.Equal(t, fmt.Sprintf("quay.io/kiegroup/%s:%s", item.Name, version.Version), item.Spec.Tags[0].From.Name)

		case KogitoQuarkusUbi8s2iImage:
			assert.Equal(t, "Platform for building Kogito based on Quarkus", item.Annotations["openshift.io/display-name"])
			assert.Equal(t, "Platform for building Kogito based on Quarkus", item.Spec.Tags[0].Annotations["description"])
			assert.Equal(t, "builder,kogito,quarkus", item.Spec.Tags[0].Annotations["tags"])
			assert.Equal(t, "quarkus", item.Spec.Tags[0].Annotations["supports"])
			assert.Equal(t, imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy}, item.Spec.Tags[0].ReferencePolicy)
			assert.Equal(t, "DockerImage", item.Spec.Tags[0].From.Kind)
			assert.Equal(t, fmt.Sprintf("quay.io/kiegroup/%s:%s", item.Name, version.Version), item.Spec.Tags[0].From.Name)

		case KogitoSpringbootUbi8Image:
			assert.Equal(t, "Runtime image for Kogito based on SpringBoot", item.Annotations["openshift.io/display-name"])
			assert.Equal(t, "Runtime image for Kogito based on SpringBoot", item.Spec.Tags[0].Annotations["description"])
			assert.Equal(t, "runtime,kogito,springboot", item.Spec.Tags[0].Annotations["tags"])
			assert.Equal(t, "springboot", item.Spec.Tags[0].Annotations["supports"])
			assert.Equal(t, imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy}, item.Spec.Tags[0].ReferencePolicy)
			assert.Equal(t, "DockerImage", item.Spec.Tags[0].From.Kind)
			assert.Equal(t, fmt.Sprintf("quay.io/kiegroup/%s:%s", item.Name, version.Version), item.Spec.Tags[0].From.Name)

		case KogitoSpringbootUbi8s2iImage:
			assert.Equal(t, "Platform for building Kogito based on SpringBoot", item.Annotations["openshift.io/display-name"])
			assert.Equal(t, "Platform for building Kogito based on SpringBoot", item.Spec.Tags[0].Annotations["description"])
			assert.Equal(t, "builder,kogito,springboot", item.Spec.Tags[0].Annotations["tags"])
			assert.Equal(t, "springboot", item.Spec.Tags[0].Annotations["supports"])
			assert.Equal(t, imgv1.TagReferencePolicy{Type: imgv1.LocalTagReferencePolicy}, item.Spec.Tags[0].ReferencePolicy)
			assert.Equal(t, "DockerImage", item.Spec.Tags[0].From.Kind)
			assert.Equal(t, fmt.Sprintf("quay.io/kiegroup/%s:%s", item.Name, version.Version), item.Spec.Tags[0].From.Name)
		}
	}

}

func TestQuarkusKogitoImageStreamGenerationNonNative(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{Namespace: "quarkus"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build:   &v1alpha1.KogitoAppBuildObject{Native: false},
		},
	}
	itemsTest := CreateKogitoImageStream(kogitoApp, version.Version)
	assert.Equal(t, 2, len(itemsTest.Items))
	assert.True(t, containsIsName(KogitoQuarkusJVMUbi8Image, itemsTest.Items))
	assert.True(t, containsIsName(KogitoQuarkusUbi8s2iImage, itemsTest.Items))
}

func TestQuarkusKogitoImageStreamGenerationNative(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{Namespace: "quarkus"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.QuarkusRuntimeType,
			Build:   &v1alpha1.KogitoAppBuildObject{Native: true},
		},
	}
	itemsTest := CreateKogitoImageStream(kogitoApp, version.Version)
	assert.Equal(t, 2, len(itemsTest.Items))
	assert.True(t, containsIsName(KogitoQuarkusUbi8s2iImage, itemsTest.Items))
	assert.True(t, containsIsName(KogitoQuarkusUbi8Image, itemsTest.Items))
}

func TestSpringbootKogitoImageStreamGenerationNative(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{Namespace: "springboot"},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: v1alpha1.SpringbootRuntimeType,
			Build:   &v1alpha1.KogitoAppBuildObject{Native: false},
		},
	}
	itemsTest := CreateKogitoImageStream(kogitoApp, version.Version)
	assert.Equal(t, 2, len(itemsTest.Items))
	assert.True(t, containsIsName(KogitoSpringbootUbi8Image, itemsTest.Items))
	assert.True(t, containsIsName(KogitoSpringbootUbi8s2iImage, itemsTest.Items))
}

// containsIsName checks if the is name is present on the imageStream list
func containsIsName(s string, array []imgv1.ImageStream) bool {
	for _, item := range array {
		if s == item.Name {
			return true
		}
	}
	return false
}
