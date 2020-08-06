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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"net/url"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

const (
	// DefaultDataIndexImageName is just the image name for the Data Index Service
	DefaultDataIndexImageName = "kogito-data-index"
	// DefaultDataIndexName is the default name for the Data Index instance service
	DefaultDataIndexName = "data-index"

	dataIndexHTTPRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	dataIndexWSRouteEnv   = "KOGITO_DATAINDEX_WS_URL"
	webSocketScheme       = "ws"
	webSocketSecureScheme = "wss"
	httpScheme            = "http"
)

// InjectDataIndexURLIntoKogitoRuntimeServices will query for every KogitoRuntime in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoKogitoRuntimeServices(client *client.Client, namespace string) error {
	log.Debugf("Querying Kogito runtime services in the namespace '%s' to inject Data Index Route ", namespace)

	deployments, err := getKogitoRuntimeDeployments(namespace, client)
	if err != nil {
		return err
	}
	log.Debugf("Found %s KogitoRuntime services in the namespace '%s' ", len(deployments), namespace)
	var dataIndexEndpoints ServiceEndpoints
	if len(deployments) > 0 {
		log.Debug("Querying Data Index route to inject into Kogito runtime ")
		dataIndexEndpoints, err = GetDataIndexEndpoints(client, namespace)
		if err != nil {
			return err
		}
		log.Debugf("Data Index route is '%s'", dataIndexEndpoints.HTTPRouteURI)
	}

	for _, dc := range deployments {
		updateHTTP, updateWS := updateDataIndexURLIntoKogitoRuntimeEnv(&dc, dataIndexEndpoints)
		// update only once
		if updateWS || updateHTTP {
			if err := kubernetes.ResourceC(client).Update(&dc); err != nil {
				return err
			}
		}
	}
	return nil
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
	route := ""
	dataIndexEndpoints = ServiceEndpoints{HTTPRouteEnv: dataIndexHTTPRouteEnv, WSRouteEnv: dataIndexWSRouteEnv}
	route, err = getSingletonKogitoServiceRoute(client, namespace, &v1alpha1.KogitoDataIndexList{})
	if err != nil {
		return
	}
	if len(route) > 0 {
		var routeURL *url.URL
		routeURL, err = url.Parse(route)
		if err != nil {
			log.Warnf("Failed to parse data index route url (%s), set to empty: %s", route, err)
			return
		}
		dataIndexEndpoints.HTTPRouteURI = routeURL.String()
		if httpScheme == routeURL.Scheme {
			dataIndexEndpoints.WSRouteURI = fmt.Sprintf("%s://%s", webSocketScheme, routeURL.Host)
		} else {
			dataIndexEndpoints.WSRouteURI = fmt.Sprintf("%s://%s", webSocketSecureScheme, routeURL.Host)
		}
	}
	return
}
