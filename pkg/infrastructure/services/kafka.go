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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/types"
)

func getKafkaServerURIFromAppProps(appProps map[string]string) string {
	return appProps[kafka.QuarkusKafkaBootstrapAppProp]
}

func (s *serviceDeployer) createKafkaTopics(kogitoInfraInstance *v1alpha1.KogitoInfra, kafkaURI string) error {
	log.Debugf("Going to apply kafka topic configurations")
	if len(kafkaURI) > 0 {
		for _, kafkaTopic := range s.definition.KafkaTopics {
			err := s.createKafkaTopicIfNotExists(kafkaTopic, kogitoInfraInstance)
			if err != nil {
				return err
			}
		}
	}
	log.Debugf("Skipping to apply kafka topics configuration as kafka URI is not received in Infra app props")
	return nil
}

func (s *serviceDeployer) createKafkaTopicIfNotExists(topicName string, instance *v1alpha1.KogitoInfra) error {
	log.Debugf("Going to create kafka topic it is not exists %s", topicName)

	kafkaNamespaceName := getKafkaInstanceNamespaceName(instance)
	log.Debugf("Resolved kafka instance name %s and namespace %s", kafkaNamespaceName.Name, kafkaNamespaceName.Namespace)

	kafkaTopic, err := loadDeployedKafkaTopic(s.client, topicName, kafkaNamespaceName.Namespace)
	if err != nil {
		return err
	}

	if kafkaTopic == nil {
		_, err := createNewKafkaTopic(s.client, topicName, kafkaNamespaceName.Name, kafkaNamespaceName.Namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func getKafkaInstanceNamespaceName(instance *v1alpha1.KogitoInfra) *types.NamespacedName {
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

func loadDeployedKafkaTopic(cli *client.Client, topicName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debugf("Going to load deployed kafka topic %s", topicName)
	kafkaTopic := &kafkav1beta1.KafkaTopic{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: topicName, Namespace: kafkaNamespace}, kafkaTopic); err != nil {
		log.Errorf("Error occurs while fetching kogito kafka topic %s", topicName)
		return nil, err
	} else if exits {
		return kafkaTopic, nil
	}
	log.Debugf("kafka topic(%s) not exists", topicName)
	return nil, nil
}

func createNewKafkaTopic(cli *client.Client, topicName, kafkaName, kafkaNamespace string) (*kafkav1beta1.KafkaTopic, error) {
	log.Debugf("Going to create kafka topic %s", topicName)
	kafkaTopic := infrastructure.GetKafkaTopic(topicName, kafkaNamespace, kafkaName)
	if err := kubernetes.ResourceC(cli).Create(kafkaTopic); err != nil {
		log.Error("Error occurs while creating kogito Kafka topic")
		return nil, err
	}
	log.Debugf("Kogito Kafka topic(%s) created successfully", topicName)
	return kafkaTopic, nil
}
