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

package test

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// AssertFetchMustExist fetches the given object and verify if exists in the context without errors
func AssertFetchMustExist(t *testing.T, client *kogitocli.Client, resource client.Object) {
	exists, err := kubernetes.ResourceC(client).Fetch(resource)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// AssertFetchWithKeyMustExist fetches the given object with the defined key and verify if it exists in the context without errors
func AssertFetchWithKeyMustExist(t *testing.T, client *kogitocli.Client, resource client.Object, instance metav1.Object) {
	exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, resource)
	assert.NoError(t, err)
	assert.True(t, exists)
}
