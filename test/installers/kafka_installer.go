// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package installers

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	kafkabetav1 "github.com/kiegroup/kogito-operator/core/infrastructure/kafka/v1beta1"
	"github.com/kiegroup/kogito-operator/test/framework"
)

var (
	// kafkaOlmClusterWideInstaller installs Kafka cluster wide using OLM
	kafkaOlmClusterWideInstaller = OlmClusterWideServiceInstaller{
		SubscriptionName:                   kafkaOperatorSubscriptionName,
		Channel:                            kafkaOperatorSubscriptionChannel,
		Catalog:                            framework.CommunityCatalog,
		InstallationTimeoutInMinutes:       kafkaOperatorTimeoutInMin,
		GetAllClusterWideOlmCrsInNamespace: getKafkaCrsInNamespace,
	}

	kafkaOperatorSubscriptionName    = "strimzi-kafka-operator"
	kafkaOperatorSubscriptionChannel = "strimzi-0.22.x"
	kafkaOperatorTimeoutInMin        = 10
)

// GetKafkaInstaller returns Kafka installer
func GetKafkaInstaller() ServiceInstaller {
	return &kafkaOlmClusterWideInstaller
}

func getKafkaCrsInNamespace(namespace string) ([]kubernetes.ResourceObject, error) {
	crs := []kubernetes.ResourceObject{}

	kafkas := &kafkabetav1.KafkaList{}
	if err := framework.GetObjectsInNamespace(namespace, kafkas); err != nil {
		return nil, err
	}
	for i := range kafkas.Items {
		crs = append(crs, &kafkas.Items[i])
	}

	kafkaTopics := &kafkabetav1.KafkaTopicList{}
	if err := framework.GetObjectsInNamespace(namespace, kafkaTopics); err != nil {
		return nil, err
	}
	for i := range kafkaTopics.Items {
		crs = append(crs, &kafkaTopics.Items[i])
	}

	return crs, nil
}
