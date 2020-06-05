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

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	strimziServerGroup = "kafka.strimzi.io"
	// StrimziOperatorName is the default Strimzi operator name
	StrimziOperatorName = "strimzi-cluster-operator"
	// defaultKafkaPort default port for plain connections
	defaultKafkaPort = 9092
)

// 9093 -> secure, 9091 -> fallback
var kafkaFallbackPorts = []int32{9093, 9091}

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

// GetKafkaServiceURI fetches for the Kafka service linked with the given Kogito
// Infra and returns a formatted URI
func GetKafkaServiceURI(cli *client.Client, infra *v1alpha1.KogitoInfra) (uri string, err error) {
	uri = ""
	err = nil
	if infra == nil || len(infra.Status.Kafka.Service) == 0 {
		return
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: infra.Status.Kafka.Service, Namespace: infra.Namespace},
	}
	exists := false
	if exists, err = kubernetes.ResourceC(cli).Fetch(service); err != nil {
		return
	}

	if exists {
		// prefer default port
		for _, port := range service.Spec.Ports {
			if port.TargetPort.IntVal == defaultKafkaPort {
				return fmt.Sprintf("%s:%d", service.Name, port.TargetPort.IntVal), nil
			}
		}
		log.Warnf("Kafka default port (%d) not found in service %s. Trying %s", defaultKafkaPort, service.Name, kafkaFallbackPorts)
		// lets try others
		for _, port := range service.Spec.Ports {
			for _, kafkaPort := range kafkaFallbackPorts {
				if port.TargetPort.IntVal == kafkaPort {
					return fmt.Sprintf("%s:%d", service.Name, port.TargetPort.IntVal), nil
				}
			}
		}
		return "", fmt.Errorf("Kafka port (%d) not found in service %s ", defaultKafkaPort, service.Name)
	}
	return
}

// GetReadyKafkaInstanceName fetches for the Kafka Instance linked with the given Kogito and checks if it is ready
func GetReadyKafkaInstanceName(cli *client.Client, infra *v1alpha1.KogitoInfra) (kafka string, err error) {
	kafka = ""
	err = nil
	if infra == nil || len(infra.Status.Kafka.Name) == 0 {
		return
	}

	instance, err := GetKafkaInstanceWithName(infra.Status.Kafka.Name, infra.Namespace, cli)
	if err != nil {
		return
	}
	if instance != nil {
		uri := ResolveKafkaServerURI(instance)
		if len(uri) > 0 {
			kafka = instance.Name
		}
	}
	return
}

// ResolveKafkaServerURI returns the uri of the kafka instance
func ResolveKafkaServerURI(kafka *v1beta1.Kafka) string {
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

// ResolveKafkaServerReplicas returns the number of replicas of the kafka instance
func ResolveKafkaServerReplicas(kafka *v1beta1.Kafka) int32 {
	if kafka != nil {
		if kafka.Spec.Kafka.Replicas > 0 {
			return kafka.Spec.Kafka.Replicas
		}
		return 1
	}
	return 0
}

// GetKafkaInstanceWithName fetches the Kafka instance of the given name
func GetKafkaInstanceWithName(name string, namespace string, client *client.Client) (*v1beta1.Kafka, error) {
	kafka := &v1beta1.Kafka{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafka); err != nil {
		return nil, err
	} else if exists {
		return kafka, nil
	}
	return nil, nil
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
