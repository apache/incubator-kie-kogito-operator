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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
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

func (k *kafkaMessagingDeployer) createRequiredResources(service v1alpha1.KogitoService) error {
	infra, err := k.fetchInfraDependency(service, infrastructure.IsKafkaResource)
	if err != nil || infra == nil {
		return err
	}
	if err := k.createRequiredKafkaTopics(infra, service); err != nil {
		return err
	}
	return nil
}

func (k *kafkaMessagingDeployer) createRequiredKafkaTopics(infra *v1alpha1.KogitoInfra, service v1alpha1.KogitoService) error {
	log.Debugf("Going to apply kafka topic configurations required by the deployed service '%s'", service.GetName())
	kafkaURI := infra.Status.AppProps[QuarkusKafkaBootstrapAppProp]
	if len(kafkaURI) == 0 {
		log.Debugf("Ignoring Kafka Topics creation, Kafka URI is empty from the given KogitoInfra: %s", infra.Name)
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
	topics, err := k.fetchTopicsAndSetCEStatus(service)
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

func (k *kafkaMessagingDeployer) createKafkaTopicIfNotExists(topicName string, instance *v1alpha1.KogitoInfra) error {
	log.Debugf("Going to create kafka topic it is not exists %k", topicName)

	kafkaNamespaceName := k.getKafkaInstanceNamespaceName(instance)
	log.Debugf("Resolved kafka instance name %k and namespace %k", kafkaNamespaceName.Name, kafkaNamespaceName.Namespace)

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

func (k *kafkaMessagingDeployer) getKafkaInstanceNamespaceName(instance *v1alpha1.KogitoInfra) *types.NamespacedName {
	// Step 1: check whether user has provided custom Kafka instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom kafka instance reference is provided")
		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debugf("Namespace is not provided for custom resource, taking instance namespace(%s) as default", namespace)
		}
		return &types.NamespacedName{Namespace: namespace, Name: instance.Spec.Resource.Name}
	}
	// create/refer kogito-kafka instance
	log.Debugf("Custom kafka instance reference is not provided")
	return &types.NamespacedName{Namespace: instance.Namespace, Name: infrastructure.KafkaInstanceName}
}

func (k *kafkaMessagingDeployer) loadDeployedKafkaTopic(topicName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debugf("Going to load deployed kafka topic %s", topicName)
	kafkaTopic := &kafkav1beta1.KafkaTopic{}
	if exits, err := kubernetes.ResourceC(k.cli).FetchWithKey(types.NamespacedName{Name: topicName, Namespace: kafkaNamespace}, kafkaTopic); err != nil {
		log.Errorf("Error occurs while fetching kogito kafka topic %s", topicName)
		return nil, err
	} else if exits {
		return kafkaTopic, nil
	}
	log.Debugf("kafka topic(%s) not exists", topicName)
	return nil, nil
}

func (k *kafkaMessagingDeployer) createNewKafkaTopic(topicName, kafkaName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debugf("Going to create kafka topic %s", topicName)
	kafkaTopic := infrastructure.GetKafkaTopic(topicName, kafkaNamespace, kafkaName)
	if err := kubernetes.ResourceC(k.cli).Create(kafkaTopic); err != nil {
		log.Error("Error occurs while creating kogito Kafka topic")
		return nil, err
	}
	log.Debugf("Kogito Kafka topic(%s) created successfully", topicName)
	return kafkaTopic, nil
}
