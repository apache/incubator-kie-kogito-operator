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
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// DefaultExplainabilityImageName is just the image name for the Explainability Service
	DefaultExplainabilityImageName = "kogito-explainability"
	// DefaultExplainabilityName is the default name for the Explainability instance service
	DefaultExplainabilityName = "explainability"

	explainabilityHTTPRouteEnv = "KOGITO_RUNTIME_HTTP_URL"
	explainabilityWSRouteEnv   = "KOGITO_RUNTIME_WS_URL"
)

// InjectKogitoRuntimeURLIntoKogitoExplainability will query for every KogitoExplainability in the given namespace to inject the KogitoRuntime route to each one
// Won't trigger an update if the KogitoExplainability already has the route set to avoid unnecessary reconciliation triggers
func InjectKogitoRuntimeURLIntoKogitoExplainability(client *client.Client, namespace string) error {
	log.Debugf("Injecting Kogito Runtime Route in kogito explainability")
	return injectURLIntoKogitoApps(client, namespace, explainabilityHTTPRouteEnv, explainabilityWSRouteEnv, &v1alpha1.KogitoRuntimeList{}, getKogitoExplainabilityDeployments)
}

// getKogitoExplainabilityDeployments gets all dcs owned by KogitoExplainability services within the given namespace
func getKogitoExplainabilityDeployments(namespace string, cli *client.Client) ([]appsv1.Deployment, error) {
	var kdcs []appsv1.Deployment
	kogitoExplainabilityServices := &v1alpha1.KogitoExplainabilityList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, kogitoExplainabilityServices); err != nil {
		return nil, err
	}
	log.Debugf("Found %d KogitoExplainability services in the namespace '%s' ", len(kogitoExplainabilityServices.Items), namespace)
	if len(kogitoExplainabilityServices.Items) == 0 {
		return kdcs, nil
	}
	dcs := &appsv1.DeploymentList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	log.Debug("Looking for DeploymentConfigs owned by KogitoRuntime")
	for _, dc := range dcs.Items {
		for _, owner := range dc.OwnerReferences {
			for _, app := range kogitoExplainabilityServices.Items {
				if owner.UID == app.UID {
					kdcs = append(kdcs, dc)
					break
				}
			}
		}
	}
	return kdcs, nil
}
