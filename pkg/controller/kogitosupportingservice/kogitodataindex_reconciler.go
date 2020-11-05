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
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller1 "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// DataIndexSupportingServiceResource implementation of SupportingServiceResource
type DataIndexSupportingServiceResource struct {
}

// Reconcile reconcile Data Index
func (*DataIndexSupportingServiceResource) Reconcile(client *client.Client, instance *appv1alpha1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, err error) {
	log.Infof("Reconciling KogitoDataIndex for %s in %s", instance.Name, instance.Namespace)
	if err = infrastructure.InjectDataIndexURLIntoKogitoRuntimeServices(client, instance.Namespace); err != nil {
		return
	}
	if err = infrastructure.InjectDataIndexURLIntoSupportingService(client, instance.Namespace, appv1alpha1.MgmtConsole); err != nil {
		return
	}
	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultDataIndexImageName,
		OnDeploymentCreate: dataIndexOnDeploymentCreate,
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

func dataIndexOnDeploymentCreate(cli *client.Client, deployment *appsv1.Deployment, kogitoService appv1alpha1.KogitoService) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if err := infrastructure.MountProtoBufConfigMapsOnDeployment(cli, deployment); err != nil {
			return err
		}
	} else {
		log.Warnf("No container definition for service %s. Skipping applying custom Data Index deployment configuration", kogitoService.GetName())
	}
	return nil
}
