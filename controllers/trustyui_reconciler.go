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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// trustyUISupportingServiceResource implementation of SupportingServiceResource
type trustyUISupportingServiceResource struct {
	log logger.Logger
}

// Reconcile reconcile TrustyUI Service
func (t *trustyUISupportingServiceResource) Reconcile(client *client.Client, instance *v1beta1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, err error) {
	t.log.Info("Reconciling for", "KogitoTrustyUI", instance.Name, "Namespace", instance.Namespace)
	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultTrustyUIImageName,
		Request:            controller.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		SingleReplica:      false,
		OnDeploymentCreate: trustyUIOnDeploymentCreate,
	}
	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}

func trustyUIOnDeploymentCreate(cli *client.Client, deployment *appsv1.Deployment, kogitoService v1beta1.KogitoService) error {
	if err := infrastructure.InjectTrustyURLIntoDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}
	return nil
}
