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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("kogitoruntime_controller")

// Add creates a new KogitoRuntime Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoRuntime{client: kogitocli.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoRuntime controller")
	// Create a new controller
	c, err := controller.New("kogitoruntime-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1beta1.KogitoRuntime{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchedObjects := []framework.WatchedObjects{
		{
			GroupVersion: routev1.GroupVersion,
			AddToScheme:  routev1.Install,
			Objects:      []runtime.Object{&routev1.Route{}},
		},
		{
			GroupVersion: imagev1.GroupVersion,
			AddToScheme:  imagev1.Install,
			Objects:      []runtime.Object{&imagev1.ImageStream{}},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.ConfigMap{}},
		},
	}
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoRuntime).client, mgr, c, &v1beta1.KogitoRuntime{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoRuntime implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoRuntime{}

// ReconcileKogitoRuntime reconciles a KogitoRuntime object
type ReconcileKogitoRuntime struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoRuntime object and makes changes based on the state read
// and what is in the KogitoRuntime.Spec
func (r *ReconcileKogitoRuntime) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	log.Infof("Reconciling KogitoRuntime for %s in %s", request.Name, request.Namespace)

	instance, err := infrastructure.FetchKogitoRuntimeService(r.client, request.Name, request.Namespace)
	if err != nil {
		return
	}
	if instance == nil {
		log.Debugf("KogitoRuntime instance with name %s not found in namespace %s. Going to return reconciliation request", request.Name, request.Namespace)
		return
	}

	if err = r.setupRBAC(request.Namespace); err != nil {
		return
	}

	if err = infrastructure.MountProtoBufConfigMapOnDataIndex(r.client, instance); err != nil {
		log.Errorf("Fail to mount Proto Buf config map of Kogito runtime service(%s) on DataIndex", instance.Name, err)
		return
	}

	definition := services.ServiceDefinition{
		Request:            request,
		DefaultImageTag:    infrastructure.LatestTag,
		SingleReplica:      false,
		OnDeploymentCreate: onDeploymentCreate,
		OnObjectsCreate:    onObjectsCreate,
		OnGetComparators:   onGetComparators,
		CustomService:      true,
	}
	requeueAfter, err := services.NewServiceDeployer(definition, instance, r.client, r.scheme).Deploy()
	if err != nil {
		return
	}
	if requeueAfter > 0 {
		log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = requeueAfter
		result.Requeue = true
	}
	return
}

func (r *ReconcileKogitoRuntime) setupRBAC(namespace string) (err error) {
	// create service viewer role
	if err = kubernetes.ResourceC(r.client).CreateIfNotExists(getServiceViewerRole(namespace)); err != nil {
		log.Errorf("Fail to create role for service viewer: %v", err)
		return
	}

	// create service viewer service account
	if err = kubernetes.ResourceC(r.client).CreateIfNotExists(getServiceViewerServiceAccount(namespace)); err != nil {
		log.Errorf("Fail to create service account for service viewer: %v", err)
		return
	}

	// create service viewer rolebinding
	if err = kubernetes.ResourceC(r.client).CreateIfNotExists(getServiceViewerRoleBinding(namespace)); err != nil {
		log.Errorf("Fail to create role binding for service viewer: %v", err)
		return
	}
	return
}
