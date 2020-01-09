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

	"github.com/stretchr/testify/assert"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	dockerv10 "github.com/openshift/api/image/docker10"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_deploymentConfigResource_NewWithValidDocker(t *testing.T) {
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
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http,8181;https",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusPath:         "/metrics",
				framework.LabelPrometheusPort:         "8080",
				framework.LabelPrometheusScheme:       "http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage)
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	// we should have only one port. the 8181 is invalid.
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, int32(8080), dc.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	// this one where added by the docker image :)
	assert.Equal(t, "kogito", dc.Labels["operator"])
	// prometheus labels
	assert.Equal(t, "/metrics", dc.Spec.Template.Annotations[framework.LabelPrometheusPath])
	assert.Equal(t, "8080", dc.Spec.Template.Annotations[framework.LabelPrometheusPort])
	assert.Equal(t, "http", dc.Spec.Template.Annotations[framework.LabelPrometheusScheme])
	assert.Equal(t, "true", dc.Spec.Template.Annotations[framework.LabelPrometheusScrape])
}

func Test_deploymentConfigResource_NewWithInvalidDocker(t *testing.T) {
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
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, &dockerv10.DockerImage{})
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	assert.Len(t, dc.Spec.Selector, 1)
	assert.Len(t, dc.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, bcRuntime.Spec.Output.To.Name, dc.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "test", dc.Labels[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Selector[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Template.Labels[LabelKeyAppName])
}
