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

package services

import (
	"encoding/json"
	"fmt"
	grafanav1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"strings"
)

// GrafanaDashboard is a structure that contains the fetched dashboards
type GrafanaDashboard struct {
	Name             string
	RawJSONDashboard string
}

const (
	// dashboardPath which the dashboards are fetched
	dashboardsPath = "/monitoring/dashboards/"
)

func fetchGrafanaDashboards(cli *client.Client, instance v1alpha1.KogitoService) ([]GrafanaDashboard, error) {
	available, err := IsDeploymentAvailable(cli, instance)
	if err != nil {
		return nil, err
	}
	if !available {
		log.Debugf("Deployment not available yet for KogitoService %s ", instance.GetName())
		return nil, nil
	}

	svcURL := infrastructure.GetKogitoServiceEndpoint(instance)
	dashboardNames, err := fetchGrafanaDashboardNamesForURL(svcURL)
	if err != nil {
		return nil, err
	}

	return fetchDashboards(svcURL, dashboardNames)
}

func fetchGrafanaDashboardNamesForURL(serverURL string) ([]string, error) {
	dashboardsURL := fmt.Sprintf("%s%s%s", serverURL, dashboardsPath, "list.json")
	resp, err := http.Get(dashboardsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		log.Debugf("Dashboard list not found, the monitoring addon is disabled on the service. There are no dashboards to deploy.")
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received NOT expected status code %d while making GET request to %s ", resp.StatusCode, dashboardsURL)
	}
	var dashboardNames []string
	if err := json.NewDecoder(resp.Body).Decode(&dashboardNames); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s into dashboard names", dashboardsURL)
	}

	return dashboardNames, nil
}

func fetchDashboards(serverURL string, dashboardNames []string) ([]GrafanaDashboard, error) {
	var dashboards []GrafanaDashboard
	for _, name := range dashboardNames {
		dashboardURL := fmt.Sprintf("%s%s%s", serverURL, dashboardsPath, name)
		if dashboard, err := fetchDashboard(name, dashboardURL); err != nil {
			return nil, err
		} else if dashboard != nil {
			dashboards = append(dashboards, *dashboard)
		}
	}
	return dashboards, nil
}

// we create a separate function to be able to `defer` the HTTP response after the function call.
func fetchDashboard(name, dashboardURL string) (*GrafanaDashboard, error) {
	resp, err := http.Get(dashboardURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		log.Debugf("Dashboard %s not found, ignoring the resource.", name)
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received NOT expected status code %d while making GET request to %s ", resp.StatusCode, dashboardURL)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &GrafanaDashboard{Name: name, RawJSONDashboard: string(bodyBytes)}, nil
}

func configureGrafanaDashboards(client *client.Client, kogitoService v1alpha1.KogitoService, scheme *runtime.Scheme, namespace string) error {
	dashboards, err := fetchGrafanaDashboards(client, kogitoService)
	if err != nil {
		return errorForDashboards(err)
	}

	err = deployGrafanaDashboards(dashboards, client, kogitoService, scheme, namespace)
	return err
}

func deployGrafanaDashboards(dashboards []GrafanaDashboard, cli *client.Client, kogitoService v1alpha1.KogitoService, scheme *runtime.Scheme, namespace string) error {
	for _, dashboard := range dashboards {
		resourceName := strings.ReplaceAll(strings.ToLower(dashboard.Name), ".json", "")
		dashboardDefinition := &grafanav1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: namespace,
				Labels: map[string]string{
					framework.LabelAppKey: kogitoService.GetName(),
				},
			},
			Spec: grafanav1.GrafanaDashboardSpec{
				Json: dashboard.RawJSONDashboard,
				Name: dashboard.Name,
			},
		}
		if err := kubernetes.ResourceC(cli).CreateIfNotExistsForOwner(dashboardDefinition, kogitoService, scheme); err != nil {
			log.Warnf("Error occurs while creating dashboard %s, not going to reconcile the resource: %v", dashboard.Name, err)
			return err
		}
		log.Infof("Successfully created grafana dashboard %s", dashboard.Name)
	}
	return nil
}
