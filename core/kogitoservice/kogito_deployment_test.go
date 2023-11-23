// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kogitoservice

import (
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/kiegroup/kogito-operator/version/app"
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
