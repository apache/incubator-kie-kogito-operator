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
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var metaKogitoApp = &v1alpha1.KogitoApp{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test",
	},
	Spec: v1alpha1.KogitoAppSpec{},
}

func Test_addDefaultMeta_whenLabelsAreNotDefined(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	addDefaultMeta(objectMeta, metaKogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyAppName] == "test")
}

func Test_addDefaultMeta_whenAlreadyHasLabels(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"app":      "test123",
			"operator": "kogito",
		},
	}
	addDefaultMeta(objectMeta, metaKogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyAppName] == "test")
	assert.True(t, objectMeta.Labels["operator"] == "kogito")
}

func Test_addDefaultMeta_whenAlreadyHasAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Annotations: map[string]string{
			"test": "test",
		},
	}
	addDefaultMeta(objectMeta, metaKogitoApp)
	assert.True(t, objectMeta.Annotations["test"] == "test")
	assert.True(t, objectMeta.Annotations["org.kie.kogito/managed-by"] == "Kogito Operator")
}

func Test_addServiceLabels_whenLabelsAreNotProvided(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}

	kogitoApp = &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Service: v1alpha1.KogitoAppServiceObject{},
		},
	}

	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyServiceName] == "test")
}

func Test_addServiceLabels_whenAlreadyHasLabels(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"service":  "test123",
			"operator": "kogito",
		},
	}

	kogitoApp = &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Service: v1alpha1.KogitoAppServiceObject{},
		},
	}

	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyServiceName] == "test")
	assert.True(t, objectMeta.Labels["operator"] == "kogito")
}

func Test_addServiceLabels_whenLabelsAreProvided(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"service":  "test123",
			"operator": "kogito123",
		},
	}

	kogitoApp = &v1alpha1.KogitoApp{
		Spec: v1alpha1.KogitoAppSpec{
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: map[string]string{
					"service":  "test456",
					"operator": "kogito456",
				},
			},
		},
	}

	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels["service"] == "test456")
	assert.True(t, objectMeta.Labels["operator"] == "kogito456")
}
