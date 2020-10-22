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

	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
)

// getKogitoInfraResource identify and return request kogito infra resource on bases of information provided in kogitoInfra value
func getKogitoInfraResource(instance *v1alpha1.KogitoInfra) (InfraResource, error) {
	log.Debugf("going to fetch related kogito infra resource for given infra instance : %s", instance.Name)
	if infraRes, ok := getSupportedInfraResources()[resourceClassForInstance(instance)]; ok {
		return infraRes, nil
	}
	return nil, newUnsupportedAPIError(instance)
}

func resourceClassForInstance(instance *v1alpha1.KogitoInfra) string {
	return getResourceClass(instance.Spec.Resource.Kind, instance.Spec.Resource.APIVersion)
}

func getResourceClass(kind, APIVersion string) string {
	return strings.ToLower(fmt.Sprintf("%s.%s", kind, APIVersion))
}

func getSupportedInfraResources() map[string]InfraResource {
	return map[string]InfraResource{
		getResourceClass(infrastructure.InfinispanKind, infrastructure.InfinispanAPIVersion):                 &infinispanInfraResource{},
		getResourceClass(infrastructure.KafkaKind, infrastructure.KafkaAPIVersion):                           &kafkaInfraResource{},
		getResourceClass(infrastructure.KeycloakKind, infrastructure.KeycloakAPIVersion):                     &keycloakInfraResource{},
		getResourceClass(infrastructure.KnativeEventingBrokerKind, infrastructure.KnativeEventingAPIVersion): &knativeInfraResource{},
	}
}
