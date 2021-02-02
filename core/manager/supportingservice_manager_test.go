// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package manager

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_ensureSingletonService(t *testing.T) {
	ns := t.Name()
	instance1 := &api2.KogitoSupportingServiceTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index1",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
	}
	instance2 := &api2.KogitoSupportingServiceTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index2",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
	}

	cli := test2.NewFakeClientBuilder().AddK8sObjects(instance1, instance2).OnOpenShift().Build()
	supportingServiceHandler := test2.CreateFakeKogitoSupportingServiceHandler(cli)
	supportingServiceManager := NewKogitoSupportingServiceManager(cli, test2.TestLogger, supportingServiceHandler)
	assert.Errorf(t, supportingServiceManager.EnsureSingletonService(ns, api.DataIndex), "kogito Supporting Service(%s) already exists, please delete the duplicate before proceeding", api.DataIndex)
}
