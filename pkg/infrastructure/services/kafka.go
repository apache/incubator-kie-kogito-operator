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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
)

const (
	quarkusTopicBootstrapAppProp = "mp.messaging.%s.%s.bootstrap.servers"
)

func (s *serviceDeployer) applyKafkaConfigurations(appProps map[string]string) {
	log.Debugf("Going to apply kafka topic configurations")
	kafkaURI := getKafkaServerURIFromAppProps(appProps)
	if len(kafkaURI) > 0 {
		if s.instance.GetSpec().GetRuntime() == v1alpha1.QuarkusRuntimeType {
			for _, kafkaTopic := range s.definition.KafkaTopics {
				appProps[fromKafkaTopicToQuarkusAppProp(kafkaTopic)] = kafkaURI
			}
		}
	}
	log.Debugf("Skipping to apply kafka topics configuration as kafka URI is not received in Infra app props")
}

// fromKafkaTopicToQuarkusAppProp transforms a given Kafka Topic name into a application properties to be read by Quarkus Kafka client used by Kogito Services
func fromKafkaTopicToQuarkusAppProp(topic KafkaTopicDefinition) string {
	if len(topic.TopicName) > 0 && len(topic.MessagingType) > 0 {
		return fmt.Sprintf(quarkusTopicBootstrapAppProp, topic.MessagingType, topic.TopicName)
	}
	return ""
}

func getKafkaServerURIFromAppProps(appProps map[string]string) string {
	return appProps[kafka.QuarkusKafkaBootstrapAppProp]
}
