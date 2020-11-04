// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package kogitoinfra

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("kogitoinfra_controller")

const reconciliationStandardInterval = time.Second * 30

// Add creates a new KogitoInfra Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoInfra{client: client.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoInfra controller")
	// Create a new controller
	c, err := controller.New("kogitoinfra-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoInfra
	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by Kubernetes itself
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoInfra{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	var watchedObjects []framework.WatchedObjects
	watchedObjects = append(watchedObjects, getInfinispanWatchedObjects()...)
	watchedObjects = append(watchedObjects, getKafkaWatchedObjects()...)
	watchedObjects = append(watchedObjects, getKeycloakWatchedObjects()...)

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoInfra).client, mgr, c, &appv1alpha1.KogitoInfra{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKogitoInfra implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoInfra{}

// ReconcileKogitoInfra reconciles a KogitoInfra object
type ReconcileKogitoInfra struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoInfra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconciling KogitoInfra for %s in %s", request.Name, request.Namespace)

	// Fetch the KogitoInfra instance
	instance, resultErr := infrastructure.MustFetchKogitoInfraInstance(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return reconcile.Result{}, resultErr
	}

	// make KogitoInfra as self owner so that it will not removed when kogito service referring to it deleted because
	// kogito services are also become owner of kogitoInfra when reference of infra provided in Kogito services.
	if resultErr = framework.AddOwnerReference(instance, r.scheme, instance); resultErr != nil {
		return reconcile.Result{}, resultErr
	}

	defer updateBaseStatus(r.client, instance, &resultErr)

	reconciler, resultErr := getKogitoInfraReconciler(r.client, instance, r.scheme)
	if resultErr != nil {
		return r.getReconcileResultFor(resultErr, false)
	}

	requeue, resultErr := reconciler.Reconcile()
	return r.getReconcileResultFor(resultErr, requeue)
}

func (r *ReconcileKogitoInfra) getReconcileResultFor(err error, requeue bool) (reconcile.Result, error) {
	// generic reconciliation error
	if reasonForError(err) == appv1alpha1.ReconciliationFailure {
		log.Warnf("Error while reconciling KogitoInfra: %s", err.Error())
		return reconcile.Result{RequeueAfter: 0, Requeue: false}, err
	}
	// no requeue, no errors, stop reconciliation
	if !requeue && err == nil {
		log.Info("No need reconciliation for KogitoInfra")
		return reconcile.Result{RequeueAfter: 0, Requeue: false}, nil
	}
	// caller is asking for a reconciliation
	if err == nil {
		log.Info("Waiting for all resources to be created, scheduling reconciliation")
	} else { // reconciliation duo to a problem in the env (CRDs missing), infra deployments not ready, operators not installed.. etc. See errors.go
		log.Infof("Waiting for all resources to be created, scheduling reconciliation: %s", err.Error())
	}
	log.Info("Scheduling reconciliation for %d", reconciliationStandardInterval)
	return reconcile.Result{RequeueAfter: reconciliationStandardInterval}, nil
}
