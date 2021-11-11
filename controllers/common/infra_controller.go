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
	"github.com/kiegroup/kogito-operator/core/kogitoinfra"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/record"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// KogitoInfraReconciler reconciles a KogitoInfra object
type KogitoInfraReconciler struct {
	*kogitocli.Client
	Scheme            *runtime.Scheme
	Version           string
	InfraHandler      func(context operator.Context) manager.KogitoInfraHandler
	ReconcilingObject client.Object
}

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
func (r *KogitoInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling KogitoInfra")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:   r.Client,
		Log:      log,
		Scheme:   r.Scheme,
		Version:  r.Version,
		Recorder: record.NewRecorder(r.Client, r.Scheme, req.Name),
	}

	// Fetch the KogitoInfra instance
	infraHandler := r.InfraHandler(kogitoContext)
	instance, err := infraHandler.FetchKogitoInfraInstance(req.NamespacedName)
	if err != nil {
		return reconcile.Result{}, err
	}
	if instance == nil {
		log.Debug("KogitoInfra instance not found")
		return reconcile.Result{}, nil
	}
	var resultErr error
	statusHandler := kogitoinfra.NewStatusHandler(kogitoContext, infraHandler)
	defer statusHandler.UpdateBaseStatus(instance, &resultErr)

	instance.GetStatus().SetEnvs(nil)
	instance.GetStatus().SetConfigMapEnvFromReferences(nil)
	instance.GetStatus().SetConfigMapVolumeReferences(nil)
	instance.GetStatus().SetSecretEnvFromReferences(nil)
	instance.GetStatus().SetSecretVolumeReferences(nil)

	reconcilerHandler := kogitoinfra.NewReconcilerHandler(kogitoContext)
	if !instance.GetSpec().IsResourceEmpty() {
		var reconciler kogitoinfra.Reconciler
		reconciler, resultErr = reconcilerHandler.GetInfraReconciler(instance)
		if resultErr != nil {
			return reconcilerHandler.GetReconcileResultFor(resultErr, false)
		}

		resultErr = reconciler.Reconcile()
		if resultErr != nil {
			return reconcilerHandler.GetReconcileResultFor(resultErr, false)
		}
	}

	appConfigMapReconciler := reconcilerHandler.GetInfraPropertiesReconciler(instance)
	if resultErr = appConfigMapReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	// Set envs in status
	if len(instance.GetSpec().GetEnvs()) > 0 {
		instance.GetStatus().AddEnvs(instance.GetSpec().GetEnvs())
	}

	configMapReferenceReconciler := reconcilerHandler.GetConfigMapReferenceReconciler(instance)
	if resultErr = configMapReferenceReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	secretReferenceReconciler := reconcilerHandler.GetSecretReferenceReconciler(instance)
	if resultErr = secretReferenceReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	return reconcile.Result{}, nil
}

// SetupWithManager registers the controller with manager
func (r *KogitoInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// do not call reconciler if only owner reference is added or deleted
			return reflect.DeepEqual(e.ObjectNew.GetOwnerReferences(), e.ObjectOld.GetOwnerReferences())
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).For(r.ReconcilingObject, builder.WithPredicates(pred))
	b = kogitoinfra.AppendInfinispanWatchedObjects(b)
	b = kogitoinfra.AppendKafkaWatchedObjects(b)
	b = kogitoinfra.AppendKeycloakWatchedObjects(b)
	b = kogitoinfra.AppendMongoDBWatchedObjects(b)
	b = kogitoinfra.AppendConfigMapWatchedObjects(b)
	b = kogitoinfra.AppendSecretWatchedObjects(b)
	return b.Complete(r)
}
