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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createServiceMonitor_defaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.NewFakeClientBuilder().Build()
	kogitoService := test.CreateFakeKogitoRuntime(ns)
	monitoringManager := prometheusManager{client: cli, log: logger.GetLogger("monitoring"), scheme: meta.GetRegisteredSchema()}
	serviceMonitor, err := monitoringManager.createServiceMonitor(kogitoService)
	assert.NoError(t, err)
	assert.Equal(t, api.MonitoringDefaultPath, serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, api.MonitoringDefaultScheme, serviceMonitor.Spec.Endpoints[0].Scheme)
}

func Test_createServiceMonitor_customConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.NewFakeClientBuilder().Build()
	kogitoService := test.CreateFakeKogitoRuntime(ns)
	kogitoService.GetSpec().SetMonitoring(api.Monitoring{
		Path:   "/testPath",
		Scheme: "https",
	})
	monitoringManager := prometheusManager{client: cli, log: logger.GetLogger("monitoring"), scheme: meta.GetRegisteredSchema()}
	serviceMonitor, err := monitoringManager.createServiceMonitor(kogitoService)
	assert.NoError(t, err)
	assert.Equal(t, "/testPath", serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, "https", serviceMonitor.Spec.Endpoints[0].Scheme)
}
