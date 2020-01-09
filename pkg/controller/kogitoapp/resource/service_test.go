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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	dockerv10 "github.com/openshift/api/image/docker10"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_serviceResource_NewWithAndWithoutDockerImg(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
			},
		},
	}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, _ := newDeploymentConfig(kogitoApp, &bcRuntime, nil)
	svc := newService(kogitoApp, dc)
	assert.Nil(t, svc)
	// try again, now with ports
	dc, _ = newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage)
	svc = newService(kogitoApp, dc)
	assert.NotNil(t, svc)
	assert.Len(t, svc.Spec.Ports, 1)
	assert.Equal(t, int32(8080), svc.Spec.Ports[0].Port)
	assert.Contains(t, svc.Annotations, framework.LabelPrometheusScrape)
}

func Test_addServiceLabels_whenLabelsAreNotProvided(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}

	kogitoApp := &v1alpha1.KogitoApp{
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

	kogitoApp := &v1alpha1.KogitoApp{
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

	kogitoApp := &v1alpha1.KogitoApp{
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
