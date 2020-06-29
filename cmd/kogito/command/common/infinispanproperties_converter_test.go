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

package common

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_FromInfinispanFlagsToInfinispanProperties_EnablePersistenceWithUserDefineProperties(t *testing.T) {
	ns := t.Name()
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	infinispanFlags := &InfinispanFlags{
		URI:                "test-uri",
		InfinispanUser:     "user",
		InfinispanPassword: "password",
	}
	infinispanProperties, err := FromInfinispanFlagsToInfinispanProperties(client, ns, infinispanFlags, true)
	assert.Nil(t, err)
	assert.False(t, infinispanProperties.UseKogitoInfra)
	assert.Equal(t, "test-uri", infinispanProperties.URI)
}

func Test_FromInfinispanFlagsToInfinispanProperties_EnablePersistenceWithDefaultProperties(t *testing.T) {
	ns := t.Name()
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	infinispanFlags := &InfinispanFlags{}
	infinispanProperties, err := FromInfinispanFlagsToInfinispanProperties(client, ns, infinispanFlags, true)
	assert.Nil(t, err)
	assert.True(t, infinispanProperties.UseKogitoInfra)
}

func Test_FromInfinispanFlagsToInfinispanProperties_DisablePersistence(t *testing.T) {
	ns := t.Name()
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	infinispanFlags := &InfinispanFlags{}
	infinispanProperties, err := FromInfinispanFlagsToInfinispanProperties(client, ns, infinispanFlags, false)
	assert.Nil(t, err)
	assert.False(t, infinispanProperties.UseKogitoInfra)
}
