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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// getSingletonKogitoServiceRoute gets the route from a kogito service that's unique in the given namespace
func getSingletonKogitoServiceRoute(client *client.Client, namespace string, serviceListRef v1alpha1.KogitoServiceList) (string, error) {
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, serviceListRef); err != nil {
		return "", err
	}
	if serviceListRef.GetItemsCount() > 0 {
		return serviceListRef.GetItemAt(0).GetStatus().GetExternalURI(), nil
	}
	return "", nil
}
