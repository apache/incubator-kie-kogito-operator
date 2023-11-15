/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package app

import (
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestFetchKogitoRuntimeService_InstanceFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoRuntime).Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	runtimeHandler := NewKogitoRuntimeHandler(context)
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoRuntimeService_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	cli := test.NewFakeClientBuilder().Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	runtimeHandler := NewKogitoRuntimeHandler(context)
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.Nil(t, instance)
}
