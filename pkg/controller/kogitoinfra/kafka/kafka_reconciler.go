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

package kafka

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
)

// KafkaResource implementation of KogitoInfraResource
type KafkaResource struct {
}

// GetWatchedObjects provide list of object that needs to be watched to maintain Kafka kogitoInfra resource
func GetWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: kafkav1beta1.SchemeGroupVersion,
			AddToScheme:  kafkav1beta1.SchemeBuilder.AddToScheme,
			Objects:      []runtime.Object{&kafkav1beta1.Kafka{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (k *KafkaResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	var kafkaInstance *kafkav1beta1.Kafka

	// Step 1: check whether user has provided custom infinispan instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom kafka instance reference is provided")
		resourceName := instance.Spec.Resource.Name
		resourceNameSpace := instance.Spec.Resource.Namespace

		kafkaInstance, resultErr = loadDeployedKafkaInstance(client, resourceName, resourceNameSpace)
		if resultErr != nil {
			return false, resultErr
		}
		if kafkaInstance == nil {
			return false, fmt.Errorf("custom kafka instance(%s) not found in namespace %s", resourceName, resourceNameSpace)
		}
	} else {
		// create/refer kogito-kafka instance
		log.Debugf("Custom kafka instance reference is not provided")
		// if resource name is not provided then Infinispan instance should be created with default name = kogito-infinispan
		resourceName := InstanceName

		// if resource name is not provided then Infinispan instance should be created with default name = kogito-infinispan
		resourceNameSpace := instance.Namespace

		// Verify kafka
		if !infrastructure.IsStrimziAvailable(client) {
			resultErr = fmt.Errorf("Kafka operator is not available in the namespace %s, impossible to continue ", resourceNameSpace)
			return false, resultErr
		}

		// check whether kafka instance exist
		kafkaInstance, resultErr := loadDeployedKafkaInstance(client, resourceName, resourceNameSpace)
		if resultErr != nil {
			return false, resultErr
		}

		if kafkaInstance == nil {
			// if not exist then create new Infinispan instance. Infinispan operator creates Infinispan instance, secret & service resource
			kafkaInstance, resultErr = createNewKafkaInstance(client, resourceName, resourceNameSpace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}

	uri, resultErr := infrastructure.GetKafkaServerURI(kafkaInstance.Name, kafkaInstance.Namespace, client)
	if resultErr != nil {
		return false, resultErr
	}

	kafkaProperties := &instance.Status.Kafka.KafkaProperties
	kafkaProperties.ExternalURI = uri

	log.Debugf("Updating kogitoInfra(%s) value with new properties", instance.Name)
	if resultErr = kubernetes.ResourceC(client).Update(instance); resultErr != nil {
		log.Errorf("Error occurs while update kogitoInfra values")
		return false, resultErr
	}

	return false, nil
}
