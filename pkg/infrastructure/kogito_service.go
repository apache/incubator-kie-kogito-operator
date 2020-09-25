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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/url"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
)

// getSingletonKogitoServiceRoute gets the route from a kogito service that's unique in the given namespace
func getSingletonKogitoServiceRoute(client *client.Client, namespace string, serviceListRef v1alpha1.KogitoServiceList) (string, error) {
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, serviceListRef); err != nil {
		return "", err
	}
	if serviceListRef.GetItemsCount() > 0 {
		return serviceListRef.GetItemAt(0).GetStatus().GetExternalURI(), nil
	}
	return "", nil
}

// injectURLIntoKogitoApps will query for every KogitoApp in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
func injectURLIntoKogitoApps(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, serviceListRef v1alpha1.KogitoServiceList) error {
	log.Debugf("Querying KogitoApps in the namespace '%s' to inject a route ", namespace)

	deployments, err := getKogitoRuntimeDeployments(namespace, client)
	if err != nil {
		return err
	}
	log.Debugf("Found %s KogitoApps in the namespace '%s' ", len(deployments), namespace)
	var serviceEndpoints ServiceEndpoints
	if len(deployments) > 0 {
		log.Debug("Querying route to inject into KogitoApps")
		serviceEndpoints, err = getServiceEndpoints(client, namespace, serviceHTTPRouteEnv, serviceWSRouteEnv, serviceListRef)
		if err != nil {
			return err
		}
		log.Debugf("The route is '%s'", serviceEndpoints.HTTPRouteURI)
	}

	for _, dc := range deployments {
		// here we compare the current value to avoid updating the app every time
		if len(dc.Spec.Template.Spec.Containers) == 0 {
			break
		}
		updateHTTP := framework.GetEnvVarFromContainer(serviceEndpoints.HTTPRouteEnv, &dc.Spec.Template.Spec.Containers[0]) != serviceEndpoints.HTTPRouteURI
		updateWS := framework.GetEnvVarFromContainer(serviceEndpoints.WSRouteEnv, &dc.Spec.Template.Spec.Containers[0]) != serviceEndpoints.WSRouteURI
		if updateHTTP {
			log.Debugf("Updating dc '%s' to inject route %s ", dc.GetName(), serviceEndpoints.HTTPRouteURI)
			framework.SetEnvVar(serviceEndpoints.HTTPRouteEnv, serviceEndpoints.HTTPRouteURI, &dc.Spec.Template.Spec.Containers[0])
		}
		if updateWS {
			log.Debugf("Updating dc '%s' to inject route %s ", dc.GetName(), serviceEndpoints.WSRouteURI)
			framework.SetEnvVar(serviceEndpoints.WSRouteEnv, serviceEndpoints.WSRouteURI, &dc.Spec.Template.Spec.Containers[0])
		}
		// update only once
		if updateWS || updateHTTP {
			if err := kubernetes.ResourceC(client).Update(&dc); err != nil {
				return err
			}
		}
	}
	return nil
}

func getServiceEndpoints(client *client.Client, namespace string, serviceHTTPRouteEnv string, serviceWSRouteEnv string, serviceListRef v1alpha1.KogitoServiceList) (endpoints ServiceEndpoints, err error) {
	route := ""
	endpoints = ServiceEndpoints{HTTPRouteEnv: serviceHTTPRouteEnv, WSRouteEnv: serviceWSRouteEnv}
	route, err = getSingletonKogitoServiceRoute(client, namespace, serviceListRef)
	if err != nil {
		return
	}
	if len(route) > 0 {
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
	}
	return
}

// FetchKogitoServiceURI provide kogito service URI for given instance name
func FetchKogitoServiceURI(cli *client.Client, name, namespace string) (string, error) {
	log.Debugf("Fetching kogito service instance URI.")
	service := &v1.Service{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, service); err != nil {
		return "", err
	} else if !exits {
		return "", fmt.Errorf("service with name %s not exist for Kogito service instance in given namespace %s", name, namespace)
	}
	port := service.Spec.Ports[0]
	uri := fmt.Sprintf("%s://%s:%d", port.Name, service.Name, port.TargetPort.IntVal)
	log.Debugf("kogito service instance URI : %s", uri)
	return uri, nil
}
