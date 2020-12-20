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

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

const (
	webSocketScheme       = "ws"
	webSocketSecureScheme = "wss"
	httpScheme            = "http"

	envVarKogitoServiceURL = "LOCAL_KOGITO_SERVICE_URL"

	// KogitoHomeDir path for Kogito home mounted within the pod of a Kogito Service
	KogitoHomeDir = "/home/kogito"
)

// injectSupportingServiceURLIntoKogitoRuntime will query for every KogitoApp in the given namespace to inject the Supporting service route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
// it will call when supporting service reconcile
func injectSupportingServiceURLIntoKogitoRuntime(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, resourceType v1beta1.ServiceType) error {
	serviceEndpoints, err := getServiceEndpoints(client, namespace, serviceHTTPRouteEnv, serviceWSRouteEnv, resourceType)
	if err != nil {
		return err
	}
	if serviceEndpoints != nil {

		log.Debug("", "resourceType", resourceType, "route", serviceEndpoints.HTTPRouteURI)

		log.Debug("Querying KogitoRuntime instances to inject a route ", "namespace", namespace)
		deployments, err := getKogitoRuntimeDeployments(namespace, client)
		if err != nil {
			return err
		}
		log.Debug("", "Found KogitoRuntime instances", len(deployments), "namespace", namespace)
		if len(deployments) == 0 {
			log.Debug("No deployment found for KogitoRuntime, skipping to inject request resource type URL into KogitoRuntime", "request resource type", resourceType)
			return nil
		}
		log.Debug("Querying resource route to inject into KogitoRuntimes", "resource", resourceType)

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
	log.Debug("Service Endpoint is nil")
	return nil
}

// InjectDataIndexURLIntoDeployment will inject Supporting service route URL in to kogito runtime deployment env var
// It will call when Kogito runtime reconcile
func injectSupportingServiceURLIntoDeployment(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, deployment *appsv1.Deployment, resourceType v1beta1.ServiceType) error {
	log.Debug("Querying Data Index route to inject into Kogito runtime ")
	dataIndexEndpoints, err := getServiceEndpoints(client, namespace, serviceHTTPRouteEnv, serviceWSRouteEnv, resourceType)
	if err != nil {
		return err
	}
	if dataIndexEndpoints != nil {
		log.Debug("", "resourceType", resourceType, "route", dataIndexEndpoints.HTTPRouteURI)
		updateServiceEndpointIntoDeploymentEnv(deployment, dataIndexEndpoints)
	}
	return nil
}

// GetKogitoServiceEndpoint gets the endpoint depending on
// if the envVarKogitoServiceURL is set (for when running
// operator locally). Else, the internal endpoint is
// returned.
func GetKogitoServiceEndpoint(kogitoService v1beta1.KogitoService) string {
	externalURL := os.Getenv(envVarKogitoServiceURL)
	if len(externalURL) > 0 {
		return externalURL
	}
	return getKogitoServiceURL(kogitoService)
}

func getServiceEndpoints(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, resourceType v1beta1.ServiceType) (endpoints *ServiceEndpoints, err error) {
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
			log.Error(err, "Failed to parse route url, set to empty", "route url", route)
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
			log.Debug("Updating dc to inject route", "dc", deployment.GetName(), "route", serviceEndpoints.HTTPRouteURI)
			framework.SetEnvVar(serviceEndpoints.HTTPRouteEnv, serviceEndpoints.HTTPRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
		if updateWS {
			log.Debug("Updating dc to inject route ", "dc", deployment.GetName(), "route", serviceEndpoints.WSRouteURI)
			framework.SetEnvVar(serviceEndpoints.WSRouteEnv, serviceEndpoints.WSRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
	}
	return
}

// getKogitoServiceURL provides kogito service URL for given instance name
func getKogitoServiceURL(service v1beta1.KogitoService) string {
	log.Debug("Creating kogito service instance URL.")
	// resolves to http://servicename.mynamespace for example
	url := fmt.Sprintf("http://%s.%s", service.GetName(), service.GetNamespace())
	log.Debug("", "kogito service instance URL", url)
	return url
}

// AddFinalizer add finalizer to provide KogitoService instance
func AddFinalizer(client *client.Client, instance v1beta1.KogitoService) error {
	if len(instance.GetFinalizers()) < 1 && instance.GetDeletionTimestamp() == nil {
		log.Debug("Adding Finalizer for the KogitoService")
		instance.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer"})

		// Update CR
		if err := kubernetes.ResourceC(client).Update(instance); err != nil {
			log.Error(err, "Failed to update finalizer in KogitoService")
			return err
		}
		log.Debug("Successfully added finalizer into KogitoService instance", "instance", instance.GetName())
	}
	return nil
}

// HandleFinalization remove owner reference of provided Kogito service from KogitoInfra instances and remove finalizer from KogitoService
func HandleFinalization(client *client.Client, instance v1beta1.KogitoService) error {
	// Remove KogitoSupportingService ownership from referred KogitoInfra instances
	if err := RemoveKogitoInfraOwnership(client, instance); err != nil {
		return err
	}

	// Update finalizer to allow delete CR
	log.Debug("Removing finalizer from KogitoService instance", "instance", instance.GetName())
	instance.SetFinalizers(nil)
	if err := kubernetes.ResourceC(client).Update(instance); err != nil {
		log.Error(err, "Error occurs while removing finalizer from KogitoService instance", "instance", instance.GetName())
		return err
	}
	log.Debug("Successfully removed finalizer from KogitoService instance", "instance", instance.GetName())
	return nil
}
