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
	"context"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/internal"
	"github.com/kiegroup/kogito-operator/version"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// FinalizeKogitoSupportingService reconciles a KogitoSupportingService object
type FinalizeKogitoSupportingService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
}

// SetupWithManager adds a new Controller to mgr with r as the reconcile.Reconciler
func (f *FinalizeKogitoSupportingService) SetupWithManager(mgr manager.Manager) error {
	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return !e.ObjectNew.GetDeletionTimestamp().IsZero()
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).For(&v1beta1.KogitoSupportingService{}, builder.WithPredicates(pred))
	// Create a new controller

	return b.Complete(f)
}

// Reconcile reads that state of the cluster for a KogitoSupportingService object and makes changes based on the state read
// and what is in the KogitoSupportingService.Spec
func (f *FinalizeKogitoSupportingService) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling for KogitoSupportingService finalizer")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:  f.Client,
		Log:     log,
		Scheme:  f.Scheme,
		Version: version.Version,
	}

	supportingServiceHandler := internal.NewKogitoSupportingServiceHandler(kogitoContext)
	instance, err := supportingServiceHandler.FetchKogitoSupportingService(request.NamespacedName)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debug("KogitoSupportingService instance not found. Going to return reconciliation request", "name", request.Name, "namespace", request.Namespace)
		return
	}
	infraHandler := internal.NewKogitoInfraHandler(kogitoContext)
	infraFinalizer := kogitoservice.NewInfraFinalizerHandler(kogitoContext, infraHandler)
	configMapFinalizer := kogitoservice.NewConfigMapFinalizerHandler(kogitoContext)
	// examine DeletionTimestamp to determine if object is under deletion
	if instance.GetDeletionTimestamp().IsZero() {
		// Add finalizer for this CR
		if err = infraFinalizer.AddFinalizer(instance); err != nil {
			result.Requeue = true
			return
		}
		if err = configMapFinalizer.AddFinalizer(instance); err != nil {
			result.Requeue = true
			return
		}
		return
	}

	// The object is being deleted
	log.Info("KogitoSupportingService has been deleted")
	if err = infraFinalizer.HandleFinalization(instance); err != nil {
		result.Requeue = true
		return
	}
	if err = configMapFinalizer.HandleFinalization(instance); err != nil {
		result.Requeue = true
		return
	}
	result.Requeue = false
	return
}
