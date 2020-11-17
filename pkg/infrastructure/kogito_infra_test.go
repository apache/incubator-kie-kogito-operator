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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoInfra).Build()
	instance, err := FetchKogitoInfraInstance(cli, name, ns)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test.NewFakeClientBuilder().Build()
	instance, err := FetchKogitoInfraInstance(cli, name, ns)
	assert.NoError(t, err)
	assert.Nil(t, instance)
}

func TestMustFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test.NewFakeClientBuilder().Build()
	_, err := MustFetchKogitoInfraInstance(cli, name, ns)
	assert.Error(t, err)
}
