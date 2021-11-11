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

package infrastructure

import (
	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createServiceMonitor_defaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.NewFakeClientBuilder().Build()
	deployment := test.CreateFakeDeployment(ns)
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	monitoringManager := prometheusManager{Context: context, instance: deployment}
	serviceMonitor, err := monitoringManager.createServiceMonitor()
	assert.NoError(t, err)
	assert.Equal(t, api.MonitoringDefaultPath, serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, api.MonitoringDefaultScheme, serviceMonitor.Spec.Endpoints[0].Scheme)
}

func Test_createServiceMonitor_customConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.NewFakeClientBuilder().Build()
	deployment := test.CreateFakeDeployment(ns)
	deployment.Annotations = map[string]string{
		MonitoringPathAnnotation:   "/testPath",
		MonitoringSchemeAnnotation: "https",
	}
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	monitoringManager := prometheusManager{Context: context, instance: deployment}
	serviceMonitor, err := monitoringManager.createServiceMonitor()
	assert.NoError(t, err)
	assert.Equal(t, "/testPath", serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, "https", serviceMonitor.Spec.Endpoints[0].Scheme)
}
