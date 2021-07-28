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
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/connector"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/apps/v1"
	"reflect"
)

const (
	envVarExternalURL = "KOGITO_SERVICE_URL"
	envVarNamespace   = "NAMESPACE"
)

// RuntimeDeployerHandler ...
type RuntimeDeployerHandler interface {
	OnGetComparators(comparator compare.ResourceComparator)
	OnDeploymentCreate(deployment *v1.Deployment) error
}

type runtimeDeployerHandler struct {
	operator.Context
	instance                 api.KogitoRuntimeInterface
	supportingServiceHandler manager.KogitoSupportingServiceHandler
	runtimeHandler           manager.KogitoRuntimeHandler
}

// NewRuntimeDeployerHandler ...
func NewRuntimeDeployerHandler(context operator.Context, instance api.KogitoRuntimeInterface, supportingServiceHandler manager.KogitoSupportingServiceHandler, runtimeHandler manager.KogitoRuntimeHandler) RuntimeDeployerHandler {
	return &runtimeDeployerHandler{
		Context:                  context,
		instance:                 instance,
		supportingServiceHandler: supportingServiceHandler,
		runtimeHandler:           runtimeHandler,
	}
}

func (d *runtimeDeployerHandler) OnGetComparators(comparator compare.ResourceComparator) {
	comparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(monv1.ServiceMonitor{})).
			WithCustomComparator(framework.CreateServiceMonitorComparator()).
			Build())
}

// onDeploymentCreate hooks into the infrastructure package to add additional capabilities/properties to the deployment creation
func (d *runtimeDeployerHandler) OnDeploymentCreate(deployment *v1.Deployment) error {
	// NAMESPACE service discovery
	framework.SetEnvVar(envVarNamespace, d.instance.GetNamespace(), &deployment.Spec.Template.Spec.Containers[0])
	// external URL
	if d.instance.GetStatus().GetExternalURI() != "" {
		framework.SetEnvVar(envVarExternalURL, d.instance.GetStatus().GetExternalURI(), &deployment.Spec.Template.Spec.Containers[0])
	}
	// sa
	deployment.Spec.Template.Spec.ServiceAccountName = infrastructure.RuntimeServiceAccountName
	// istio
	if d.instance.GetRuntimeSpec().IsEnableIstio() {
		framework.AddIstioInjectSidecarAnnotation(&deployment.Spec.Template.ObjectMeta)
	}

	urlHandler := connector.NewURLHandler(d.Context, d.runtimeHandler, d.supportingServiceHandler)
	if err := urlHandler.InjectDataIndexURLIntoDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := urlHandler.InjectJobsServiceURLIntoKogitoRuntimeDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := urlHandler.InjectTrustyURLIntoDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	return nil
}
