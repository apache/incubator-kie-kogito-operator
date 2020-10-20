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
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitosupportingservice/dataindex"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
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

// Add creates a new KogitoSupportingService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddDataIndex(mgr manager.Manager) error {
	return addDataIndex(mgr, newDataIndexReconciler(mgr))
}

// newDataIndexReconciler returns a new reconcile.Reconciler
func newDataIndexReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoSupportingService{client: client.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addDataIndex(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for DataIndex controller")
	// Create a new controller
	c, err := controller.New("DataIndex-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoSupportingService
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			supportingService := e.Object.(*appv1alpha1.KogitoSupportingService)
			return supportingService.Spec.ServiceType == appv1alpha1.DataIndex
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			supportingService := e.ObjectNew.(*appv1alpha1.KogitoSupportingService)
			return supportingService.Spec.ServiceType == appv1alpha1.DataIndex
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
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
			GroupVersion: imgv1.GroupVersion,
			AddToScheme:  imgv1.Install,
			Objects:      []runtime.Object{&imgv1.ImageStream{}},
		},
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.ConfigMap{}}},
	}
	resource := &dataindex.SupportingServiceResource{}
	watchedObjects = append(watchedObjects, resource.GetWatchedObjects()...)
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoSupportingService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoDataIndex{}

// ReconcileKogitoSupportingService reconciles a KogitoSupportingService object
type ReconcileKogitoDataIndex struct {
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
func (r *ReconcileKogitoDataIndex) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoSupportingService for %s in %s", request.Name, request.Namespace)

	instance, resultErr := fetchKogitoSupportingService(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return
	}

	if resultErr = ensureSingletonService(r.client, request.Namespace, instance.Spec.ServiceType); resultErr != nil {
		return
	}

	supportingResource := &dataindex.SupportingServiceResource{}

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
