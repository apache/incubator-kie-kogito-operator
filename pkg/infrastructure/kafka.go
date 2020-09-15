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
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	strimziServerGroup = "kafka.strimzi.io"
	// StrimziOperatorName is the default Strimzi operator name
	StrimziOperatorName = "strimzi-cluster-operator"
)

// IsStrimziOperatorAvailable verify if Strimzi Operator is running in the given namespace and the CRD is available
// Deprecated: rethink the way we check for the operator since the deployment resource could be in another namespace if installed cluster wide
func IsStrimziOperatorAvailable(cli *client.Client, namespace string) (available bool, err error) {
	log.Debugf("Checking if Strimzi Operator is available in the namespace %s", namespace)
	available = false
	if IsStrimziAvailable(cli) {
		log.Debugf("Strimzi CRDs available. Checking if Strimzi Operator is deployed in the namespace %s", namespace)
		list := &v1.DeploymentList{}
		if err = kubernetes.ResourceC(cli).ListWithNamespace(namespace, list); err != nil {
			return
		}
		for _, strimzi := range list.Items {
			for _, owner := range strimzi.OwnerReferences {
				if strings.HasPrefix(owner.Name, StrimziOperatorName) {
					available = true
					return
				}
			}
		}
	}
	return
}

// IsStrimziAvailable checks if Strimzi CRD is available in the cluster
func IsStrimziAvailable(client *client.Client) bool {
	return client.HasServerGroup(strimziServerGroup)
}

// GetKafkaServerURI provide kafka URI for given kafka instance name
func GetKafkaServerURI(kafkaInstanceName string, namespace string, client *client.Client) (string, error) {
	log.Debugf("Fetching Kafka server URI for instance %s in namespace %s", kafkaInstanceName, namespace)
	if kafkaInstance, err := getKafkaInstanceWithName(kafkaInstanceName, namespace, client); err != nil {
		return "", err
	} else if kafkaInstance == nil {
		return "", fmt.Errorf("kafka instance not found with name %s in namespace %s", kafkaInstanceName, namespace)
	} else {
		return resolveKafkaServerURI(kafkaInstance), nil
	}
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

// resolveKafkaServerURI returns the uri of the kafka instance
func resolveKafkaServerURI(kafka *v1beta1.Kafka) string {
	log.Debugf("Resolving kafka URI for given kafka instance %s", kafka.Name)
	log.Debugf("kafka instance : %s", kafka)
	log.Debugf("len(kafka.Status.Listeners) : %s", len(kafka.Status.Listeners))
	if len(kafka.Status.Listeners) > 0 {
		for _, listenerStatus := range kafka.Status.Listeners {
			log.Debugf("listenerStatus.Type : %s", listenerStatus.Type)
			log.Debugf("len(listenerStatus.Addresses) : %s", len(listenerStatus.Addresses))
			if listenerStatus.Type == "plain" && len(listenerStatus.Addresses) > 0 {
				for _, listenerAddress := range listenerStatus.Addresses {
					log.Debugf("listenerAddress.Host : %s, listenerAddress.Port : %s", listenerAddress.Host, listenerAddress.Port)
					if len(listenerAddress.Host) > 0 && listenerAddress.Port > 0 {
						kafkaURI := fmt.Sprintf("%s:%d", listenerAddress.Host, listenerAddress.Port)
						log.Debugf("Success fetch kafka URI for kafka instance(%s) : %s", kafka.Name, kafkaURI)
						return kafkaURI
					}
				}
			}
		}
	}
	log.Debug("Not able resolve URI for given kafka instance")
	return ""
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
