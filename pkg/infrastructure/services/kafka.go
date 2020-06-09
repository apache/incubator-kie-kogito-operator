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
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
)

const (
	quarkusTopicBootstrapAppProp = "mp.messaging.%s.%s.bootstrap.servers"

	// QuarkusBootstrapAppProp quarkus application property for setting kafka server
	QuarkusBootstrapAppProp = "kafka.bootstrap.servers"

	// SpringBootstrapAppProp spring boot application property for setting kafka server
	SpringBootstrapAppProp = "kafka.bootstrap-servers"

	// Deprecated, keep it for Job Service scripts
	quarkusBootstrapEnvVar = "KAFKA_BOOTSTRAP_SERVERS"
)

// fromKafkaTopicToQuarkusAppProp transforms a given Kafka Topic name into a application properties to be read by Quarkus Kafka client used by Kogito Services
func fromKafkaTopicToQuarkusAppProp(topic KafkaTopicDefinition) string {
	if len(topic.TopicName) > 0 && len(topic.MessagingType) > 0 {
		return fmt.Sprintf(quarkusTopicBootstrapAppProp, topic.MessagingType, topic.TopicName)
	}
	return ""
}

func getKafkaServerURI(kafkaProp v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (string, error) {
	if len(kafkaProp.ExternalURI) > 0 {
		return kafkaProp.ExternalURI, nil
	}
	if kafka, err := getKafkaInstance(kafkaProp, namespace, client); err != nil {
		return "", err
	} else if kafka != nil {
		if uri := infrastructure.ResolveKafkaServerURI(kafka); len(uri) > 0 {
			return uri, nil
		}
	}

	return "", nil
}

func getKafkaInstance(kafka v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (*kafkabetav1.Kafka, error) {
	if len(kafka.Instance) > 0 {
		return infrastructure.GetKafkaInstanceWithName(kafka.Instance, namespace, client)
	}
	return nil, nil
}
