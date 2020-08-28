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
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
)

// KogitoMgmtConsoleReconciler reconciles a KogitoMgmtConsole object
type KogitoMgmtConsoleReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
}

func (r *KogitoMgmtConsoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoMgmtConsole controller")
	// Create a new controller
	c, err := controller.New("kogitomgmtconsole-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoMgmtConsole
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoMgmtConsole{}}, &handler.EnqueueRequestForObject{})
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
			Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}},
		},
		{
			GroupVersion: routev1.GroupVersion,
			AddToScheme:  routev1.Install,
			Objects:      []runtime.Object{&routev1.Route{}},
			Owner:        &appv1alpha1.KogitoDataIndex{},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}},
			Owner:   &appv1alpha1.KogitoDataIndex{},
		},
	}
	controllerWatcher := framework.NewControllerWatcher(r.Client, mgr, c, &appv1alpha1.KogitoMgmtConsole{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoMgmtConsole implements reconcile.Reconciler
var _ reconcile.Reconciler = &KogitoMgmtConsoleReconciler{}

// Reconcile reads that state of the cluster for a KogitoMgmtConsole object and makes changes based on the state read
// and what is in the KogitoMgmtConsole.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *KogitoMgmtConsoleReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.Log.Infof("Reconciling KogitoMgmtConsole for %s in %s", request.Name, request.Namespace)
	definition := services.ServiceDefinition{
		DefaultImageName:  infrastructure.DefaultMgmtConsoleImageName,
		Request:           request,
		SingleReplica:     false,
		RequiresDataIndex: true,
	}
	if requeueAfter, err := services.NewSingletonServiceDeployer(definition, &appv1alpha1.KogitoMgmtConsoleList{}, r.Client, r.Scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}
