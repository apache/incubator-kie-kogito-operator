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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// JobsServiceSupportingServiceResource implementation of SupportingServiceResource
type JobsServiceSupportingServiceResource struct {
}

// Reconcile reconcile Jobs service
func (*JobsServiceSupportingServiceResource) Reconcile(client *client.Client, instance *v1beta1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, err error) {
	log.Infof("Reconciling KogitoJobsService for %s in %s", instance.Name, instance.Namespace)

	// clean up variables if needed
	if err = infrastructure.InjectJobsServicesURLIntoKogitoRuntimeServices(client, instance.Namespace); err != nil {
		return
	}
	definition := services.ServiceDefinition{
		DefaultImageName: infrastructure.DefaultJobsServiceImageName,
		Request:          controller.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		SingleReplica:    true,
		HealthCheckProbe: services.QuarkusHealthCheckProbe,
		KafkaTopics:      jobsServicekafkaTopics,
	}

	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}

// Collection of kafka topics that should be handled by the Jobs service
var jobsServicekafkaTopics = []string{
	"kogito-job-service-job-status-events",
}
