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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fetchDashboards(t *testing.T) {
	dashboardNames := `["dashboard1", "dashboard2"]`

	server := mockKogitoSvcReplies(t, dashboardNames)
	defer server.Close()

	dashboards, err := FetchGrafanaDashboardNamesForURL(server.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, dashboards)
}

func mockKogitoSvcReplies(t *testing.T, jsonResponse string) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(dashboardsPath, func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(jsonResponse))
		assert.NoError(t, err)
	})

	return httptest.NewServer(handler)
}
