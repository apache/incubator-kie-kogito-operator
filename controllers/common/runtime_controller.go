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

package common

import (
	"context"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/shared"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KogitoRuntimeReconciler reconciles a KogitoRuntime object
type KogitoRuntimeReconciler struct {
	*kogitocli.Client
	Scheme                *runtime.Scheme
	Version               string
	RuntimeHandler        func(context operator.Context) manager.KogitoRuntimeHandler
	SupportServiceHandler func(context operator.Context) manager.KogitoSupportingServiceHandler
	InfraHandler          func(context operator.Context) manager.KogitoInfraHandler
	ReconcilingObject     client.Object
	MeteringLabels        map[string]string
	DeploymentLabels      map[string]string
}

// Reconcile reads that state of the cluster for a KogitoRuntime object and makes changes based on the state read
// and what is in the KogitoRuntime.Spec
func (r *KogitoRuntimeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling for KogitoRuntime")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:           r.Client,
		Log:              log,
		Scheme:           r.Scheme,
		Version:          r.Version,
		MeteringLabels:   r.MeteringLabels,
		DeploymentLabels: r.DeploymentLabels,
	}

	// fetch the requested instance
	runtimeHandler := r.RuntimeHandler(kogitoContext)
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(req.NamespacedName)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debug("KogitoRuntime instance not found")
		return
	}

	rbacHandler := infrastructure.NewRBACHandler(kogitoContext)
	if err = rbacHandler.SetupRBAC(req.Namespace); err != nil {
		return
	}

	supportingServiceHandler := r.SupportServiceHandler(kogitoContext)
	deploymentHandler := NewRuntimeDeployerHandler(kogitoContext, instance, supportingServiceHandler, runtimeHandler)
	definition := kogitoservice.ServiceDefinition{
		Request:            req,
		DefaultImageTag:    infrastructure.LatestTag,
		SingleReplica:      false,
		OnDeploymentCreate: deploymentHandler.OnDeploymentCreate,
		CustomService:      true,
	}
	infraHandler := r.InfraHandler(kogitoContext)
	err = kogitoservice.NewServiceDeployer(kogitoContext, definition, instance, infraHandler).Deploy()
	if err != nil {
		return framework.NewReconciliationErrorHandler(kogitoContext).GetReconcileResultFor(err)
	}

	protoBufHandler := shared.NewProtoBufHandler(kogitoContext, supportingServiceHandler)
	err = protoBufHandler.MountProtoBufConfigMapOnDataIndex(instance)
	if err != nil {
		log.Error(err, "Fail to mount Proto Buf config map of Kogito runtime on DataIndex")
		return framework.NewReconciliationErrorHandler(kogitoContext).GetReconcileResultFor(err)
	}

	log.Debug("Finish reconciliation", "requeue", result.Requeue, "requeueAfter", result.RequeueAfter)
	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).
		For(r.ReconcilingObject, builder.WithPredicates(pred)).
		Owns(&corev1.Service{}).Owns(&appsv1.Deployment{}).Owns(&corev1.ConfigMap{})

	if r.IsOpenshift() {
		b.Owns(&routev1.Route{}).Owns(&imagev1.ImageStream{})
	}

	return b.Complete(r)
}
