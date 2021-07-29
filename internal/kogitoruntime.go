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
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoRuntimeHandler struct {
	operator.Context
}

// NewKogitoRuntimeHandler ...
func NewKogitoRuntimeHandler(context operator.Context) manager.KogitoRuntimeHandler {
	return &kogitoRuntimeHandler{
		context,
	}
}

// FetchKogitoRuntimeService provide KogitoRuntime instance for given name and namespace
func (k *kogitoRuntimeHandler) FetchKogitoRuntimeInstance(key types.NamespacedName) (api.KogitoRuntimeInterface, error) {
	k.Log.Debug("going to fetch deployed kogito runtime service")
	instance := &v1beta1.KogitoRuntime{}
	if exists, resultErr := kubernetes.ResourceC(k.Client).FetchWithKey(key, instance); resultErr != nil {
		k.Log.Error(resultErr, "Error occurs while fetching deployed kogito runtime service instance")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.Log.Debug("Successfully fetch deployed kogito runtime reference")
		return instance, nil
	}
}

func (k *kogitoRuntimeHandler) FetchAllKogitoRuntimeInstances(namespace string) (api.KogitoRuntimeListInterface, error) {
	kogitoRuntimeServices := &v1beta1.KogitoRuntimeList{}
	if err := kubernetes.ResourceC(k.Client).ListWithNamespace(namespace, kogitoRuntimeServices); err != nil {
		return nil, err
	}
	k.Log.Debug("Found KogitoRuntime services", "count", len(kogitoRuntimeServices.Items))
	return kogitoRuntimeServices, nil
}
