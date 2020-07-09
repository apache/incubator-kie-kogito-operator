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
)

const (
	// DefaultDataIndexImageName is just the image name for the Data Index Service
	DefaultDataIndexImageName = "kogito-data-index"
	// DefaultDataIndexName is the default name for the Data Index instance service
	DefaultDataIndexName = "data-index"

	dataIndexHTTPRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	dataIndexWSRouteEnv   = "KOGITO_DATAINDEX_WS_URL"
)

// InjectDataIndexURLIntoKogitoApps will query for every KogitoApp in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoKogitoApps(client *client.Client, namespace string) error {
	log.Debugf("Injecting Data-Index Route in kogito apps")
	return injectURLIntoKogitoApps(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, &v1alpha1.KogitoDataIndexList{})
}

// GetDataIndexEndpoints queries for the Data Index URIs
func GetDataIndexEndpoints(client *client.Client, namespace string) (dataIndexEndpoints ServiceEndpoints, err error) {
	return getServiceEndpoints(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, &v1alpha1.KogitoDataIndexList{})
}
