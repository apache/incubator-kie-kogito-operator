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

// NewKogitoInfraHandler ...
func NewKogitoInfraHandler(context *operator.Context) manager.KogitoInfraHandler {
	return &kogitoInfraHandler{context}
}

// NewNoOpKogitoInfraHandler returns a new instance of manager.KogitoInfraHandler which won't try to interact with the api.KogitoInfraInterface object
func NewNoOpKogitoInfraHandler(context *operator.Context) manager.KogitoInfraHandler {
	return &kogitoInfraHandlerNoOp{context}
}

type kogitoInfraHandlerNoOp struct {
	*operator.Context
}

func (k *kogitoInfraHandlerNoOp) FetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	k.Log.Debug("NoOp InfraHandler: Returning a blank KogitoInfra. Infra features are not enabled for this instance.", "KogitoInfra", key)
	return &v1beta1.KogitoInfra{}, nil
}

type kogitoInfraHandler struct {
	*operator.Context
}

// FetchKogitoInfraInstance loads a given infra instance by name and namespace.
// If the KogitoInfra resource is not present, nil will return.
func (k *kogitoInfraHandler) FetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	k.Log.Debug("going to fetch deployed kogito infra instance")
	instance := &v1beta1.KogitoInfra{}
	if exists, resultErr := kubernetes.ResourceC(k.Client).FetchWithKey(key, instance); resultErr != nil {
		k.Log.Error(resultErr, "Error occurs while fetching deployed kogito infra instance")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.Log.Debug("Successfully fetch deployed kogito infra reference")
		return instance, nil
	}
}
