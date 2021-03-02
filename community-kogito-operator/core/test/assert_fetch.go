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

package test

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/core/client/kubernetes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// AssertFetchMustExist fetches the given object and verify if exists in the context without errors
func AssertFetchMustExist(t *testing.T, client *client.Client, resource kubernetes.ResourceObject) {
	exists, err := kubernetes.ResourceC(client).Fetch(resource)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// AssertFetchWithKeyMustExist fetches the given object with the defined key and verify if it exists in the context without errors
func AssertFetchWithKeyMustExist(t *testing.T, client *client.Client, resource kubernetes.ResourceObject, instance metav1.Object) {
	exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}, resource)
	assert.NoError(t, err)
	assert.True(t, exists)
}
