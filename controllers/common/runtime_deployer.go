// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package common

import (
	"github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/core/connector"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	v1 "k8s.io/api/apps/v1"
)

const (
	envVarExternalURL = "KOGITO_SERVICE_URL"
	envVarNamespace   = "NAMESPACE"
)

// RuntimeDeployerHandler ...
type RuntimeDeployerHandler interface {
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

// OnDeploymentCreate hooks into the infrastructure package to add additional capabilities/properties
// to the deployment creation
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
	if err := urlHandler.InjectDataIndexEndpointOnDeployment(deployment); err != nil {
		return err
	}

	if err := urlHandler.InjectJobsServiceEndpointOnDeployment(deployment); err != nil {
		return err
	}

	return urlHandler.InjectTrustyEndpointOnDeployment(deployment)
}
