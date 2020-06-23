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
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
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
	return &ReconcileKogitoRuntime{client: kogitocli.NewForController(mgr.GetConfig(), mgr.GetClient()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoRuntime controller")
	// Create a new controller
	c, err := controller.New("kogitoruntime-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoRuntime
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoRuntime{}}, &handler.EnqueueRequestForObject{})
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
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoRuntime).client, mgr, c, &appv1alpha1.KogitoRuntime{})
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
func (r *ReconcileKogitoRuntime) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconciling KogitoRuntime for %s in %s", request.Name, request.Namespace)
	definition := services.ServiceDefinition{
		Request:            request,
		DefaultImageTag:    infrastructure.LatestTag,
		SingleReplica:      false,
		OnDeploymentCreate: onDeploymentCreate,
		OnObjectsCreate:    onObjectsCreate,
		OnGetComparators:   onGetComparators,
	}
	if requeueAfter, err := services.NewCustomServiceDeployer(definition, &appv1alpha1.KogitoRuntime{}, r.client, r.scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
