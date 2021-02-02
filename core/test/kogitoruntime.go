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

package test

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type fakeKogitoRuntimeHandler struct {
	client *client.Client
}

// CreateFakeKogitoRuntimeHandler ...
func CreateFakeKogitoRuntimeHandler(client *client.Client) api.KogitoRuntimeHandler {
	return &fakeKogitoRuntimeHandler{
		client: client,
	}
}

// FetchKogitoRuntimeService provide KogitoRuntime instance for given name and namespace
func (k *fakeKogitoRuntimeHandler) FetchKogitoRuntimeInstance(key types.NamespacedName) (api.KogitoRuntimeInterface, error) {
	instance := &api2.KogitoRuntimeTest{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		return instance, nil
	}
}

func (k *fakeKogitoRuntimeHandler) FetchAllKogitoRuntimeInstances(namespace string) (api.KogitoRuntimeListInterface, error) {
	kogitoRuntimeServices := &api2.KogitoRuntimeTestList{}
	if err := kubernetes.ResourceC(k.client).ListWithNamespace(namespace, kogitoRuntimeServices); err != nil {
		return nil, err
	}
	if len(kogitoRuntimeServices.Items) == 0 {
		return nil, nil
	}
	return kogitoRuntimeServices, nil
}

// CreateFakeKogitoRuntime ...
func CreateFakeKogitoRuntime(namespace string) *api2.KogitoRuntimeTest {
	replicas := int32(1)
	return &api2.KogitoRuntimeTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-kogito-runtime",
			Namespace: namespace,
		},
		Spec: api2.KogitoRuntimeSpecTest{
			KogitoServiceSpec: api.KogitoServiceSpec{Replicas: &replicas},
		},
	}
}
