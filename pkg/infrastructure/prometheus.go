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

	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
)

// DeployPrometheusWithKogitoInfra deploys KogitoInfra with Kafka enabled
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func DeployPrometheusWithKogitoInfra(instance v1alpha1.PrometheusAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// Overrides any parameters not set
	if instance.GetPrometheusProperties().UseKogitoInfra {
		// ensure infra
		infra, ready, err := EnsureKogitoInfra(namespace, cli).WithPrometheus().Apply()
		if err != nil {
			return false, 0, err
		}

		log.Debugf("Checking KogitoInfra status to make sure we are ready to use Prometheus. Status are: %s", infra.Status.Prometheus)
		if ready {
			log.Debug("KogitoInfra Prometheus is ready.")
			return true, 0, nil
		}
		log.Debug("KogitoInfra Prometheus is not ready, requeue")
		// waiting for infinispan deployment
		return false, time.Second * 10, nil
	}

	return false, 0, nil
}

// IsPrometheusAvailable checks if Strimzi CRD is available in the cluster
func IsPrometheusAvailable(client *client.Client) bool {
	return client.HasServerGroup(prometheusv1.SchemeGroupVersion.Group)
}
