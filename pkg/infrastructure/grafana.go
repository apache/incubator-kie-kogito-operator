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
	"strings"

	grafana "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
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
		list := &v1.DeploymentList{}
		if err = kubernetes.ResourceC(cli).ListWithNamespace(namespace, list); err != nil {
			return
		}
		for _, grafana := range list.Items {
			for _, owner := range grafana.OwnerReferences {
				if strings.HasPrefix(owner.Name, GrafanaOperatorName) {
					available = true
					return
				}
			}
		}
	}
	return
}

// IsGrafanaAvailable checks if Grafana CRD is available in the cluster
func IsGrafanaAvailable(client *client.Client) bool {
	return client.HasServerGroup(grafanaServerGroup)
}

// GetGrafanaDefaultResource returns a Grafana resource with default configuration
func GetGrafanaDefaultResource(name, namespace string, defaultReplicas int32) *grafana.Grafana {
	return &grafana.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: grafana.GrafanaSpec{},
	}
}
