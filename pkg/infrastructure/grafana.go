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
	"time"

	grafanav1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
)

// DeployGrafanaWithKogitoInfra deploys KogitoInfra with Kafka enabled
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func DeployGrafanaWithKogitoInfra(instance v1alpha1.GrafanaAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// Overrides any parameters not set
	if instance.GetGrafanaProperties().UseKogitoInfra {
		// ensure infra
		infra, ready, err := EnsureKogitoInfra(namespace, cli).WithGrafana().Apply()
		if err != nil {
			return false, 0, err
		}

		log.Debugf("Checking KogitoInfra status to make sure we are ready to use Grafana. Status are: %s", infra.Status.Grafana)
		if ready {
			log.Debug("KogitoInfra Grafana is ready.")
			return true, 0, nil
		}
		log.Debug("KogitoInfra Grafana is not ready, requeue")
		// waiting for infinispan deployment
		return false, time.Second * 10, nil
	}

	return false, 0, nil
}

// IsGrafanaAvailable checks if Strimzi CRD is available in the cluster
func IsGrafanaAvailable(client *client.Client) bool {
	return client.HasServerGroup(grafanav1.SchemeGroupVersion.Group)
}
