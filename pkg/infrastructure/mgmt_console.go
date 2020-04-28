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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
)

const (
	// DefaultMgmtConsoleName ...
	DefaultMgmtConsoleName = "management-console"
	// DefaultMgmtConsoleImageName ...
	DefaultMgmtConsoleImageName = "kogito-management-console"
)

// GetManagementConsoleEndpoint gets the route for the Management Console deployed in the given namespace
func GetManagementConsoleEndpoint(client *client.Client, namespace string) (ServiceEndpoints, error) {
	endpoints := ServiceEndpoints{}
	route, err := getSingletonKogitoServiceRoute(client, namespace, &v1alpha1.KogitoMgmtConsoleList{})
	if err != nil {
		return endpoints, err
	}
	endpoints.HTTPRouteURI = route
	return endpoints, nil
}
