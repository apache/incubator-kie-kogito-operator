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

package kogitoservice

import (
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/kiegroup/kogito-operator/version/app"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultKogitoImageFullTag = infrastructure.GetKogitoImageVersion(app.Version) + ":latest"

func Test_createRequiredDeployment_CheckQuarkusProbe(t *testing.T) {
	dataIndex := test.CreateFakeDataIndex(t.Name())
	serviceDef := ServiceDefinition{}
	cli := test.NewFakeClientBuilder().Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	deploymentHandler := NewKogitoDeploymentHandler(context)
	deployment := deploymentHandler.CreateDeployment(dataIndex, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
}

func Test_createRequiredDeployment_CheckDefaultProbe(t *testing.T) {
	dataIndex := test.CreateFakeDataIndex(t.Name())
	serviceDef := ServiceDefinition{}
	cli := test.NewFakeClientBuilder().Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	deploymentHandler := NewKogitoDeploymentHandler(context)
	deployment := deploymentHandler.CreateDeployment(dataIndex, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet)
	assert.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket)
}

func Test_createRequiredDeployment_CheckNilEnvs(t *testing.T) {
	dataIndex := test.CreateFakeDataIndex(t.Name())
	serviceDef := ServiceDefinition{}
	cli := test.NewFakeClientBuilder().Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	deploymentHandler := NewKogitoDeploymentHandler(context)
	deployment := deploymentHandler.CreateDeployment(dataIndex, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, deployment)
	assert.Nil(t, deployment.Spec.Template.Spec.Containers[0].Env)
}

func Test_createRequiredDeployment_CheckSecContext(t *testing.T) {
	dataIndex := test.CreateFakeDataIndex(t.Name())
	managementConsole := test.CreateFakeMgmtConsole(t.Name())
	jobsService := test.CreateFakeJobsService(t.Name())
	explainability := test.CreateFakeExplainabilityService(t.Name())
	taskConsole := test.CreateFakeTaskConsole(t.Name())
	trustUI := test.CreateFakeTrustyUIService(t.Name())
	trustAI := test.CreateFakeTrustyAIService(t.Name())
	serviceDef := ServiceDefinition{}
	cli := test.NewFakeClientBuilder().Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	deploymentHandler := NewKogitoDeploymentHandler(context)
	dataIndexDeployment := deploymentHandler.CreateDeployment(dataIndex, defaultKogitoImageFullTag, serviceDef)
	managementConsoleDeployment := deploymentHandler.CreateDeployment(managementConsole, defaultKogitoImageFullTag, serviceDef)
	jobsServiceDeployment := deploymentHandler.CreateDeployment(jobsService, defaultKogitoImageFullTag, serviceDef)
	explainabilityDeployment := deploymentHandler.CreateDeployment(explainability, defaultKogitoImageFullTag, serviceDef)
	taskConsoleDeployment := deploymentHandler.CreateDeployment(taskConsole, defaultKogitoImageFullTag, serviceDef)
	trustUIDeployment := deploymentHandler.CreateDeployment(trustUI, defaultKogitoImageFullTag, serviceDef)
	trustAIDeployment := deploymentHandler.CreateDeployment(trustAI, defaultKogitoImageFullTag, serviceDef)
	assert.NotNil(t, dataIndexDeployment)
	assert.NotNil(t, managementConsoleDeployment)
	assert.NotNil(t, jobsServiceDeployment)
	assert.NotNil(t, explainabilityDeployment)
	assert.NotNil(t, taskConsoleDeployment)
	assert.NotNil(t, trustUIDeployment)
	assert.NotNil(t, trustAIDeployment)
	specScc := &corev1.PodSecurityContext{RunAsNonRoot: pointer.Bool(true)}
	assert.Equal(t, specScc, dataIndexDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, managementConsoleDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, jobsServiceDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, explainabilityDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, taskConsoleDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, trustUIDeployment.Spec.Template.Spec.SecurityContext)
	assert.Equal(t, specScc, trustAIDeployment.Spec.Template.Spec.SecurityContext)

	containerScc := &corev1.SecurityContext{
		RunAsNonRoot:             pointer.Bool(true),
		AllowPrivilegeEscalation: pointer.Bool(false),
		Privileged:               pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}
	assert.Equal(t, containerScc, dataIndexDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, managementConsoleDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, jobsServiceDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, explainabilityDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, taskConsoleDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, trustUIDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
	assert.Equal(t, containerScc, trustAIDeployment.Spec.Template.Spec.Containers[0].SecurityContext)
}
