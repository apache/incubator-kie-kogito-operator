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

package kogitoinfra

import (
	"testing"

	api "github.com/apache/incubator-kie-kogito-operator/apis"
	"github.com/apache/incubator-kie-kogito-operator/core/client/kubernetes"
	"github.com/apache/incubator-kie-kogito-operator/core/operator"
	"github.com/apache/incubator-kie-kogito-operator/core/test"
	"github.com/apache/incubator-kie-kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInfinispanCredentialReconciler(t *testing.T) {
	ns := t.Name()
	kogitoInfinispanInstance := test.CreateFakeKogitoInfinispan(ns)
	infinispanInstance := test.CreateFakeInfinispan(ns)
	infinispanService := test.CreateFakeInfinispanService(ns)
	infinispanSecret := test.CreateFakeInfinispanCredentialSecret(ns)

	cli := test.NewFakeClientBuilder().AddK8sObjects(infinispanInstance, infinispanService, infinispanSecret).Build()
	infraContext := infraContext{
		Context: operator.Context{
			Client: cli,
			Log:    test.TestLogger,
			Scheme: meta.GetRegisteredSchema(),
		},
		instance: kogitoInfinispanInstance,
	}
	infinispanCredentialReconciler := newInfinispanCredentialReconciler(infraContext, infinispanInstance, api.QuarkusRuntimeType)
	err := infinispanCredentialReconciler.Reconcile()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(kogitoInfinispanInstance.GetStatus().GetSecretEnvFromReferences()))
	secretName := kogitoInfinispanInstance.GetStatus().GetSecretEnvFromReferences()[0]
	credentialSecret := &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
	}
	exist, err := kubernetes.ResourceC(cli).Fetch(credentialSecret)
	assert.True(t, exist)
	assert.NoError(t, err)
	assert.NotNil(t, credentialSecret.StringData["QUARKUS_INFINISPAN_CLIENT_USERNAME"])
	assert.NotNil(t, credentialSecret.StringData["QUARKUS_INFINISPAN_CLIENT_PASSWORD"])
}

func TestInfinispanCredentialReconciler_EncryptionDisabled(t *testing.T) {
	ns := t.Name()
	kogitoInfinispanInstance := test.CreateFakeKogitoInfinispan(ns)
	infinispanInstance := test.CreateFakeInfinispan(ns)
	falseValue := false
	infinispanInstance.Spec.Security.EndpointAuthentication = &falseValue

	cli := test.NewFakeClientBuilder().AddK8sObjects(infinispanInstance).Build()
	infraContext := infraContext{
		Context: operator.Context{
			Client: cli,
			Log:    test.TestLogger,
			Scheme: meta.GetRegisteredSchema(),
		},
		instance: kogitoInfinispanInstance,
	}
	infinispanCredentialReconciler := newInfinispanCredentialReconciler(infraContext, infinispanInstance, api.QuarkusRuntimeType)
	err := infinispanCredentialReconciler.Reconcile()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(kogitoInfinispanInstance.GetStatus().GetSecretEnvFromReferences()))

}
