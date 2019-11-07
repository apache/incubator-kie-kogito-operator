// Copyright 2019 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// getKogitoDataIndexRoute gets the deployed data index route
func getKogitoDataIndexRoute(client *client.Client, namespace string) (string, error) {
	route := ""
	dataIndexes := &v1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, dataIndexes); err != nil {
		return route, err
	}
	if len(dataIndexes.Items) > 0 {
		// should be only one data index guaranteed by OLM, but still we are looking for the first one
		route = dataIndexes.Items[0].Status.Route
	}
	return route, nil
}
