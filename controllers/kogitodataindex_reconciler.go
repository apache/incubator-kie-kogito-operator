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
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller1 "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// dataIndexSupportingServiceResource implementation of SupportingServiceResource
type dataIndexSupportingServiceResource struct {
	log logger.Logger
}

// Reconcile reconcile Data Index
func (d *dataIndexSupportingServiceResource) Reconcile(client *client.Client, instance *appv1alpha1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, err error) {
	d.log.Info("Reconciling for", "KogitoDataIndex", instance.Name, "Namespace", instance.Namespace)
	if err = infrastructure.InjectDataIndexURLIntoKogitoRuntimeServices(client, instance.Namespace); err != nil {
		return
	}
	if err = infrastructure.InjectDataIndexURLIntoSupportingService(client, instance.Namespace, appv1alpha1.MgmtConsole); err != nil {
		return
	}
	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultDataIndexImageName,
		OnDeploymentCreate: d.dataIndexOnDeploymentCreate,
		KafkaTopics:        dataIndexKafkaTopics,
		Request:            controller1.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		HealthCheckProbe:   services.QuarkusHealthCheckProbe,
	}
	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}

// Collection of kafka topics that should be handled by the Data-Index service
var dataIndexKafkaTopics = []string{
	"kogito-processinstances-events",
	"kogito-usertaskinstances-events",
	"kogito-processdomain-events",
	"kogito-usertaskdomain-events",
	"kogito-jobs-events",
}

func (d *dataIndexSupportingServiceResource) dataIndexOnDeploymentCreate(cli *client.Client, deployment *appsv1.Deployment, kogitoService appv1alpha1.KogitoService) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if err := infrastructure.MountProtoBufConfigMapsOnDeployment(cli, deployment); err != nil {
			return err
		}
	} else {
		d.log.Warn("No container definition found for", "Service", kogitoService.GetName())
		d.log.Warn("Skipping applying custom Data Index deployment configuration")
	}
	return nil
}
