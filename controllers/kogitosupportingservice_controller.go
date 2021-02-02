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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	imgv1 "github.com/openshift/api/image/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1beta1 "github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
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
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;list;delete
// +kubebuilder:rbac:groups=apps,resourceNames=kogito-operator,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=image.openshift.io,resources=*,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=*,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=*,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=apps.openshift.io,resources=*,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=core,resources=*,verbs=create;delete;get;list;patch;update;watch

// Reconcile reads that state of the cluster for a KogitoSupportingService object and makes changes based on the state read
// and what is in the KogitoSupportingService.Spec
func (r *KogitoSupportingServiceReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, resultErr error) {
	// Fetch the KogitoSupportingService instance
	r.Log.Info("Reconciling KogitoSupportingService for", "Instance", req.Name, "Namespace", req.Namespace)
	instance, resultErr := infrastructure.FetchKogitoSupportingService(r.Client, req.Name, req.Namespace)
	if resultErr != nil {
		return
	}
	if instance == nil {
		r.Log.Debug("Instance not found", "kogitoSupportingService", req.Name, "Namespace", req.Namespace)
		return
	}

	r.Log.Debug("Going to reconcile service", "Type", instance.Spec.ServiceType)
	if resultErr = ensureSingletonService(r.Client, req.Namespace, instance.Spec.ServiceType); resultErr != nil {
		return
	}

	r.Log.Debug("going to fetch related kogito supporting Service resource", "Instance", instance.Name, "Type", instance.Spec.ServiceType)
	supportingResource := r.getKogitoSupportingServices(instance)[instance.Spec.ServiceType]

	requeueAfter, resultErr := supportingResource.Reconcile(r.Client, instance, r.Scheme)
	if resultErr != nil {
		return
	}

	if requeueAfter > 0 {
		r.Log.Info("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = requeueAfter
		result.Requeue = true
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
	b.Watches(&source.Kind{Type: &appv1beta1.KogitoInfra{}}, infraHandler)

	if r.IsOpenshift() {
		b.Owns(&routev1.Route{}).Owns(&imgv1.ImageStream{})
	}
	return b.Complete(r)
}

// fetches all the supported services managed by kogitoSupportingService controller
func (r *KogitoSupportingServiceReconciler) getKogitoSupportingServices(instance *appv1beta1.KogitoSupportingService) map[appv1beta1.ServiceType]SupportingServiceResource {
	return map[appv1beta1.ServiceType]SupportingServiceResource{
		appv1beta1.DataIndex:      &dataIndexSupportingServiceResource{log: logger.GetLogger("data-index")},
		appv1beta1.Explainability: &explainabilitySupportingServiceResource{log: logger.GetLogger("explainability")},
		appv1beta1.JobsService:    &jobsServiceSupportingServiceResource{log: logger.GetLogger("jobs-service")},
		appv1beta1.MgmtConsole:    &mgmtConsoleSupportingServiceResource{log: logger.GetLogger("mgmt-console")},
		appv1beta1.TaskConsole:    &taskConsoleSupportingServiceResource{log: logger.GetLogger("task-console")},
		appv1beta1.TrustyAI:       &trustyAISupportingServiceResource{log: logger.GetLogger("trusty-AI")},
		appv1beta1.TrustyUI:       &trustyUISupportingServiceResource{log: logger.GetLogger("trusty-UI")},
	}
}

func ensureSingletonService(client *client.Client, namespace string, resourceType appv1beta1.ServiceType) error {
	supportingServiceList := &appv1beta1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return err
	}

	var kogitoSupportingService []appv1beta1.KogitoSupportingService
	for _, service := range supportingServiceList.Items {
		if service.Spec.ServiceType == resourceType {
			kogitoSupportingService = append(kogitoSupportingService, service)
		}
	}

	if len(kogitoSupportingService) > 1 {
		return fmt.Errorf("kogito Supporting Service(%s) already exists, please delete the duplicate before proceeding", resourceType)
	}
	return nil
}

// Check is the testService is available in the slice of allServices
func contains(allServices []appv1beta1.ServiceType, testService appv1beta1.ServiceType) bool {
	for _, a := range allServices {
		if a == testService {
			return true
		}
	}
	return false
}

// SupportingServiceResource Interface to represent type of kogito supporting service resources like JobsService & MgmtConcole
type SupportingServiceResource interface {
	Reconcile(client *client.Client, instance *appv1beta1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, resultErr error)
}
