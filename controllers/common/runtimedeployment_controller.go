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

package common

import (
	"context"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	dep "github.com/kiegroup/kogito-operator/core/deployment"
	"github.com/kiegroup/kogito-operator/core/framework/util"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RuntimeDeploymentReconciler ...
type RuntimeDeploymentReconciler struct {
	*kogitocli.Client
	Scheme                *runtime.Scheme
	Version               string
	RuntimeHandler        func(context operator.Context) manager.KogitoRuntimeHandler
	SupportServiceHandler func(context operator.Context) manager.KogitoSupportingServiceHandler
	Labels                map[string]string
}

// Reconcile ...
func (r *RuntimeDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling for Deployment")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:  r.Client,
		Log:     log,
		Scheme:  r.Scheme,
		Version: r.Version,
		Labels:  r.Labels,
	}

	deploymentHandler := infrastructure.NewDeploymentHandler(kogitoContext)
	deployment, err := deploymentHandler.FetchDeployment(req.NamespacedName)
	if err != nil {
		return
	}
	if deployment == nil {
		log.Debug("KogitoDeployment instance not found")
		return
	}

	runtimeHandler := r.RuntimeHandler(kogitoContext)
	supportingServiceHandler := r.SupportServiceHandler(kogitoContext)
	depProcessor := dep.NewDeploymentProcessor(kogitoContext, deployment, runtimeHandler, supportingServiceHandler)
	err = depProcessor.Process()
	if err != nil {
		return infrastructure.NewReconciliationErrorHandler(kogitoContext).GetReconcileResultFor(err)
	}
	log.Debug("Finish reconciliation", "requeue", result.Requeue, "requeueAfter", result.RequeueAfter)
	return
}

// SetupWithManager registers the controller with manager
func (r *RuntimeDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {

	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return util.MapContains(e.Object.GetAnnotations(), operator.KogitoRuntimeKey, "true")
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return util.MapContains(e.ObjectNew.GetAnnotations(), operator.KogitoRuntimeKey, "true") && e.ObjectNew.GetDeletionTimestamp().IsZero()
		},
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(pred))
	return b.Complete(r)
}
