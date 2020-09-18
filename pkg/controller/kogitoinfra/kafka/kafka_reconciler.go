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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
)

// InfraResource implementation of KogitoInfraResource
type InfraResource struct {
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
func (k *InfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {

	var kafkaInstance *kafkav1beta1.Kafka

	// Step 1: check whether user has provided custom Kafka instance reference
	isCustomReferenceProvided := len(instance.Spec.Resource.Name) > 0
	if isCustomReferenceProvided {
		log.Debugf("Custom kafka instance reference is provided")

		kafkaInstance, resultErr = loadDeployedKafkaInstance(client, instance.Spec.Resource.Name, instance.Spec.Resource.Namespace)
		if resultErr != nil {
			return false, resultErr
		}
		if kafkaInstance == nil {
			return false, fmt.Errorf("custom kafka instance(%s) not found in namespace %s", instance.Spec.Resource.Name, instance.Spec.Resource.Namespace)
		}
	} else {
		// create/refer kogito-kafka instance
		log.Debugf("Custom kafka instance reference is not provided")

		// Verify kafka
		if !infrastructure.IsStrimziAvailable(client) {
			resultErr = fmt.Errorf("Kafka operator is not available in the namespace %s, impossible to continue ", instance.Namespace)
			return false, resultErr
		}

		// check whether kafka instance exist
		kafkaInstance, resultErr = loadDeployedKafkaInstance(client, InstanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if kafkaInstance == nil {
			// if not exist then create new Kafka instance. Strimzi operator creates Kafka instance, secret & service resource
			_, resultErr = createNewKafkaInstance(client, InstanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	updateAppPropsInStatus(kafkaInstance, instance)
	updateEnvVarsInStatus(kafkaInstance, instance)
	return false, nil
}
