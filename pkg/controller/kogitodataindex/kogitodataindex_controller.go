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

package kogitodataindex

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	commonres "github.com/kiegroup/kogito-cloud-operator/pkg/resource"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"

	v1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("controller_kogitodataindex")

// Add creates a new KogitoDataIndex Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	imageClient, err := imagev1.NewForConfig(mgr.GetConfig())
	if err != nil {
		panic(fmt.Sprintf("Error getting image client: %v", err))
	}
	discover, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		panic(fmt.Sprintf("Error getting discovery client: %v", err))
	}

	return &ReconcileKogitoDataIndex{
		client: &client.Client{
			ImageCli:   imageClient,
			ControlCli: mgr.GetClient(),
			Discovery:  discover},
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kogitodataindex-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoDataIndex
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoDataIndex{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to KogitoApp since we need their runtime images to check for labels, persistence and so on
	err = c.Watch(&source.Kind{Type: &v1.BuildConfig{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appv1alpha1.KogitoApp{}})
	if err != nil {
		return err
	}

	watchOwnedObjects := []runtime.Object{
		&corev1.ConfigMap{},
		&corev1.Service{},
		&appsv1.StatefulSet{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.KogitoDataIndex{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoDataIndex implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoDataIndex{}

// ReconcileKogitoDataIndex reconciles a KogitoDataIndex object
type ReconcileKogitoDataIndex struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoDataIndex object and makes changes based on the state read
// and what is in the KogitoDataIndex.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoDataIndex) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoDataIndex")

	// If it's an exclusion, the Data Index won't exist anymore. Routes need to be cleaned.
	reqLogger.Infof("Injecting Data Index URL into KogitoApps in the namespace '%s'", request.Namespace)
	if err := infrastructure.InjectDataIndexURLIntoKogitoApps(r.client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	instances := &appv1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(r.client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}

	if len(instances.Items) > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(5) * time.Minute},
			fmt.Errorf("There's more than one KogitoDataIndex resource in this namespace, please delete one of them")
	}

	// Fetch the KogitoDataIndex instance
	instance := &appv1alpha1.KogitoDataIndex{}
	if exists, err := kubernetes.ResourceC(r.client).FetchWithKey(request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, err
	} else if !exists {
		return reconcile.Result{}, nil
	}

	// Create our inventory
	reqLogger.Infof("Ensure Kogito Data Index '%s' resources are created", instance.Name)
	resources, err := resource.CreateOrFetchResources(instance, commonres.FactoryContext{
		Client: r.client,
		PreCreate: func(object meta.ResourceObject) error {
			if object != nil {
				return controllerutil.SetControllerReference(instance, object, r.scheme)
			}
			return nil
		},
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	if !resources.StatefulSetStatus.New {
		reqLogger.Infof("Handling changes in Kogito Data Index '%s'", instance.Name)
		if err = resource.ManageResources(instance, &resources, r.client); err != nil {
			return reconcile.Result{}, err
		}
	}

	reqLogger.Infof("Handling Status updates on '%s'", instance.Name)
	if err = status.ManageStatus(instance, &resources, r.client); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
