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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TrustyAISupportingServiceResource implementation of SupportingServiceResource
type TrustyAISupportingServiceResource struct {
}

// Reconcile reconcile TrustyAI Service
func (*TrustyAISupportingServiceResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoSupportingService, scheme *runtime.Scheme) (requeue bool, err error) {
	log.Info("Reconciling KogitoTrusty")
	log.Infof("Injecting Trusty Index URL into KogitoApps in the namespace '%s'", instance.Namespace)
	if err := infrastructure.InjectTrustyURLIntoKogitoRuntimeServices(client, instance.Namespace); err != nil {
		return false, err
	}
	definition := services.ServiceDefinition{
		DefaultImageName: infrastructure.DefaultTrustyImageName,
		Request:          controller.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		KafkaTopics:      trustyAiKafkaTopics,
		HealthCheckProbe: services.QuarkusHealthCheckProbe,
	}
	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}

// Collection of kafka topics that should be handled by the Trusty service
var trustyAiKafkaTopics = []string{
	"kogito-tracing-decision",
	"kogito-tracing-model",
	"trusty-explainability-result",
	"trusty-explainability-request",
}
