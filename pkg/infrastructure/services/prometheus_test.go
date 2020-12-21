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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_createServiceMonitor_defaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	kogitoService := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "travels",
			Namespace: ns,
		},
	}

	serviceMonitor, err := createServiceMonitor(cli, kogitoService, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.Equal(t, v1beta1.MonitoringDefaultPath, serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, v1beta1.MonitoringDefaultScheme, serviceMonitor.Spec.Endpoints[0].Scheme)
}

func Test_createServiceMonitor_customConfiguration(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	kogitoService := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "travels",
			Namespace: ns,
		},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Monitoring: v1beta1.Monitoring{
					Path:   "/testPath",
					Scheme: "https",
				},
			},
		},
	}

	serviceMonitor, err := createServiceMonitor(cli, kogitoService, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.Equal(t, "/testPath", serviceMonitor.Spec.Endpoints[0].Path)
	assert.Equal(t, "https", serviceMonitor.Spec.Endpoints[0].Scheme)
}
