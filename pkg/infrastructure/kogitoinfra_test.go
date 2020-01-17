// Copyright 2019 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_CreateOrFetchKogitoInfra_NotExists(t *testing.T) {
	ns := t.Name()
	cli := test.CreateFakeClient(nil, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithKafka().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
}

func Test_CreateOrFetchKogitoInfra_Exists(t *testing.T) {
	ns := t.Name()
	infra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: DefaultKogitoInfraName, Namespace: ns},
		Spec:       v1alpha1.KogitoInfraSpec{InstallInfinispan: false},
	}
	cli := test.CreateFakeClient([]runtime.Object{infra}, nil, nil)
	infra, ready, err := EnsureKogitoInfra(ns, cli).WithInfinispan().Apply()
	assert.NoError(t, err)
	assert.NotNil(t, infra)
	// should not be ready since we just deployed Infinispan
	assert.False(t, ready)
	assert.Equal(t, DefaultKogitoInfraName, infra.GetName())
	assert.True(t, infra.Spec.InstallInfinispan)
}
