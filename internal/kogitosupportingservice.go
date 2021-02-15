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
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoSupportingServiceHandler struct {
	*operator.Context
}

// KogitoSupportingServiceHandler ...
type KogitoSupportingServiceHandler interface {
	FetchKogitoSupportingService(key types.NamespacedName) (api.KogitoSupportingServiceInterface, error)
	FetchKogitoSupportingServiceList(namespace string) (api.KogitoSupportingServiceListInterface, error)
}

// NewKogitoSupportingServiceHandler ...
func NewKogitoSupportingServiceHandler(context *operator.Context) KogitoSupportingServiceHandler {
	return &kogitoSupportingServiceHandler{
		context,
	}
}

// FetchKogitoSupportingService provide kogito supporting service instance
func (k kogitoSupportingServiceHandler) FetchKogitoSupportingService(key types.NamespacedName) (api.KogitoSupportingServiceInterface, error) {
	k.Log.Info("going to fetch deployed kogito supporting service")
	instance := &v1beta1.KogitoSupportingService{}
	if exists, resultErr := kubernetes.ResourceC(k.Client).FetchWithKey(key, instance); resultErr != nil {
		k.Log.Error(resultErr, "Error occurs while fetching deployed kogito supporting service")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.Log.Debug("Successfully fetch deployed kogito supporting reference")
		return instance, nil
	}
}

func (k kogitoSupportingServiceHandler) FetchKogitoSupportingServiceList(namespace string) (api.KogitoSupportingServiceListInterface, error) {
	supportingServiceList := &v1beta1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(k.Client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return nil, err
	}
	return supportingServiceList, nil
}
