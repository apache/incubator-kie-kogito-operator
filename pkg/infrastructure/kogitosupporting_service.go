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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

// getKogitoSupportingServiceRoute gets the route from a kogito service that's unique in the given namespace
func getKogitoSupportingServiceRoute(client *client.Client, namespace string, resourceType v1beta1.ServiceType) (route string, err error) {
	supportingService, err := getKogitoSupportingService(client, namespace, resourceType)
	if err != nil {
		return
	}
	if supportingService != nil {
		return supportingService.GetStatus().GetExternalURI(), nil
	}
	return
}

func getKogitoSupportingService(client *client.Client, namespace string, resourceType v1beta1.ServiceType) (*v1beta1.KogitoSupportingService, error) {
	supportingServiceList := &v1beta1.KogitoSupportingServiceList{}
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

// getSupportingServiceDeployment gets deployment owned by supporting service within the given namespace
func getSupportingServiceDeployment(namespace string, cli *client.Client, serviceType v1beta1.ServiceType) (*appsv1.Deployment, error) {
	supportingService, err := getKogitoSupportingService(cli, namespace, serviceType)
	if err != nil {
		return nil, err
	} else if supportingService == nil {
		log.Debug("KogitoSupportingServce objects not found", "service type", serviceType, "namespace", namespace)
		return nil, nil
	}
	log.Debug("Found", "services", serviceType, "namespace", namespace)

	dcs := &appsv1.DeploymentList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	log.Debug("Looking for owned Deployments", "service type", serviceType)
	for _, dc := range dcs.Items {
		for _, owner := range dc.OwnerReferences {
			if owner.UID == supportingService.UID {
				return &dc, nil
			}
		}
	}
	return nil, nil
}

// FetchKogitoSupportingService provide kogito supporting service instance
func FetchKogitoSupportingService(cli *client.Client, name string, namespace string) (*v1beta1.KogitoSupportingService, error) {
	log.Warn("going to fetch deployed kogito supporting service", "instance", name, "Namespace", namespace)
	instance := &v1beta1.KogitoSupportingService{}
	if exists, resultErr := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, instance); resultErr != nil {
		log.Error(resultErr, "Error occurs while fetching deployed kogito supporting service", "Instance", name)
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		log.Debug("Successfully fetch deployed kogito supporting reference", "kogitoSupportingService", name)
		return instance, nil
	}
}
