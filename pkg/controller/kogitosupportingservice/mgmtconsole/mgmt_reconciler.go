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

package mgmtconsole

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logger.GetLogger("mgmtconsole_reconciler")

// SupportingServiceResource implementation of SupportingServiceResource
type SupportingServiceResource struct {
}

// Reconcile reconcile Management Console
func (*SupportingServiceResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoSupportingService, scheme *runtime.Scheme) (requeue bool, err error) {
	log.Infof("Reconciling KogitoMgmtConsole for %s in %s", instance.Name, instance.Namespace)
	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultMgmtConsoleImageName,
		Request:            controller.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		SingleReplica:      false,
		OnDeploymentCreate: onDeploymentCreate,
	}
	return services.NewServiceDeployer(definition, instance, client, scheme).Deploy()
}
