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
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/record"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KogitoDeploymentReconciler ...
type KogitoDeploymentReconciler struct {
	*kogitocli.Client
	Scheme         *runtime.Scheme
	Version        string
	MeteringLabels map[string]string
	Predicate      predicate.Predicate
}

// Reconcile ...
func (r *KogitoDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling for Deployment")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:   r.Client,
		Log:      log,
		Scheme:   r.Scheme,
		Version:  r.Version,
		Recorder: record.NewRecorder(r.Client, r.Scheme, corev1.EventSource{Component: req.Name, Host: record.GetHostName()}),
	}

	deploymentHandler := infrastructure.NewDeploymentHandler(kogitoContext)
	deployment, err := deploymentHandler.FetchDeployment(types.NamespacedName{Name: req.Name, Namespace: req.Namespace})
	if err != nil {
		return
	}
	if deployment == nil {
		log.Debug("KogitoDeployment instance not found")
		return
	}

	depProcessor := dep.NewDeploymentProcessor(kogitoContext, deployment)
	err = depProcessor.Process()
	if err != nil {
		return framework.NewReconciliationErrorHandler(kogitoContext).GetReconcileResultFor(err)
	}
	log.Debug("Finish reconciliation", "requeue", result.Requeue, "requeueAfter", result.RequeueAfter)
	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(r.Predicate)).
		Owns(&corev1.Service{})
	if r.IsOpenshift() {
		b.Owns(&routev1.Route{})
	}
	return b.Complete(r)
}
