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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	kafkaClusterLabel string = "strimzi.io/cluster"

	kafkaURINotFoundError string = "there's no Kafka instance URI found in the namespace and Kafka external URI is not specified, cannot deploy Data Index service"
	kafkaNotExist         string = "there is no Kafka instance found in the namespace, cannot deploy Kafka Topics"

	kafkaTopicConfigRetentionKey   string = "retention.ms"
	kafkaTopicConfigSegmentKey     string = "segment.bytes"
	kafkaTopicConfigRetentionValue string = "604800000"
	kafkaTopicConfigSegmentValue   string = "1073741824"

	quarkusTopicBootstrapEnvVar = "MP_MESSAGING_%s_%s_BOOTSTRAP_SERVERS"
	quarkusBootstrapEnvVar      = "QUARKUS_KAFKA_BOOTSTRAP_SERVERS"
)

// fromKafkaTopicToQuarkusEnvVar transforms a given Kafka Topic name into a environment variable to be read by Quarkus Kafka client used by Kogito Services
func fromKafkaTopicToQuarkusEnvVar(topic KafkaTopicDefinition) string {
	if &topic != nil && len(topic.TopicName) > 0 && len(topic.MessagingType) > 0 {
		return fmt.Sprintf(quarkusTopicBootstrapEnvVar, topic.MessagingType, strings.ToUpper(strings.ReplaceAll(topic.TopicName, "-", "_")))
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

	return "", fmt.Errorf(kafkaURINotFoundError)
}

func getKafkaServerReplicas(kafkaProp v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (string, int32, error) {
	if len(kafkaProp.ExternalURI) <= 0 {
		if kafka, err := getKafkaInstance(kafkaProp, namespace, client); err != nil {
			return "", 0, err
		} else if kafka != nil {
			if replicas := infrastructure.ResolveKafkaServerReplicas(kafka); replicas > 0 {
				return kafka.Name, replicas, nil
			}
		}

		return "", 0, fmt.Errorf(kafkaNotExist)
	}

	return "", 0, nil
}

func getKafkaInstance(kafka v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (*kafkabetav1.Kafka, error) {
	if len(kafka.Instance) > 0 {
		return infrastructure.GetKafkaInstanceWithName(kafka.Instance, namespace, client)
	}
	return nil, nil
}

func newKafkaTopic(topicName string, kafkaName string, kafkaReplicas int32, namespace string) *kafkabetav1.KafkaTopic {
	return &kafkabetav1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      topicName,
			Namespace: namespace,
			Labels: map[string]string{
				kafkaClusterLabel: kafkaName,
			},
		},
		Spec: kafkabetav1.KafkaTopicSpec{
			Replicas:   kafkaReplicas,
			Partitions: 10,
			Config: map[string]string{
				kafkaTopicConfigRetentionKey: kafkaTopicConfigRetentionValue,
				kafkaTopicConfigSegmentKey:   kafkaTopicConfigSegmentValue,
			},
		},
	}
}
