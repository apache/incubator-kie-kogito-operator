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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"time"
)

// DeployKafkaWithKogitoInfra deploys KogitoInfra with Kafka enabled
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func DeployKafkaWithKogitoInfra(instance v1alpha1.KafkaAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// Overrides any parameters not set
	if instance.GetKafkaProperties().UseKogitoInfra {
		// ensure infra
		infra, ready, err := EnsureKogitoInfra(namespace, cli).WithKafka().Apply()
		if err != nil {
			return false, 0, err
		}

		log.Debugf("Checking KogitoInfra status to make sure we are ready to use Kafka. Status are: %s", infra.Status.Kafka)
		if ready {
			kafka, err := GetReadyKafkaInstanceName(cli, infra)
			if err != nil {
				return false, 0, err
			}
			if len(kafka) > 0 {
				if instance.GetKafkaProperties().Instance == kafka {
					return false, 0, nil
				}

				log.Debug("Looks ok, we are ready to use Kafka!")
				instance.SetKafkaProperties(v1alpha1.KafkaConnectionProperties{
					Instance: kafka,
				})

				return true, 0, nil
			}
		}
		log.Debug("KogitoInfra is not ready, requeue")
		// waiting for infinispan deployment
		return false, time.Second * 10, nil
	}

	// Ensure default values
	if instance.AreKafkaPropertiesBlank() {
		instance.SetKafkaProperties(v1alpha1.KafkaConnectionProperties{
			UseKogitoInfra: true,
		})
		return true, 0, nil
	}

	return false, 0, nil
}
