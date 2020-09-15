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
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/keycloak"
	corev1 "k8s.io/api/core/v1"
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
	case isKafkaResource(instance):
		log.Debugf("Kogito infra reference is for Kafka resource")
		return &kafka.InfraResource{}, nil
	case isKeycloakResource(instance):
		log.Debugf("Kogito infra reference is for Keycloak resource")
		return &keycloak.InfraResource{}, nil
	}
	return nil, fmt.Errorf("no Kogito infra resource found for given definition, %s", instance.Name)
}

// FetchKogitoInfraProperties provide infra properties
func FetchKogitoInfraProperties(client *client.Client, name string, namespace string, runtimeType v1alpha1.RuntimeType) (appProps map[string]string, envProps []corev1.EnvVar, err error) {
	log.Debug("Going to fetch kogito infra properties for reference %s in namespace %s", name, namespace)
	kogitoInstance, err := FetchKogitoInfraInstance(client, name, namespace)
	if err != nil {
		return
	}

	kogitoResource, err := GetKogitoInfraResource(kogitoInstance)
	if err != nil {
		return
	}

	appProps, envProps = kogitoResource.FetchInfraProperties(kogitoInstance, runtimeType)
	log.Debugf("Following infra properties are fetched : appProperties (%s) and envProp (%s)", appProps, envProps)
	return appProps, envProps, nil
}

func isInfinispanResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == infinispanv1.SchemeGroupVersion.String() && instance.Spec.Resource.Kind == "Infinispan"
}

func isKafkaResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == kafkabetav1.SchemeGroupVersion.String() && instance.Spec.Resource.Kind == "Kafka"
}

func isKeycloakResource(instance *v1alpha1.KogitoInfra) bool {
	return instance.Spec.Resource.APIVersion == keycloakv1alpha1.SchemeGroupVersion.String() && instance.Spec.Resource.Kind == "Keycloak"
}
