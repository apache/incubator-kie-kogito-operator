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

type kogitoInfraHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewKogitoInfraHandler ...
func NewKogitoInfraHandler(client *client.Client, log logger.Logger) api.KogitoInfraHandler {
	return &kogitoInfraHandler{
		client: client,
		log:    log,
	}
}

// FetchKogitoInfraInstance loads a given infra instance by name and namespace.
// If the KogitoInfra resource is not present, nil will return.
func (k *kogitoInfraHandler) FetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	k.log.Debug("going to fetch deployed kogito infra instance")
	instance := &v1beta1.KogitoInfra{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		k.log.Error(resultErr, "Error occurs while fetching deployed kogito infra instance")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.log.Debug("Successfully fetch deployed kogito infra reference")
		return instance, nil
	}
}
