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

package kogitoservice

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"os"
)

const (
	envVarKogitoServiceURL = "LOCAL_KOGITO_SERVICE_URL"
)

// ServiceHandler ...
type ServiceHandler interface {
	GetKogitoServiceEndpoint(kogitoService api.KogitoService) string
}

type kogitoServiceHandler struct {
	log logger.Logger
}

// NewKogitoServiceHandler ...
func NewKogitoServiceHandler(log logger.Logger) ServiceHandler {
	return &kogitoServiceHandler{
		log: log,
	}
}

// GetKogitoServiceEndpoint gets the endpoint depending on
// if the envVarKogitoServiceURL is set (for when running
// operator locally). Else, the internal endpoint is
// returned.
func (k *kogitoServiceHandler) GetKogitoServiceEndpoint(kogitoService api.KogitoService) string {
	externalURL := os.Getenv(envVarKogitoServiceURL)
	if len(externalURL) > 0 {
		return externalURL
	}
	return k.getKogitoServiceURL(kogitoService)
}

// getKogitoServiceURL provides kogito service URL for given instance name
func (k *kogitoServiceHandler) getKogitoServiceURL(service api.KogitoService) string {
	k.log.Debug("Creating kogito service instance URL.")
	// resolves to http://servicename.mynamespace for example
	serviceURL := fmt.Sprintf("http://%s.%s", service.GetName(), service.GetNamespace())
	k.log.Debug("", "kogito service instance URL", serviceURL)
	return serviceURL
}
