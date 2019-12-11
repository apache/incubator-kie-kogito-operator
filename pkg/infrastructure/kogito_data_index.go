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
	kogitoDataIndexHttpRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	kogitoDataIndexWsRouteEnv   = "KOGITO_DATAINDEX_WS_URL"
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
	routeHttp := ""
	routeWs := ""
	if len(kogitoApps.Items) > 0 {
		log.Debug("Querying Data Index route to inject into KogitoApps ")
		var err error
		routeHttp, routeWs, err = getKogitoDataIndexURLs(client, namespace)
		if err != nil {
			return err
		}
		log.Debugf("Data Index route is '%s'", routeHttp)
	}

	for _, kogitoApp := range kogitoApps.Items {
		// here we compare the current value to avoid updating the app every time
		updateHttp := util.GetEnvValue(kogitoDataIndexHttpRouteEnv, kogitoApp.Spec.Env) != routeHttp
		updateWs := util.GetEnvValue(kogitoDataIndexWsRouteEnv, kogitoApp.Spec.Env) != routeWs
		if updateHttp {
			log.Debugf("Updating kogitoApp '%s' to inject route %s ", kogitoApp.GetName(), routeHttp)
			kogitoApp.Spec.Env = util.AppendOrReplaceEnv(v1alpha1.Env{Name: kogitoDataIndexHttpRouteEnv, Value: routeHttp}, kogitoApp.Spec.Env)
		}
		if updateWs {
			log.Debugf("Updating kogitoApp '%s' to inject route %s ", kogitoApp.GetName(), routeWs)
			kogitoApp.Spec.Env = util.AppendOrReplaceEnv(v1alpha1.Env{Name: kogitoDataIndexWsRouteEnv, Value: routeWs}, kogitoApp.Spec.Env)
		}
		// update only once
		if updateWs || updateHttp {
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

func getKogitoDataIndexURLs(client *client.Client, namespace string) (httpUrl, wsUrl string, err error) {
	route := ""
	httpUrl = ""
	wsUrl = ""
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
		httpUrl = routeURL.String()
		if httpScheme == routeURL.Scheme {
			wsUrl = fmt.Sprintf("%s://%s", webSocketScheme, routeURL.Host)
		} else {
			wsUrl = fmt.Sprintf("%s://%s", webSocketSecureScheme, routeURL.Host)
		}
	}
	return
}
