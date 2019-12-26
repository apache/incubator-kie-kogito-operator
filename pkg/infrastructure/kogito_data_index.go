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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"net/url"
)

const (
	kogitoDataIndexHTTPRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	kogitoDataIndexWSRouteEnv   = "KOGITO_DATAINDEX_WS_URL"
	webSocketScheme             = "ws"
	webSocketSecureScheme       = "wss"
	httpScheme                  = "http"
)

// InjectDataIndexURLIntoKogitoApps will query for every KogitoApp in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoKogitoApps(client *client.Client, namespace string) error {
	log.Debugf("Querying KogitoApps in the namespace '%s' to inject Data Index Route ", namespace)
	kogitoApps := &v1alpha1.KogitoAppList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, kogitoApps); err != nil {
		return err
	}
	log.Debugf("Found %s KogitoApps in the namespace '%s' ", len(kogitoApps.Items), namespace)
	routeHTTP := ""
	routeWS := ""
	if len(kogitoApps.Items) > 0 {
		log.Debug("Querying Data Index route to inject into KogitoApps ")
		var err error
		routeHTTP, routeWS, err = getKogitoDataIndexURLs(client, namespace)
		if err != nil {
			return err
		}
		log.Debugf("Data Index route is '%s'", routeHTTP)
	}

	for _, kogitoApp := range kogitoApps.Items {
		// here we compare the current value to avoid updating the app every time
		updateHTTP := util.GetEnvValue(kogitoDataIndexHTTPRouteEnv, kogitoApp.Spec.Env) != routeHTTP
		updateWS := util.GetEnvValue(kogitoDataIndexWSRouteEnv, kogitoApp.Spec.Env) != routeWS
		if updateHTTP {
			log.Debugf("Updating kogitoApp '%s' to inject route %s ", kogitoApp.GetName(), routeHTTP)
			kogitoApp.Spec.Env = util.AppendOrReplaceEnv(v1alpha1.Env{Name: kogitoDataIndexHTTPRouteEnv, Value: routeHTTP}, kogitoApp.Spec.Env)
		}
		if updateWS {
			log.Debugf("Updating kogitoApp '%s' to inject route %s ", kogitoApp.GetName(), routeWS)
			kogitoApp.Spec.Env = util.AppendOrReplaceEnv(v1alpha1.Env{Name: kogitoDataIndexWSRouteEnv, Value: routeWS}, kogitoApp.Spec.Env)
		}
		// update only once
		if updateWS || updateHTTP {
			if err := kubernetes.ResourceC(client).Update(&kogitoApp); err != nil {
				return err
			}
		}
	}
	return nil
}

// getKogitoDataIndexRoute gets the deployed data index route
func getKogitoDataIndexRoute(client *client.Client, namespace string) (string, error) {
	route := ""
	dataIndexes := &v1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, dataIndexes); err != nil {
		return route, err
	}
	if len(dataIndexes.Items) > 0 {
		// should be only one data index guaranteed by OLM, but still we are looking for the first one
		route = dataIndexes.Items[0].Status.Route
	}
	return route, nil
}

func getKogitoDataIndexURLs(client *client.Client, namespace string) (httpURL, wsURL string, err error) {
	route := ""
	httpURL = ""
	wsURL = ""
	route, err = getKogitoDataIndexRoute(client, namespace)
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
		httpURL = routeURL.String()
		if httpScheme == routeURL.Scheme {
			wsURL = fmt.Sprintf("%s://%s", webSocketScheme, routeURL.Host)
		} else {
			wsURL = fmt.Sprintf("%s://%s", webSocketSecureScheme, routeURL.Host)
		}
	}
	return
}
