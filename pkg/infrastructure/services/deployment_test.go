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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultKogitoImageFullTag = infrastructure.GetKogitoImageVersion() + ":latest"

func Test_createRequiredDeployment_CheckQuarkusProbe(t *testing.T) {
	kogitoService := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultDataIndexImageName, Namespace: t.Name()},
	}
	serviceDef := ServiceDefinition{HealthCheckProbe: QuarkusHealthCheckProbe}
	deployment := createRequiredDeployment(kogitoService, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
}

func Test_createRequiredDeployment_CheckDefaultProbe(t *testing.T) {
	kogitoService := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultDataIndexImageName, Namespace: t.Name()},
	}
	serviceDef := ServiceDefinition{}
	deployment := createRequiredDeployment(kogitoService, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
}

func Test_createRequiredDeployment_CheckNilEnvs(t *testing.T) {
	kogitoService := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultJobsServiceImageName, Namespace: t.Name()},
	}
	serviceDef := ServiceDefinition{}
	deployment := createRequiredDeployment(kogitoService, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].Env)

	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoService, deployment).Build()
	exists, err := kubernetes.ResourceC(cli).Fetch(deployment)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].Env)
}
