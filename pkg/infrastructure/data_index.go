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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// DefaultDataIndexImageName is just the image name for the Data Index Service
	DefaultDataIndexImageName = "kogito-data-index"
	// DefaultDataIndexName is the default name for the Data Index instance service
	DefaultDataIndexName = "data-index"
	// DefaultDataIndexPersistence Defines which Data Index Service persistence implementation will be used.
	// If empty the kogito-data-index image defaults to 'infinispan'. Allowed types are 'infinispan' and 'mongodb'.
	DefaultDataIndexPersistence = "infinispan"
	// DataIndexPersistenceTypeProp Defines the system property to be injected on runtime services to define the
	// persistence type
	DataIndexPersistenceTypeProp = "kogito.persistence.type"
	// DataIndexPersistenceEnv Holds the environment variable name to be used by the data-index image
	DataIndexPersistenceEnv = "DATA_INDEX_PERSISTENCE"
	// InfinispanPersistenceProvider Infinispan Persistence Provider
	infinispanPersistenceProvider = DefaultDataIndexPersistence
	// MongoDBPersistenceProvider MongoDB Persistence Provider
	mongoDBPersistenceProvider = "mongodb"

	dataIndexHTTPRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	dataIndexWSRouteEnv   = "KOGITO_DATAINDEX_WS_URL"

	webSocketScheme       = "ws"
	webSocketSecureScheme = "wss"
	httpScheme            = "http"
)

// PersistenceProvider describes the possible persistence provider
type PersistenceProvider string

var (
	// PersistenceProviders Map holds the supported persistence providers
	PersistenceProviders = map[PersistenceProvider]string{
		infinispanPersistenceProvider: "infinispan",
		mongoDBPersistenceProvider:    "mongodb",
	}
)

// InjectDataIndexURLIntoKogitoRuntimeServices will query for every KogitoRuntime in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoKogitoRuntimeServices(client *client.Client, namespace string) error {
	log.Debugf("Injecting Data-Index Route in kogito apps")
	return injectURLIntoKogitoApps(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, &v1alpha1.KogitoDataIndexList{})
}

// InjectDataIndexURLIntoKogitoRuntimeDeployment will inject data-index route URL in to kogito runtime deployment env var
func InjectDataIndexURLIntoKogitoRuntimeDeployment(client *client.Client, namespace string, runtimeDeployment *appsv1.Deployment) error {
	log.Debug("Querying Data Index route to inject into Kogito runtime ")
	dataIndexEndpoints, err := GetDataIndexEndpoints(client, namespace)
	if err != nil {
		return err
	}
	log.Debugf("Data Index route is '%s'", dataIndexEndpoints.HTTPRouteURI)
	updateDataIndexURLIntoKogitoRuntimeEnv(runtimeDeployment, dataIndexEndpoints)
	return nil
}

func updateDataIndexURLIntoKogitoRuntimeEnv(deployment *appsv1.Deployment, dataIndexEndpoints ServiceEndpoints) (updateHTTP bool, updateWS bool) {
	// here we compare the current value to avoid updating the app every time
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		updateHTTP = framework.GetEnvVarFromContainer(dataIndexEndpoints.HTTPRouteEnv, &deployment.Spec.Template.Spec.Containers[0]) != dataIndexEndpoints.HTTPRouteURI
		updateWS = framework.GetEnvVarFromContainer(dataIndexEndpoints.WSRouteEnv, &deployment.Spec.Template.Spec.Containers[0]) != dataIndexEndpoints.WSRouteURI
		if updateHTTP {
			log.Debugf("Updating dc '%s' to inject route %s ", deployment.GetName(), dataIndexEndpoints.HTTPRouteURI)
			framework.SetEnvVar(dataIndexEndpoints.HTTPRouteEnv, dataIndexEndpoints.HTTPRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
		if updateWS {
			log.Debugf("Updating dc '%s' to inject route %s ", deployment.GetName(), dataIndexEndpoints.WSRouteURI)
			framework.SetEnvVar(dataIndexEndpoints.WSRouteEnv, dataIndexEndpoints.WSRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
	}
	return
}

// GetDataIndexEndpoints queries for the Data Index URIs
func GetDataIndexEndpoints(client *client.Client, namespace string) (dataIndexEndpoints ServiceEndpoints, err error) {
	return getServiceEndpoints(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, &v1alpha1.KogitoDataIndexList{})
}
