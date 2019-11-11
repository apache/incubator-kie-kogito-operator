// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package resource

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	kafkaClusterLabel string = "strimzi.io/cluster"

	kafkaURINotFoundError string = "there's no Kafka instance URI found in the namespace and Kafka external URI is not specified, cannot deploy Data Index service"
	kafkaNotExist         string = "there is no Kafka instance found in the namespace, cannot deploy Kafka Topics"

	kafkaTopicConfigRetentionKey   string = "retention.ms"
	kafkaTopicConfigSegmentKey     string = "segment.bytes"
	kafkaTopicConfigRetentionValue string = "604800000"
	kafkaTopicConfigSegmentValue   string = "1073741824"
)

func fromKafkaToStringMap(externalURI string) map[string]string {
	propsmap := map[string]string{}
	if len(externalURI) > 0 {
		for _, envKey := range managedKafkaKeys {
			propsmap[envKey] = externalURI
		}
	}
	return propsmap
}

// IsKafkaServerURIResolved checks if the URI of the Kafka server is provided or resolvable in the namespace
func IsKafkaServerURIResolved(instance *v1alpha1.KogitoDataIndex, client *client.Client) (bool, error) {
	if len(instance.Spec.Kafka.ExternalURI) == 0 {
		if !isStrimziAvailable(client) {
			return false, nil
		}
		if kafka, err := getKafkaInstance(instance.Spec.Kafka, instance.Namespace, client); err != nil {
			return false, err
		} else if kafka == nil {
			return false, nil
		}
	}
	return true, nil
}

func getKafkaServerURI(kafkaProp v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (string, error) {
	if len(kafkaProp.ExternalURI) > 0 {
		return kafkaProp.ExternalURI, nil
	} else if isStrimziAvailable(client) {
		if kafka, err := getKafkaInstance(kafkaProp, namespace, client); err != nil {
			return "", err
		} else if kafka != nil {
			if uri := resolveKafkaServerURI(kafka); len(uri) > 0 {
				return uri, nil
			}
		}
	}
	return "", fmt.Errorf(kafkaURINotFoundError)
}

func getKafkaServerReplicas(kafkaProp v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (string, int32, error) {
	if len(kafkaProp.ExternalURI) <= 0 && isStrimziAvailable(client) {
		if kafka, err := getKafkaInstance(kafkaProp, namespace, client); err != nil {
			return "", 0, err
		} else if kafka != nil {
			if replicas := resolveKafkaServerReplicas(kafka); replicas > 0 {
				return kafka.Name, replicas, nil
			}
		}

		return "", 0, fmt.Errorf(kafkaNotExist)
	}

	return "", 0, nil
}

func resolveKafkaServerURI(kafka *kafkabetav1.Kafka) string {
	if kafka != nil {
		if len(kafka.Status.Listeners) > 0 {
			for _, listenerStatus := range kafka.Status.Listeners {
				if listenerStatus.Type == "plain" && len(listenerStatus.Addresses) > 0 {
					for _, listenerAddress := range listenerStatus.Addresses {
						if len(listenerAddress.Host) > 0 && listenerAddress.Port > 0 {
							return fmt.Sprintf("%s:%d", listenerAddress.Host, listenerAddress.Port)
						}
					}
				}
			}
		}
	}
	return ""
}

func resolveKafkaServerReplicas(kafka *kafkabetav1.Kafka) int32 {
	if kafka != nil {
		if kafka.Spec.KafkaClusterSpec.Replicas > 0 {
			return kafka.Spec.KafkaClusterSpec.Replicas
		}
		return 1
	}
	return 0
}

func getKafkaInstance(kafka v1alpha1.KafkaConnectionProperties, namespace string, client *client.Client) (*kafkabetav1.Kafka, error) {
	if len(kafka.Instance) > 0 {
		return getKafkaInstanceWithName(kafka.Instance, namespace, client)
	}
	return getKafkaInstanceInNamespace(namespace, client)
}

func getKafkaInstanceWithName(name string, namespace string, client *client.Client) (*kafkabetav1.Kafka, error) {
	kafka := &kafkabetav1.Kafka{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafka); err != nil {
		return nil, err
	} else if exists {
		return kafka, nil
	}
	return nil, nil
}

func getKafkaInstanceInNamespace(namespace string, client *client.Client) (*kafkabetav1.Kafka, error) {
	kafkaList := &kafkabetav1.KafkaList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, kafkaList); err != nil {
		return nil, err
	} else if len(kafkaList.Items) == 1 {
		return &kafkaList.Items[0], nil
	}
	return nil, nil
}

func newKafkaTopic(topicName string, kafkaName string, kafkaReplicas int32, namespace string) *kafkabetav1.KafkaTopic {
	return &kafkabetav1.KafkaTopic{
		ObjectMeta: v1.ObjectMeta{
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

func isStrimziAvailable(client *client.Client) bool {
	return client.HasServerGroup("kafka.strimzi.io")
}
