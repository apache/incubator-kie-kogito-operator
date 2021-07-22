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
	appv1beta1 "github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitosupportingservice"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/internal"
	"github.com/kiegroup/kogito-operator/version"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//var log = logger.GetLogger("kogitoSupportingService_controller")

// KogitoSupportingServiceReconciler reconciles a KogitoSupportingService object
type KogitoSupportingServiceReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	*client.Client
	Log    logger.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitosupportingservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitosupportingservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitosupportingservices/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;list;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams;imagestreamtags,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=core,resources=configmaps;events;pods;secrets;services,verbs=create;delete;get;list;patch;update;watch

// Reconcile reads that state of the cluster for a KogitoSupportingService object and makes changes based on the state read
// and what is in the KogitoSupportingService.Spec
func (r *KogitoSupportingServiceReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, resultErr error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling for KogitoSupportingService")

	// create context
	context := operator.Context{
		Client:  r.Client,
		Log:     log,
		Scheme:  r.Scheme,
		Version: version.Version,
	}

	// Fetch the KogitoSupportingService instance
	supportingServiceHandler := internal.NewKogitoSupportingServiceHandler(context)
	instance, resultErr := supportingServiceHandler.FetchKogitoSupportingService(req.NamespacedName)
	if resultErr != nil {
		return
	}
	if instance == nil {
		log.Debug("kogitoSupportingService Instance not found")
		return
	}

	supportingServiceManager := manager.NewKogitoSupportingServiceManager(context, supportingServiceHandler)
	if resultErr = supportingServiceManager.EnsureSingletonService(req.Namespace, instance.GetSupportingServiceSpec().GetServiceType()); resultErr != nil {
		return
	}

	runtimeHandler := internal.NewKogitoRuntimeHandler(context)
	infraHandler := internal.NewKogitoInfraHandler(context)
	reconcileHandler := kogitosupportingservice.NewReconcilerHandler(context, infraHandler, supportingServiceHandler, runtimeHandler)
	reconciler := reconcileHandler.GetSupportingServiceReconciler(instance)
	resultErr = reconciler.Reconcile()
	if resultErr != nil {
		return infrastructure.NewReconciliationErrorHandler(context).GetReconcileResultFor(resultErr)
	}
	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoSupportingServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info("Adding watched objects for KogitoSupportingService controller")

	pred := predicate.Funcs{
		// Don't watch delete events as the resource removals will be handled by its finalizer
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.MetaNew.GetDeletionTimestamp().IsZero()
		},
	}

	b := ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.KogitoSupportingService{}, builder.WithPredicates(pred)).
		Owns(&corev1.Service{}).Owns(&appsv1.Deployment{}).Owns(&corev1.ConfigMap{})

	infraHandler := &handler.EnqueueRequestForOwner{IsController: false, OwnerType: &appv1beta1.KogitoSupportingService{}}
	infraPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return reflect.DeepEqual(e.MetaNew.GetOwnerReferences(), e.MetaOld.GetOwnerReferences())
		},
	}
	b.Watches(&source.Kind{Type: &appv1beta1.KogitoInfra{}}, infraHandler, builder.WithPredicates(infraPred))

	if r.IsOpenshift() {
		b.Owns(&routev1.Route{}).Owns(&imgv1.ImageStream{})
	}
	return b.Complete(r)
}
