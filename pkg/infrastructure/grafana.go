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
	grafana "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	grafanaServerGroup = "integreatly.org"
	// GrafanaOperatorName is the default Grafana operator name
	GrafanaOperatorName = "grafana-operator"
)

// IsGrafanaOperatorAvailable verify if Grafana Operator is running in the given namespace and the CRD is available
// Deprecated: rethink the way we check for the operator since the deployment resource could be in another namespace if installed cluster wide
func IsGrafanaOperatorAvailable(cli *client.Client, namespace string) (available bool, err error) {
	log.Debugf("Checking if Grafana Operator is available in the namespace %s", namespace)
	available = false
	if IsGrafanaAvailable(cli) {
		log.Debugf("Grafana CRDs available. Checking if Grafana Operator is deployed in the namespace %s", namespace)
		// then check if there's an Grafana Operator deployed
		deployment := &v1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: GrafanaOperatorName}}
		exists := false
		var err error
		if exists, err = kubernetes.ResourceC(cli).Fetch(deployment); err != nil {
			return false, nil
		}
		if exists {
			log.Debugf("Grafana Operator is available in the namespace %s", namespace)
			return true, nil
		}
	} else {
		log.Debug("Couldn't find Grafana CRDs")
	}
	log.Debugf("Looks like Grafana Operator is not available in the namespace %s", namespace)

	return
}

// IsGrafanaAvailable checks if Grafana CRD is available in the cluster
func IsGrafanaAvailable(client *client.Client) bool {
	return client.HasServerGroup(grafanaServerGroup)
}

// GetGrafanaDefaultResource returns a Grafana resource with default configuration
func GetGrafanaDefaultResource(name, namespace string) *grafana.Grafana {
	return &grafana.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: grafana.GrafanaSpec{
			Config: grafana.GrafanaConfig{
				Auth: &grafana.GrafanaConfigAuth{
					DisableSignoutMenu: newTrue(),
				},
				AuthAnonymous: &grafana.GrafanaConfigAuthAnonymous{
					Enabled: newTrue(),
				},
			},
			Ingress: &grafana.GrafanaIngress{
				Enabled: true,
			},
		},
	}
}

func newTrue() *bool {
	b := true
	return &b
}
