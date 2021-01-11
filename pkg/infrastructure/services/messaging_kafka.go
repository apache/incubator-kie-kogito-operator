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

package services

import (
	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// QuarkusKafkaBootstrapAppProp quarkus application property for setting kafka server
	QuarkusKafkaBootstrapAppProp = "kafka.bootstrap.servers"
)

// kafkaMessagingDeployer implementation of messagingHandler
type kafkaMessagingDeployer struct {
	messagingDeployer
}

func (k *kafkaMessagingDeployer) createRequiredResources(service v1beta1.KogitoService) error {
	infra, err := k.fetchInfraDependency(service, infrastructure.IsKafkaResource)
	if err != nil || infra == nil {
		return err
	}
	if err := k.createRequiredKafkaTopics(infra, service); err != nil {
		return err
	}
	return nil
}

func (k *kafkaMessagingDeployer) createRequiredKafkaTopics(infra *v1beta1.KogitoInfra, service v1beta1.KogitoService) error {
	log.Debug("Going to apply kafka topic configurations required by the deployed service", "KogitoService", service.GetName())
	kafkaURI := infra.Status.RuntimeProperties[v1beta1.QuarkusRuntimeType].AppProps[QuarkusKafkaBootstrapAppProp]
	if len(kafkaURI) == 0 {
		log.Debug("Ignoring Kafka Topics creation, Kafka URI is empty from the given KogitoInfra", "KogitoInfra", infra.Name)
		return nil
	}
	// topics required by definition
	for _, kafkaTopic := range k.definition.KafkaTopics {
		err := k.createKafkaTopicIfNotExists(kafkaTopic, infra)
		if err != nil {
			return err
		}
	}
	// topics required by the deployed service
	topics, err := k.fetchTopicsAndSetCloudEventsStatus(service)
	if err != nil {
		return err
	}
	for _, topic := range topics {
		err := k.createKafkaTopicIfNotExists(topic.Name, infra)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *kafkaMessagingDeployer) createKafkaTopicIfNotExists(topicName string, instance *v1beta1.KogitoInfra) error {
	log.Debug("Going to create kafka topic it is not exists", "topicName", topicName)

	kafkaNamespaceName := k.getKafkaInstanceNamespaceName(instance)
	log.Debug("Resolved kafka instance", "name", kafkaNamespaceName.Name, "namespace", kafkaNamespaceName.Namespace)

	kafkaTopic, err := k.loadDeployedKafkaTopic(topicName, kafkaNamespaceName.Namespace)
	if err != nil {
		return err
	}

	if kafkaTopic == nil {
		_, err := k.createNewKafkaTopic(topicName, kafkaNamespaceName.Name, kafkaNamespaceName.Namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *kafkaMessagingDeployer) getKafkaInstanceNamespaceName(instance *v1beta1.KogitoInfra) *types.NamespacedName {
	// Step 1: check whether user has provided custom Kafka instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debug("Custom kafka instance reference is provided")
		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debug("Namespace is not provided for custom resource, taking instance namespace as default", "instance namespace", namespace)
		}
		return &types.NamespacedName{Namespace: namespace, Name: instance.Spec.Resource.Name}
	}
	// create/refer kogito-kafka instance
	log.Debug("Custom kafka instance reference is not provided")
	return &types.NamespacedName{Namespace: instance.Namespace, Name: infrastructure.KafkaInstanceName}
}

func (k *kafkaMessagingDeployer) loadDeployedKafkaTopic(topicName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debug("Going to load deployed kafka topic", "topicName", topicName)
	kafkaTopic := &kafkav1beta1.KafkaTopic{}
	if exits, err := kubernetes.ResourceC(k.cli).FetchWithKey(types.NamespacedName{Name: topicName, Namespace: kafkaNamespace}, kafkaTopic); err != nil {
		log.Error(err, "Error occurs while fetching kogito kafka topic", "topicName", topicName)
		return nil, err
	} else if exits {
		return kafkaTopic, nil
	}
	log.Debug("kafka topic not exists", "topicName", topicName)
	return nil, nil
}

func (k *kafkaMessagingDeployer) createNewKafkaTopic(topicName, kafkaName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debug("Going to create kafka topic", "topicName", topicName)
	kafkaTopic := infrastructure.GetKafkaTopic(topicName, kafkaNamespace, kafkaName)
	if err := kubernetes.ResourceC(k.cli).Create(kafkaTopic); err != nil {
		log.Error(err, "Error occurs while creating kogito Kafka topic")
		return nil, err
	}
	log.Debug("Kogito Kafka topic created successfully", "topicName", topicName)
	return kafkaTopic, nil
}
