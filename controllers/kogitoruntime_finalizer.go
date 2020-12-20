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
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// newKogitoRuntimeFinalizerReconciler returns a new reconcile.Reconciler
//func newKogitoRuntimeFinalizerReconciler(mgr manager.Manager) reconcile.Reconciler {
//	return &FinalizeKogitoRuntime{Client: kogitocli.NewForController(mgr.GetConfig()), Scheme: mgr.GetScheme()}
//}

// FinalizeKogitoRuntime reconciles a KogitoRuntime object
type FinalizeKogitoRuntime struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    logger.Logger
}

// SetupWithManager adds a new Controller to mgr with r as the reconcile.Reconciler
func (f *FinalizeKogitoRuntime) SetupWithManager(mgr manager.Manager) error {
	// Create a new controller
	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return !e.MetaNew.GetDeletionTimestamp().IsZero()
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).For(&v1beta1.KogitoRuntime{}, builder.WithPredicates(pred))
	return b.Complete(f)
}

// Reconcile reads that state of the cluster for a KogitoRuntime object and makes changes based on the state read
// and what is in the KogitoRuntime.Spec
func (f *FinalizeKogitoRuntime) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	f.Log.Info("Reconciling KogitoRuntime finalizer", "instance", request.Name, "namespace", request.Namespace)
	instance, err := infrastructure.FetchKogitoRuntimeService(f.Client, request.Name, request.Namespace)
	if err != nil {
		return
	}
	if instance == nil {
		f.Log.Debug("KogitoRuntime instancenot found. Going to return reconciliation request", "name", request.Name, "namespace", request.Namespace)
		return
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer for this CR
		err = infrastructure.AddFinalizer(f.Client, instance)
		return
	}

	// The object is being deleted
	f.Log.Info("KogitoRuntime has been deleted", "name", instance.GetName(), "namespace", instance.GetNamespace())
	err = infrastructure.HandleFinalization(f.Client, instance)
	return
}
