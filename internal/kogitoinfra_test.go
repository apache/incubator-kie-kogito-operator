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

package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestFetchKogitoInfraInstance_InstanceFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	cli := test.NewFakeClientBuilder().UseScheme(GetRegisteredSchema()).AddK8sObjects(kogitoInfra).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	infraHandler := NewKogitoInfraHandler(context)
	instance, err := infraHandler.FetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test.NewFakeClientBuilder().UseScheme(GetRegisteredSchema()).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: test.GetRegisteredSchema(),
	}
	infraHandler := NewKogitoInfraHandler(context)
	instance, err := infraHandler.FetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.Nil(t, instance)
}
