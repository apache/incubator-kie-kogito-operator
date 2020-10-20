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

// getKogitoSupportingServiceRoute gets the route from a kogito service that's unique in the given namespace
func getKogitoSupportingServiceRoute(client *client.Client, namespace string, resourceType v1alpha1.ServiceType) (route string, err error) {
	supportingService, err := getKogitoSupportingService(client, namespace, resourceType)
	if err != nil {
		return
	}
	if supportingService != nil {
		return supportingService.GetStatus().GetExternalURI(), nil
	}
	return
}

func getKogitoSupportingService(client *client.Client, namespace string, resourceType v1alpha1.ServiceType) (*v1alpha1.KogitoSupportingService, error) {
	supportingServiceList := &v1alpha1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return nil, err
	}
	for _, service := range supportingServiceList.Items {
		if service.Spec.ServiceType == resourceType {
			return &service, nil
		}
	}
	return nil, nil
}
