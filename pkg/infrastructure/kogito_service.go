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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"net/url"
	"os"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

const (
	webSocketScheme       = "ws"
	webSocketSecureScheme = "wss"
	httpScheme            = "http"

	envVarKogitoServiceURL = "LOCAL_KOGITO_SERVICE_URL"
)

// injectSupportingServiceURLIntoKogitoRuntime will query for every KogitoApp in the given namespace to inject the Supporting service route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
// it will call when supporting service reconcile
func injectSupportingServiceURLIntoKogitoRuntime(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, resourceType v1alpha1.ServiceType) error {
	serviceEndpoints, err := getServiceEndpoints(client, namespace, serviceHTTPRouteEnv, serviceWSRouteEnv, resourceType)
	if err != nil {
		return err
	}
	if serviceEndpoints != nil {

		log.Debugf("The %s route is '%s'", resourceType, serviceEndpoints.HTTPRouteURI)

		log.Debugf("Querying KogitoRuntime instances in the namespace '%s' to inject a route ", namespace)
		deployments, err := getKogitoRuntimeDeployments(namespace, client)
		if err != nil {
			return err
		}
		log.Debugf("Found %s KogitoRuntime instances in the namespace '%s' ", len(deployments), namespace)
		if len(deployments) == 0 {
			log.Debugf("No deployment found for KogitoRuntime, skipping to inject %s URL into KogitoRuntime", resourceType)
			return nil
		}
		log.Debugf("Querying %s route to inject into KogitoRuntimes", resourceType)

		for _, dc := range deployments {
			updateHTTP, updateWS := updateServiceEndpointIntoDeploymentEnv(&dc, serviceEndpoints)
			// update only once
			if updateWS || updateHTTP {
				if err := kubernetes.ResourceC(client).Update(&dc); err != nil {
					return err
				}
			}
		}
	}
	log.Debugf("Service Endpoint is nil")
	return nil
}

// InjectDataIndexURLIntoDeployment will inject Supporting service route URL in to kogito runtime deployment env var
// It will call when Kogito runtime reconcile
func injectSupportingServiceURLIntoDeployment(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, deployment *appsv1.Deployment, resourceType v1alpha1.ServiceType) error {
	log.Debug("Querying Data Index route to inject into Kogito runtime ")
	dataIndexEndpoints, err := getServiceEndpoints(client, namespace, serviceHTTPRouteEnv, serviceWSRouteEnv, resourceType)
	if err != nil {
		return err
	}
	if dataIndexEndpoints != nil {
		log.Debugf("The %s route is '%s'", resourceType, dataIndexEndpoints.HTTPRouteURI)
		updateServiceEndpointIntoDeploymentEnv(deployment, dataIndexEndpoints)
	}
	return nil
}

// GetKogitoServiceEndpoint gets the endpoint depending on
// if the envVarKogitoServiceURL is set (for when running
// operator locally). Else, the internal endpoint is
// returned.
func GetKogitoServiceEndpoint(kogitoService v1alpha1.KogitoService) string {
	externalURL := os.Getenv(envVarKogitoServiceURL)
	if len(externalURL) > 0 {
		return externalURL
	}
	return getKogitoServiceURL(kogitoService)
}

func getServiceEndpoints(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, resourceType v1alpha1.ServiceType) (endpoints *ServiceEndpoints, err error) {
	route := ""
	route, err = getKogitoSupportingServiceRoute(client, namespace, resourceType)
	if err != nil {
		return
	}
	if len(route) > 0 {
		endpoints = &ServiceEndpoints{
			HTTPRouteEnv: serviceHTTPRouteEnv,
			WSRouteEnv:   serviceWSRouteEnv,
		}
		var routeURL *url.URL
		routeURL, err = url.Parse(route)
		if err != nil {
			log.Warnf("Failed to parse route url (%s), set to empty: %s", route, err)
			return
		}
		endpoints.HTTPRouteURI = routeURL.String()
		if httpScheme == routeURL.Scheme {
			endpoints.WSRouteURI = fmt.Sprintf("%s://%s", webSocketScheme, routeURL.Host)
		} else {
			endpoints.WSRouteURI = fmt.Sprintf("%s://%s", webSocketSecureScheme, routeURL.Host)
		}
		return
	}
	return nil, nil
}

func updateServiceEndpointIntoDeploymentEnv(deployment *appsv1.Deployment, serviceEndpoints *ServiceEndpoints) (updateHTTP bool, updateWS bool) {
	// here we compare the current value to avoid updating the app every time
	if len(deployment.Spec.Template.Spec.Containers) > 0 && serviceEndpoints != nil {
		if len(serviceEndpoints.HTTPRouteEnv) > 0 {
			updateHTTP = framework.GetEnvVarFromContainer(serviceEndpoints.HTTPRouteEnv, &deployment.Spec.Template.Spec.Containers[0]) != serviceEndpoints.HTTPRouteURI
		}
		if len(serviceEndpoints.WSRouteEnv) > 0 {
			updateWS = framework.GetEnvVarFromContainer(serviceEndpoints.WSRouteEnv, &deployment.Spec.Template.Spec.Containers[0]) != serviceEndpoints.WSRouteURI
		}
		if updateHTTP {
			log.Debugf("Updating dc '%s' to inject route %s ", deployment.GetName(), serviceEndpoints.HTTPRouteURI)
			framework.SetEnvVar(serviceEndpoints.HTTPRouteEnv, serviceEndpoints.HTTPRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
		if updateWS {
			log.Debugf("Updating dc '%s' to inject route %s ", deployment.GetName(), serviceEndpoints.WSRouteURI)
			framework.SetEnvVar(serviceEndpoints.WSRouteEnv, serviceEndpoints.WSRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
	}
	return
}

// getKogitoServiceURL provides kogito service URL for given instance name
func getKogitoServiceURL(service v1alpha1.KogitoService) string {
	log.Debugf("Creating kogito service instance URL.")
	// resolves to http://servicename.mynamespace for example
	url := fmt.Sprintf("http://%s.%s", service.GetName(), service.GetNamespace())
	log.Debugf("kogito service instance URL : %s", url)
	return url
}
