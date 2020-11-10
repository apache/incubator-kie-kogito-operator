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
	grafanav1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_fetchDashboardNames(t *testing.T) {
	dashboardNames := `["dashboard1.json", "dashboard2.json"]`

	server := mockKogitoSvcReplies(t, serverHandler{Path: dashboardsPath + "list.json", JSONResponse: dashboardNames})
	defer server.Close()

	dashboards, err := fetchGrafanaDashboardNamesForURL(server.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, dashboards)
	assert.Equal(t, "dashboard1.json", dashboards[0])
	assert.Equal(t, "dashboard2.json", dashboards[1])
}

func Test_fetchDashboards(t *testing.T) {
	dashboardNames := `["dashboard1.json", "dashboard2.json"]`
	dashboard1 := `mydashboard1`
	dashboard2 := `mydashboard2`

	handlers := []serverHandler{
		{
			Path:         dashboardsPath + "list.json",
			JSONResponse: dashboardNames,
		},
		{
			Path:         dashboardsPath + "dashboard1.json",
			JSONResponse: dashboard1,
		},
		{
			Path:         dashboardsPath + "dashboard2.json",
			JSONResponse: dashboard2,
		},
	}

	server := mockKogitoSvcReplies(t, handlers...)
	defer server.Close()

	fetchedDashboardNames, err := fetchGrafanaDashboardNamesForURL(server.URL)
	assert.NoError(t, err)
	dashboards, err := fetchDashboards(server.URL, fetchedDashboardNames)
	assert.NoError(t, err)
	assert.Equal(t, len(fetchedDashboardNames), len(dashboards))
	assert.Equal(t, dashboard1, dashboards[0].RawJSONDashboard)
	assert.Equal(t, dashboard2, dashboards[1].RawJSONDashboard)
	assert.Equal(t, fetchedDashboardNames[0], dashboards[0].Name)
}

func Test_serviceDeployer_DeployGrafanaDashboards(t *testing.T) {
	replicas := int32(1)
	service := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-kogito-runtime",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(service).OnOpenShift().Build()

	dashboards := []GrafanaDashboard{
		{
			Name:             "mydashboard",
			RawJSONDashboard: "[]",
		},
		{
			Name:             "myseconddashboard",
			RawJSONDashboard: "[]",
		},
	}

	err := deployGrafanaDashboards(dashboards, cli, service, meta.GetRegisteredSchema(), t.Name())
	assert.NoError(t, err)

	dashboard := &grafanav1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mydashboard",
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, dashboard)

	dashboard = &grafanav1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myseconddashboard",
			Namespace: t.Name(),
		},
	}
	test.AssertFetchMustExist(t, cli, dashboard)
}
