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

package kogitoinfra

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/keycloak"
	"k8s.io/apimachinery/pkg/types"
)

// FetchKogitoInfraInstance load infra instance
func FetchKogitoInfraInstance(client *client.Client, name string, namespace string) (*v1alpha1.KogitoInfra, error) {
	log.Debugf("going to fetch deployed kogito infra instance %s", name)
	instance := &v1alpha1.KogitoInfra{}
	if exists, resultErr := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, instance); resultErr != nil {
		log.Errorf("Error occurs while fetching deployed kogito infra instance %s", name)
		return nil, resultErr
	} else if !exists {
		return nil, fmt.Errorf("kogito Infra resource with name %s not found in namespace %s", name, namespace)
	} else {
		log.Debugf("Successfully fetch deployed kogito infra reference %s", name)
		return instance, nil
	}
}

// GetKogitoInfraResource identify and return request kogito infra resource on bases of information provided in kogitoInfra value
func GetKogitoInfraResource(instance *v1alpha1.KogitoInfra) (InfraResource, error) {
	log.Debugf("going to fetch related kogito infra resource for given infra instance : %s", instance.Name)
	switch {
	case isInfinispanResource(instance):
		log.Debugf("Kogito infra reference is for Infinispan resource")
		return &infinispan.InfraResource{}, nil
	case IsKafkaResource(instance):
		log.Debugf("Kogito infra reference is for Kafka resource")
		return &kafka.InfraResource{}, nil
	case isKeycloakResource(instance):
		log.Debugf("Kogito infra reference is for Keycloak resource")
		return &keycloak.InfraResource{}, nil
	}
	return nil, fmt.Errorf("no Kogito infra resource found for given definition, %s", instance.Name)
}

// isInfinispanResource check if provided KogitoInfra instance is for Infinispan resource
func isInfinispanResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == infinispan.APIVersion && instance.Spec.Resource.Kind == infinispan.Kind
}

// IsKafkaResource check if provided KogitoInfra instance is for kafka resource
func IsKafkaResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == kafka.APIVersion && instance.Spec.Resource.Kind == kafka.Kind
}

// isKeycloakResource check if provided KogitoInfra instance is for Keycloak resource
func isKeycloakResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == keycloak.APIVersion && instance.Spec.Resource.Kind == keycloak.Kind
}
