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
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	grafanav1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaDashboard is a structure that contains the fetched dashboards
type GrafanaDashboard struct {
	Name             string
	RawJSONDashboard string
}

const (
	// dashboardPath which the dashboards are fetched
	dashboardsPath = "/monitoring/dashboards/"

	// GrafanaDashboardAppName label app name to be used when a GrafanaDashboard is created
	GrafanaDashboardAppName = "grafana"
)

// FetchGrafanaDashboards fetches the grafana dashboards from the KogitoService
func FetchGrafanaDashboards(cli *client.Client, instance v1alpha1.KogitoService) ([]GrafanaDashboard, error) {
	available, err := isDeploymentAvailable(cli, instance)
	if err != nil {
		return nil, err
	}
	if !available {
		log.Debugf("Deployment not available yet for KogitoService %s ", instance.GetName())
		return nil, nil
	}

	svcURL := infrastructure.CreateKogitoServiceURI(instance)
	dashboardNames, err := FetchGrafanaDashboardNamesForURL(svcURL)
	if err != nil {
		return nil, err
	}

	return FetchDashboards(svcURL, dashboardNames)
}

// FetchGrafanaDashboardNamesForURL fetches the dashboard names available on the kogito runtime service if the monitoring addon is enabled
func FetchGrafanaDashboardNamesForURL(serverURL string) ([]string, error) {
	dashboardsURL := fmt.Sprintf("%s%s%s", serverURL, dashboardsPath, "list.json")
	resp, err := http.Get(dashboardsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received NOT expected status code %d while making GET request to %s ", resp.StatusCode, dashboardsURL)
	}
	var dashboardNames []string
	if err := json.NewDecoder(resp.Body).Decode(&dashboardNames); err != nil {
		return nil, fmt.Errorf("Failed to decode response from %s into dashboard names", dashboardsURL)
	}

	return dashboardNames, nil
}

// FetchDashboards fetches the json grafana dashboard from the kogito runtime service
func FetchDashboards(serverURL string, dashboardNames []string) ([]GrafanaDashboard, error) {
	var dashboards []GrafanaDashboard
	for _, name := range dashboardNames {
		dashboardURL := fmt.Sprintf("%s%s%s", serverURL, dashboardsPath, name)
		resp, err := http.Get(dashboardURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Received NOT expected status code %d while making GET request to %s ", resp.StatusCode, dashboardURL)
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		dashboards = append(dashboards, GrafanaDashboard{Name: name, RawJSONDashboard: string(bodyBytes)})
	}
	return dashboards, nil
}

func configureGrafanaDashboards(client *client.Client, kogitoService v1alpha1.KogitoService, namespace string) (time.Duration, error) {
	dashboards, err := FetchGrafanaDashboards(client, kogitoService)
	if err != nil {
		return reconciliationPeriodAfterDashboardsError, err
	}

	reconcileAfter, err := deployGrafanaDashboards(dashboards, client, namespace)

	return reconcileAfter, err
}

func deployGrafanaDashboards(dashboards []GrafanaDashboard, cli *client.Client, namespace string) (time.Duration, error) {
	for _, dashboard := range dashboards {
		resourceName := strings.ReplaceAll(strings.ToLower(dashboard.Name), ".json", "")
		dashboardDefinition := &grafanav1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: namespace,
				Labels: map[string]string{
					"app": GrafanaDashboardAppName,
				},
			},
			Spec: grafanav1.GrafanaDashboardSpec{
				Json: dashboard.RawJSONDashboard,
				Name: dashboard.Name,
			},
		}
		if err := kubernetes.ResourceC(cli).Create(dashboardDefinition); err != nil {
			log.Warnf("Error occurs while creating dashboard %s, not going to reconcile the resource.", dashboard.Name, err)
		} else {
			log.Infof("Successfully created grafana dashboard %s", dashboard.Name)
		}
	}
	return 0, nil
}
