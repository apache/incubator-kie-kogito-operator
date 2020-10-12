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
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/grafana"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/keycloak"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"

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
	watchedObjects = append(watchedObjects, infinispan.GetWatchedObjects()...)
	watchedObjects = append(watchedObjects, kafka.GetWatchedObjects()...)
	watchedObjects = append(watchedObjects, keycloak.GetWatchedObjects()...)
	watchedObjects = append(watchedObjects, grafana.GetWatchedObjects()...)

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
func (r *ReconcileKogitoInfra) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoInfra for %s in %s", request.Name, request.Namespace)

	// Fetch the KogitoInfra instance
	instance, resultErr := infrastructure.FetchKogitoInfraInstance(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return reconcile.Result{}, resultErr
	}

	defer updateBaseStatus(r.client, instance, &resultErr)

	infraResource, resultErr := GetKogitoInfraResource(instance)
	if resultErr != nil {
		return
	}

	requeue, resultErr := infraResource.Reconcile(r.client, instance, r.scheme)
	if resultErr != nil {
		return
	}

	if requeue {
		log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = time.Second * 30
		result.Requeue = true
	}
	return
}
