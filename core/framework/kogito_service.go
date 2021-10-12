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

package framework

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
	"os"
)

const (
	// EnvVarKogitoServiceURL ...
	EnvVarKogitoServiceURL = "LOCAL_KOGITO_SERVICE_URL"
)

// ServiceHandler ...
type ServiceHandler interface {
	GetKogitoServiceEndpoint(key types.NamespacedName) string
}

type kogitoServiceHandler struct {
	operator.Context
}

// NewKogitoServiceHandler ...
func NewKogitoServiceHandler(context operator.Context) ServiceHandler {
	return &kogitoServiceHandler{
		context,
	}
}

// GetKogitoServiceEndpoint gets the endpoint depending on
// if the EnvVarKogitoServiceURL is set (for when running
// operator locally). Else, the internal endpoint is
// returned.
func (k *kogitoServiceHandler) GetKogitoServiceEndpoint(key types.NamespacedName) string {
	externalURL := os.Getenv(EnvVarKogitoServiceURL)
	if len(externalURL) > 0 {
		return externalURL
	}
	return k.getKogitoServiceURL(key)
}

// getKogitoServiceURL provides kogito service URL for given instance name
func (k *kogitoServiceHandler) getKogitoServiceURL(key types.NamespacedName) string {
	k.Log.Debug("Creating kogito service instance URL.")
	// resolves to http://servicename.mynamespace for example
	serviceURL := fmt.Sprintf("http://%s.%s", key.Name, key.Namespace)
	k.Log.Debug("", "kogito service instance URL", serviceURL)
	return serviceURL
}
