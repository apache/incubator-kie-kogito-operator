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
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// DefaultMgmtConsoleName ...
	DefaultMgmtConsoleName = "management-console"
	// DefaultMgmtConsoleImageName ...
	DefaultMgmtConsoleImageName = "kogito-management-console"
)

// GetManagementConsoleEndpoint gets the route for the Management Console deployed in the given namespace
func GetManagementConsoleEndpoint(client *client.Client, namespace string) (*ServiceEndpoints, error) {
	return getServiceEndpoints(client, namespace, "", "", v1alpha1.MgmtConsole)
}

// getKogitoRuntimeDeployments gets all dcs owned by KogitoRuntime services within the given namespace
func getMgmtConsoleDeployment(namespace string, cli *client.Client) (*appsv1.Deployment, error) {
	mgmtConsole, err := getKogitoSupportingService(cli, namespace, v1alpha1.MgmtConsole)
	if err != nil {
		return nil, err
	} else if mgmtConsole == nil {
		log.Debugf("Not found Mgmt console service in namespace %s", namespace)
		return nil, nil
	}
	log.Debugf("Found Mgmt console services in the namespace '%s' ", namespace)

	dcs := &appsv1.DeploymentList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	log.Debug("Looking for Deployments owned by MgmtConsole")
	for _, dc := range dcs.Items {
		for _, owner := range dc.OwnerReferences {
			if owner.UID == mgmtConsole.UID {
				return &dc, nil
			}
		}
	}
	return nil, nil
}
