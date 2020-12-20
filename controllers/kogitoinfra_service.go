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

package controllers

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"

	"strings"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
)

// getKogitoInfraReconciler identify and return request kogito infra reconciliation logic on bases of information provided in kogitoInfra value
func (r *KogitoInfraReconciler) getKogitoInfraReconciler(cli *client.Client, instance *v1beta1.KogitoInfra, scheme *runtime.Scheme) (InfraReconciler, error) {
	r.Log.Debug("going to fetch related kogito infra resource for given", "Infra Instance", instance.Name)
	context := targetContext{
		client:   cli,
		instance: instance,
		scheme:   scheme,
	}
	if infraRes, ok := getSupportedInfraResources(context)[resourceClassForInstance(instance)]; ok {
		return infraRes, nil
	}
	return nil, errorForUnsupportedAPI(instance)
}

func resourceClassForInstance(instance *v1beta1.KogitoInfra) string {
	return getResourceClass(instance.Spec.Resource.Kind, instance.Spec.Resource.APIVersion)
}

func getResourceClass(kind, APIVersion string) string {
	return strings.ToLower(fmt.Sprintf("%s.%s", kind, APIVersion))
}

func getSupportedInfraResources(context targetContext) map[string]InfraReconciler {
	return map[string]InfraReconciler{
		getResourceClass(infrastructure.InfinispanKind, infrastructure.InfinispanAPIVersion):                 initInfinispanInfraReconciler(context),
		getResourceClass(infrastructure.KafkaKind, infrastructure.KafkaAPIVersion):                           initkafkaInfraReconciler(context),
		getResourceClass(infrastructure.KeycloakKind, infrastructure.KeycloakAPIVersion):                     initkeycloakInfraReconciler(context),
		getResourceClass(infrastructure.KnativeEventingBrokerKind, infrastructure.KnativeEventingAPIVersion): initknativeInfraReconciler(context),
		getResourceClass(infrastructure.MongoDBKind, infrastructure.MongoDBAPIVersion):                       initMongoDBInfraReconciler(context),
	}
}
