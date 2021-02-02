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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoRuntimeHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewKogitoRuntimeHandler ...
func NewKogitoRuntimeHandler(client *client.Client, log logger.Logger) api.KogitoRuntimeHandler {
	return &kogitoRuntimeHandler{
		client: client,
		log:    log,
	}
}

// FetchKogitoRuntimeService provide KogitoRuntime instance for given name and namespace
func (k *kogitoRuntimeHandler) FetchKogitoRuntimeInstance(key types.NamespacedName) (api.KogitoRuntimeInterface, error) {
	k.log.Debug("going to fetch deployed kogito runtime service")
	instance := &v1beta1.KogitoRuntime{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		k.log.Error(resultErr, "Error occurs while fetching deployed kogito runtime service instance")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.log.Debug("Successfully fetch deployed kogito runtime reference")
		return instance, nil
	}
}

func (k *kogitoRuntimeHandler) FetchAllKogitoRuntimeInstances(namespace string) (api.KogitoRuntimeListInterface, error) {
	kogitoRuntimeServices := &v1beta1.KogitoRuntimeList{}
	if err := kubernetes.ResourceC(k.client).ListWithNamespace(namespace, kogitoRuntimeServices); err != nil {
		return nil, err
	}
	if len(kogitoRuntimeServices.Items) == 0 {
		k.log.Debug("No instance found for KogitoRuntime service")
		return nil, nil
	}
	k.log.Debug("Found KogitoRuntime services", "count", len(kogitoRuntimeServices.Items))
	return kogitoRuntimeServices, nil
}
