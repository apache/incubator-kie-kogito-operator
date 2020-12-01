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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddFinalizer creates a new kogitosupportingservice finalizer and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddFinalizer(mgr manager.Manager) error {
	return addFinalizer(mgr, newFinalizerReconciler(mgr))
}

// newFinalizerReconciler returns a new reconcile.Reconciler
func newFinalizerReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &FinalizeKogitoSupportingService{client: kogitocli.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// FinalizeKogitoSupportingService reconciles a KogitoSupportingService object
type FinalizeKogitoSupportingService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// addFinalizer adds a new Controller to mgr with r as the reconcile.Reconciler
func addFinalizer(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoSupportingService finalizer")
	// Create a new controller
	c, err := controller.New("KogitoSupportingService-finalizer", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return !e.MetaNew.GetDeletionTimestamp().IsZero()
		},
	}
	err = c.Watch(&source.Kind{Type: &v1beta1.KogitoSupportingService{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}
	return nil
}

// Reconcile reads that state of the cluster for a KogitoSupportingService object and makes changes based on the state read
// and what is in the KogitoSupportingService.Spec
func (r *FinalizeKogitoSupportingService) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	log.Infof("Reconciling KogitoSupportingService finalizer for %s in %s", request.Name, request.Namespace)
	instance, err := fetchKogitoSupportingService(r.client, request.Name, request.Namespace)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debugf("KogitoSupportingService instance with name %s not found in namespace %s. Going to return reconciliation request", request.Name, request.Namespace)
		return
	}
	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer for this CR
		err = infrastructure.AddFinalizer(r.client, instance)
		return
	}

	// The object is being deleted
	log.Infof("KogitoSupportingService(%s) has been deleted in %s", instance.GetName(), instance.GetNamespace())
	err = infrastructure.HandleFinalization(r.client, instance)
	return
}
