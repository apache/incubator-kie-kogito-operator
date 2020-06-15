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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultDataIndexImageFullTag = infrastructure.GetRuntimeImageVersion() + ":latest"

func Test_createRequiredDeployment_CheckQuarkusProbe(t *testing.T) {
	kogitoService := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultDataIndexImageName, Namespace: t.Name()},
	}
	serviceDef := ServiceDefinition{HealthCheckProbe: QuarkusHealthCheckProbe}
	deployment := createRequiredDeployment(kogitoService, defaultDataIndexImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
}

func Test_createRequiredDeployment_CheckDefaultProbe(t *testing.T) {
	kogitoService := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultDataIndexImageName, Namespace: t.Name()},
	}
	serviceDef := ServiceDefinition{}
	deployment := createRequiredDeployment(kogitoService, defaultDataIndexImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
}

func Test_createRequiredDeployment_CheckCustomPort(t *testing.T) {
	kogitoService := &v1alpha1.KogitoDataIndex{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultDataIndexImageName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				HTTPPort: 9090,
			},
		},
	}
	serviceDef := ServiceDefinition{}
	deployment := createRequiredDeployment(kogitoService, defaultDataIndexImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port.IntVal)
	assert.Equal(t, int32(9090), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port.IntVal)
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, int32(9090), container.Ports[0].ContainerPort)
	index := framework.GetEnvVar(HTTPPortEnvKey, container.Env)
	assert.True(t, index >= 0, HTTPPortEnvKey, " not found in container env var")
	assert.Equal(t, "9090", container.Env[index].Value)
}
