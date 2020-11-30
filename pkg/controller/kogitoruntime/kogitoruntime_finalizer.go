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

package kogitoruntime

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
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

// AddFinalizer creates a new KogitoRuntime Finalizer and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddFinalizer(mgr manager.Manager) error {
	return addFinalizer(mgr, newFinalizerReconciler(mgr))
}

// newFinalizerReconciler returns a new reconcile.Reconciler
func newFinalizerReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &FinalizeKogitoRuntime{client: kogitocli.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// FinalizeKogitoRuntime reconciles a KogitoRuntime object
type FinalizeKogitoRuntime struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// addFinalizer adds a new Controller to mgr with r as the reconcile.Reconciler
func addFinalizer(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoRuntime finalizer")
	// Create a new controller
	c, err := controller.New("kogitoruntime-finalizer", mgr, controller.Options{Reconciler: r})
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
	err = c.Watch(&source.Kind{Type: &v1beta1.KogitoRuntime{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}
	return nil
}

// Reconcile reads that state of the cluster for a KogitoRuntime object and makes changes based on the state read
// and what is in the KogitoRuntime.Spec
func (r *FinalizeKogitoRuntime) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	log.Infof("Reconciling KogitoRuntime finalizer for %s in %s", request.Name, request.Namespace)
	instance, err := infrastructure.FetchKogitoRuntimeService(r.client, request.Name, request.Namespace)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debugf("KogitoRuntime instance with name %s not found in namespace %s. Going to return reconciliation request", request.Name, request.Namespace)
		return
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer for this CR
		if err := r.addFinalizer(instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// The object is being deleted
	return r.handleFinalization(instance)
}

func (r *FinalizeKogitoRuntime) addFinalizer(instance *v1beta1.KogitoRuntime) error {
	if len(instance.GetFinalizers()) < 1 && instance.GetDeletionTimestamp() == nil {
		log.Debugf("Adding Finalizer for the KogitoRuntime")
		instance.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer"})

		// Update CR
		if err := kubernetes.ResourceC(r.client).Update(instance); err != nil {
			log.Error("Failed to update finalizer in KogitoRuntime")
			return err
		}
		log.Debugf("Successfully added finalizer into KogitoRuntime instance %s", instance.Name)
	}
	return nil
}

func (r *FinalizeKogitoRuntime) handleFinalization(instance *v1beta1.KogitoRuntime) (reconcile.Result, error) {

	log.Infof("KogitoRuntime object has been deleted for %s in %s", instance.Name, instance.Namespace)

	// Remove KogitoRuntime ownership from referred KogitoInfra instances
	if err := infrastructure.RemoveKogitoInfraOwnership(r.client, instance); err != nil {
		return reconcile.Result{}, err
	}

	// Update finalizer to allow delete CR
	log.Debugf("Removing finalizer from KogitoRuntime instance %s", instance.Name)
	instance.SetFinalizers(nil)
	if err := kubernetes.ResourceC(r.client).Update(instance); err != nil {
		log.Errorf("Error occurs while removing finalizer from KogitoRuntime instance %s", instance.Name, err)
		return reconcile.Result{}, err
	}
	log.Debugf("Successfully removed finalizer from KogitoRuntime instance %s", instance.Name)
	return reconcile.Result{}, nil
}
