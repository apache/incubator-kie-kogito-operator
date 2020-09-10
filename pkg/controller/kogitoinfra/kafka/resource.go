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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"reflect"
)

const (
	// InstanceName is the default name for the Kafka cluster managed by KogitoInfra
	InstanceName    = "kogito-kafka"
	defaultReplicas = 1
)

var log = logger.GetLogger("kogitokafka_resource")

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
		kafka := infrastructure.GetKafkaDefaultResource(InstanceName, kogitoInfra.Namespace, defaultReplicas)
		resources[reflect.TypeOf(kafkabetav1.Kafka{})] = []resource.KubernetesResource{kafka}
		log.Debugf("Requested objects created as %s", resources)
	}

	return
}

func loadDeployedKafkaInstance(cli *client.Client, name string, namespace string) (*kafkabetav1.Kafka, error) {
	log.Debug("fetching deployed kogito kafka instance")
	kafkaInstance := &kafkabetav1.Kafka{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafkaInstance); err != nil {
		log.Error("Error occurs while fetching kogito kafka instance")
		return nil, err
	} else if !exits {
		log.Debug("Kogito kafka instance is not exists")
		return nil, nil
	} else {
		log.Debug("Kogito kafka instance found")
		return kafkaInstance, nil
	}
}

func createNewKafkaInstance(cli *client.Client, name, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*kafkabetav1.Kafka, error) {
	log.Debug("Going to create kogito Kafka instance")
	kafkaInstance := infrastructure.GetKafkaDefaultResource(name, namespace, defaultReplicas)
	if err := controllerutil.SetOwnerReference(instance, kafkaInstance, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(kafkaInstance); err != nil {
		log.Error("Error occurs while creating kogito Kafka instance")
		return nil, err
	}
	log.Debug("Kogito Kafka instance created successfully")
	return kafkaInstance, nil
}
