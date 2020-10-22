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

package infrastructure

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	strimziServerGroup = "kafka.strimzi.io"
	// StrimziOperatorName is the default Strimzi operator name
	StrimziOperatorName = "strimzi-cluster-operator"

	strimziBrokerLabel         = "strimzi.io/cluster"
	defaultKafkaTopicPartition = 1
	defaultKafkaTopicReplicas  = 1

	// KafkaKind refers to Kafka Kind as defined by Strimzi
	KafkaKind = "Kafka"

	// KafkaInstanceName is the default name for the Kafka cluster managed by KogitoInfra
	KafkaInstanceName = "kogito-kafka"
)

var (
	// KafkaAPIVersion refers to kafka APIVersion
	KafkaAPIVersion = v1beta1.SchemeGroupVersion.String()
)

// IsStrimziAvailable checks if Strimzi CRD is available in the cluster
func IsStrimziAvailable(client *client.Client) bool {
	return client.HasServerGroup(strimziServerGroup)
}

// GetKafkaDefaultResource returns a Kafka resource with default configuration
func GetKafkaDefaultResource(name, namespace string, defaultReplicas int32) *v1beta1.Kafka {
	return &v1beta1.Kafka{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.KafkaSpec{
			EntityOperator: v1beta1.EntityOperatorSpec{
				TopicOperator: v1beta1.EntityTopicOperatorSpec{},
				UserOperator:  v1beta1.EntityUserOperatorSpec{},
			},
			Kafka: v1beta1.KafkaClusterSpec{
				Replicas: defaultReplicas,
				Storage:  v1beta1.KafkaStorage{StorageType: v1beta1.KafkaEphemeralStorage},
				Listeners: v1beta1.KafkaListeners{
					Plain: v1beta1.KafkaListenerPlain{},
				},
				JvmOptions: map[string]interface{}{"gcLoggingEnabled": false},
				Config: map[string]interface{}{
					"log.message.format.version":               "2.3",
					"offsets.topic.replication.factor":         defaultReplicas,
					"transaction.state.log.min.isr":            1,
					"transaction.state.log.replication.factor": defaultReplicas,
					"auto.create.topics.enable":                true,
				},
			},
			Zookeeper: v1beta1.ZookeeperClusterSpec{
				Replicas: defaultReplicas,
				Storage:  v1beta1.KafkaStorage{StorageType: v1beta1.KafkaEphemeralStorage},
			},
		},
	}
}

// GetKafkaTopic returns a Kafka topic resource with default configuration
func GetKafkaTopic(name, namespace, kafkaBroker string) *v1beta1.KafkaTopic {

	labels := make(map[string]string)
	labels[strimziBrokerLabel] = kafkaBroker

	return &v1beta1.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1beta1.KafkaTopicSpec{
			Partitions: defaultKafkaTopicPartition,
			Replicas:   defaultKafkaTopicReplicas,
			TopicName:  name,
		},
	}
}

// ResolveKafkaServerURI returns the uri of the kafka instance
func ResolveKafkaServerURI(kafka *v1beta1.Kafka) (string, error) {
	log.Debugf("Resolving kafka URI for given kafka instance %s", kafka.Name)
	if len(kafka.Status.Listeners) > 0 {
		for _, listenerStatus := range kafka.Status.Listeners {
			if listenerStatus.Type == "plain" && len(listenerStatus.Addresses) > 0 {
				for _, listenerAddress := range listenerStatus.Addresses {
					if len(listenerAddress.Host) > 0 && listenerAddress.Port > 0 {
						kafkaURI := fmt.Sprintf("%s:%d", listenerAddress.Host, listenerAddress.Port)
						log.Debugf("Success fetch kafka URI for kafka instance(%s) : %s", kafka.Name, kafkaURI)
						return kafkaURI, nil
					}
				}
			}
		}
	}
	log.Debug("Not able resolve URI for given kafka instance")
	return "", fmt.Errorf("not able resolve URI for given kafka instance %s", kafka.Name)
}

// getKafkaInstanceWithName fetches the Kafka instance of the given name
func getKafkaInstanceWithName(name string, namespace string, client *client.Client) (*v1beta1.Kafka, error) {
	log.Debugf("Fetching kafka instance for given instance %s", name)
	kafka := &v1beta1.Kafka{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafka); err != nil {
		return nil, err
	} else if exists {
		log.Debugf("Successfully fetched kafka instance %s", name)
		return kafka, nil
	}
	log.Debugf("Kafka instance (%s) not found in namespace %s", name, namespace)
	return nil, nil
}
