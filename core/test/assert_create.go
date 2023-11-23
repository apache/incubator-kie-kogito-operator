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
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

// AssertCreate creates the given resource and asserts that it isn't existed before
func AssertCreate(t *testing.T, c *kogitocli.Client, resource client.Object) {
	err := kubernetes.ResourceC(c).CreateIfNotExists(resource)
	assert.NoError(t, err)
}
