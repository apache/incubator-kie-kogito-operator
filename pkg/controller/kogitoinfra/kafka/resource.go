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

package kafka

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"reflect"
)

var log = logger.GetLogger("kogitoinfra_resource")

const (
	// InstanceName is the default name for the Kafka cluster managed by KogitoInfra
	InstanceName    = "kogito-kafka"
	defaultReplicas = 1
)

// GetDeployedResources gets the resources deployed as is
func GetDeployedResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if infrastructure.IsStrimziAvailable(cli) {
		reader := read.New(cli.ControlCli).WithNamespace(kogitoInfra.Namespace).WithOwnerObject(kogitoInfra)
		resources, err = reader.ListAll(&kafkabetav1.KafkaList{})
		if err != nil {
			log.Warn("Failed to list deployed objects. ", err)
			return nil, err
		}
	}

	return
}

// CreateRequiredResources creates the very basic resources to have Kafka in the namespace
func CreateRequiredResources(kogitoInfra *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource, 1)
	if kogitoInfra.Spec.InstallKafka {
		log.Debugf("Creating default resources for Kafka installation for Kogito Infra on %s namespace", kogitoInfra.Namespace)
		kafka := &kafkabetav1.Kafka{
			ObjectMeta: metav1.ObjectMeta{
				Name:      InstanceName,
				Namespace: kogitoInfra.Namespace,
			},
			Spec: kafkabetav1.KafkaSpec{
				EntityOperator: kafkabetav1.EntityOperatorSpec{
					TopicOperator: kafkabetav1.EntityTopicOperatorSpec{},
					UserOperator:  kafkabetav1.EntityUserOperatorSpec{},
				},
				Kafka: kafkabetav1.KafkaClusterSpec{
					Replicas: defaultReplicas,
					Storage:  kafkabetav1.KafkaStorage{StorageType: kafkabetav1.KafkaEphemeralStorage},
					Listeners: kafkabetav1.KafkaListeners{
						Plain: kafkabetav1.KafkaListenerPlain{},
					},
					JvmOptions: map[string]interface{}{"gcLoggingEnabled": false},
					Config: map[string]interface{}{
						"log.message.format.version":               "2.3",
						"offsets.topic.replication.factor":         defaultReplicas,
						"transaction.state.log.min.isr":            2,
						"transaction.state.log.replication.factor": defaultReplicas,
						"auto.create.topics.enable":                true,
					},
				},
				Zookeeper: kafkabetav1.ZookeeperClusterSpec{
					Replicas: defaultReplicas,
					Storage:  kafkabetav1.KafkaStorage{StorageType: kafkabetav1.KafkaEphemeralStorage},
				},
			},
		}
		resources[reflect.TypeOf(kafkabetav1.Kafka{})] = []resource.KubernetesResource{kafka}
		log.Debugf("Requested objects created as %s", resources)
	}

	return
}
