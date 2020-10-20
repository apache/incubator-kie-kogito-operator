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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Collection of Services not to be monitored by this controller
var ignoreServices = []appv1alpha1.ServiceType{
	appv1alpha1.DataIndex,
}
var log = logger.GetLogger("kogitosupportingservice_controller")

// Add creates a new KogitoSupportingService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoSupportingService{client: client.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoSupportingService controller")
	// Create a new controller
	c, err := controller.New("kogitosupportingservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoSupportingService
	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by Kubernetes itself
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			supportingService := e.Object.(*appv1alpha1.KogitoSupportingService)
			return !contains(ignoreServices, supportingService.Spec.ServiceType)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			supportingService := e.ObjectNew.(*appv1alpha1.KogitoSupportingService)
			return !contains(ignoreServices, supportingService.Spec.ServiceType)
		},
	}
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoSupportingService{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoSupportingService).client, mgr, c, &appv1alpha1.KogitoSupportingService{})
	watchedObjects := []framework.WatchedObjects{
		{
			GroupVersion: routev1.GroupVersion,
			AddToScheme:  routev1.Install,
			Objects:      []runtime.Object{&routev1.Route{}},
		},

		{
			Objects:      []runtime.Object{&appv1alpha1.KogitoInfra{}},
			Owner: 	&appv1alpha1.KogitoSupportingService{},
			Predicate: predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return true
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return true
				},
			},
		},
		{
			GroupVersion: imgv1.GroupVersion,
			AddToScheme:  imgv1.Install,
			Objects:      []runtime.Object{&imgv1.ImageStream{}},
		},
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.ConfigMap{}}},
	}
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoSupportingService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoSupportingService{}

// ReconcileKogitoSupportingService reconciles a KogitoSupportingService object
type ReconcileKogitoSupportingService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoSupportingService object and makes changes based on the state read
// and what is in the KogitoSupportingService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoSupportingService) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoSupportingService for %s in %s", request.Name, request.Namespace)
	// Fetch the KogitoSupportingService instance
	instance, resultErr := fetchKogitoSupportingService(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return
	}

	if resultErr = ensureSingletonService(r.client, request.Namespace, instance.Spec.ServiceType); resultErr != nil {
		return
	}

	supportingResource := getKogitoSupportingResource(instance)

	requeue, resultErr := supportingResource.Reconcile(r.client, instance, r.scheme)
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
