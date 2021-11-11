// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package deployment

import (
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/shared"
	v1 "k8s.io/api/apps/v1"
)

// Processor ...
type Processor interface {
	Process() error
}

type deploymentProcessor struct {
	operator.Context
	deployment               *v1.Deployment
	supportingServiceHandler manager.KogitoSupportingServiceHandler
}

// NewDeploymentProcessor ...
func NewDeploymentProcessor(context operator.Context, deployment *v1.Deployment, supportingServiceHandler manager.KogitoSupportingServiceHandler) Processor {
	return &deploymentProcessor{
		Context:                  context,
		deployment:               deployment,
		supportingServiceHandler: supportingServiceHandler,
	}
}

func (d *deploymentProcessor) Process() (err error) {

	serviceReconciler := infrastructure.NewServiceReconciler(d.Context, d.deployment)
	if err = serviceReconciler.Reconcile(); err != nil {
		d.Recorder.Eventf(d.deployment, "Normal", "Configuring Service", "Error occurs while configuring Service. Error : %s", err.Error())
		return err
	}

	routeReconciler := infrastructure.NewRouteReconciler(d.Context, d.deployment)
	if err = routeReconciler.Reconcile(); err != nil {
		d.Recorder.Eventf(d.deployment, "Normal", "Configuring Route", "Error occurs while configuring Route. Error : %s", err.Error())
		return err
	}

	prometheusManager := infrastructure.NewPrometheusManager(d.Context, d.deployment)
	if err = prometheusManager.ConfigurePrometheus(); err != nil {
		d.Log.Error(err, "Could not deploy prometheus monitoring")
		d.Recorder.Eventf(d.deployment, "Normal", "Configuring Prometheus", "Error occurs while configuring Prometheus. Error : %s", err.Error())
		return err
	}

	grafanaDashboardManager := infrastructure.NewGrafanaDashboardManager(d.Context, d.deployment)
	if err = grafanaDashboardManager.ConfigureGrafanaDashboards(); err != nil {
		d.Log.Error(err, "Could not deploy grafana dashboards")
		d.Recorder.Eventf(d.deployment, "Normal", "Configuring Grafana", "Error occurs while configuring Grafana. Error : %s", err.Error())
		return err
	}

	protoBufHandler := shared.NewProtoBufHandler(d.Context, d.supportingServiceHandler)
	err = protoBufHandler.MountProtoBufConfigMapOnDataIndex(d.deployment)
	if err != nil {
		d.Log.Error(err, "Fail to mount Proto Buf config map of Kogito runtime on DataIndex")
		d.Recorder.Eventf(d.deployment, "Normal", "Mount Protobuf", "Error occurs while mounting Protobuf file on DataIndex. Error : %s", err.Error())
		return err
	}

	return
}
