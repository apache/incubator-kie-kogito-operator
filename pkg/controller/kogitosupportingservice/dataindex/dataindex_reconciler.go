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

package dataindex

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logger.GetLogger("dataindex_reconciler")

// SupportingServiceResource implementation of SupportingServiceResource
type SupportingServiceResource struct {
}

// GetWatchedObjects provide list of object that needs to be watched to maintain DataIndex Service
func (*SupportingServiceResource) GetWatchedObjects() []framework.WatchedObjects {

	return []framework.WatchedObjects{
		{
			GroupVersion: corev1.SchemeGroupVersion,
			Objects:      []runtime.Object{&corev1.ConfigMap{}},
			Owner:        &v1alpha1.KogitoRuntime{},
		},
	}
}

// Reconcile reconcile Data index service
func (*SupportingServiceResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoSupportingService, scheme *runtime.Scheme) (requeue bool, err error) {
	log.Infof("Injecting Data Index URL into KogitoRuntime services in the namespace '%s'", instance.Namespace)
	if err = infrastructure.InjectDataIndexURLIntoKogitoRuntimeServices(client, instance.Namespace); err != nil {
		return
	}
	if err = infrastructure.InjectDataIndexURLIntoMgmtConsole(client, instance.Namespace); err != nil {
		return
	}

	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultDataIndexImageName,
		OnDeploymentCreate: onDeploymentCreate,
		KafkaTopics:        kafkaTopics,
		Request:            controller.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		HealthCheckProbe:   services.QuarkusHealthCheckProbe,
	}
	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}

// Collection of kafka topics that should be handled by the Data-Index service
var kafkaTopics = []string{
	"kogito-processinstances-events",
	"kogito-usertaskinstances-events",
	"kogito-processdomain-events",
	"kogito-usertaskdomain-events",
	"kogito-jobs-events",
}
