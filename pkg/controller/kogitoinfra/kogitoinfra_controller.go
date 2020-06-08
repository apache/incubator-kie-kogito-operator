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
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("kogitoinfra_controller")

var watchedObjects = []framework.WatchedObjects{
	{
		GroupVersion: infinispanv1.SchemeGroupVersion,
		AddToScheme:  infinispanv1.AddToScheme,
		Objects:      []runtime.Object{&infinispanv1.Infinispan{}},
	},
	{
		GroupVersion: keycloakv1alpha1.SchemeGroupVersion,
		AddToScheme:  keycloakv1alpha1.SchemeBuilder.AddToScheme,
		Objects:      []runtime.Object{&keycloakv1alpha1.Keycloak{}},
	},
	{
		GroupVersion: kafkav1beta1.SchemeGroupVersion,
		AddToScheme:  kafkav1beta1.SchemeBuilder.AddToScheme,
		Objects:      []runtime.Object{&kafkav1beta1.Kafka{}},
	},
	{
		Objects: []runtime.Object{&corev1.Secret{}},
	},
}

var controllerWatcher framework.ControllerWatcher

// Add creates a new KogitoInfra Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoInfra{client: kogitocli.NewForController(mgr.GetConfig(), mgr.GetClient()), scheme: mgr.GetScheme()}
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
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoInfra{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	controllerWatcher = framework.NewControllerWatcher(r.(*ReconcileKogitoInfra).client, mgr, c, &appv1alpha1.KogitoInfra{})
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
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoInfra) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoInfra for %s in %s", request.Name, request.Namespace)

	// Requires only one KogitoInfra instance in this namespace
	instances := &appv1alpha1.KogitoInfraList{}
	if err := kubernetes.ResourceC(r.client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}
	if len(instances.Items) > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(5) * time.Minute},
			fmt.Errorf("There's more than one KogitoInfra resource in this namespace, please delete one of them ")
	}

	// Fetch the KogitoInfra instance
	instance := &appv1alpha1.KogitoInfra{}
	if exists, err := kubernetes.ResourceC(r.client).FetchWithKey(request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, err
	} else if !exists {
		return reconcile.Result{}, nil
	}

	// watcher will be nil on test env
	if controllerWatcher != nil && !controllerWatcher.AreAllObjectsWatched() {
		if (instance.Spec.InstallKafka && !controllerWatcher.IsGroupWatched(kafkav1beta1.SchemeGroupVersion.Group)) ||
			(instance.Spec.InstallInfinispan && !controllerWatcher.IsGroupWatched(infinispanv1.SchemeGroupVersion.Group)) ||
			(instance.Spec.InstallKeycloak && !controllerWatcher.IsGroupWatched(keycloakv1alpha1.SchemeGroupVersion.Group)) {
			// try to add them
			if err := controllerWatcher.Watch(watchedObjects...); err != nil {
				return reconcile.Result{}, err
			}
			if !controllerWatcher.AreAllObjectsWatched() {
				log.Warn("Dependencies not found for KogitoInfra, please install them.")
				return reconcile.Result{RequeueAfter: time.Minute * 5}, nil
			}
		}
	}

	defer r.updateBaseStatus(instance, &resultErr)

	// Verify Infinispan
	infinispanAvailable, resultErr := infrastructure.IsInfinispanOperatorAvailable(r.client, instance.Namespace)
	if resultErr != nil {
		return
	}
	if instance.Spec.InstallInfinispan && !infinispanAvailable {
		resultErr = fmt.Errorf("Infinispan is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	// Verify Kafka
	if instance.Spec.InstallKafka && !infrastructure.IsStrimziAvailable(r.client) {
		resultErr = fmt.Errorf("Kafka is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	// Verify Keycloak
	if instance.Spec.InstallKeycloak && !infrastructure.IsKeycloakAvailable(r.client) {
		resultErr = fmt.Errorf("Keycloak is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	requestedResources, resultErr := r.createRequiredResources(instance)
	if resultErr != nil {
		return
	}
	deployedResources, resultErr := r.getDeployedResources(instance)
	if resultErr != nil {
		return
	}
	comparator := r.getComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)

	writer := write.New(r.client.ControlCli).WithOwnerController(instance, r.scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
		_, resultErr = writer.AddResources(delta.Added)
		if resultErr != nil {
			return
		}
		_, resultErr = writer.UpdateResources(deployedResources[resourceType], delta.Updated)
		if resultErr != nil {
			return
		}
		_, resultErr = writer.RemoveResources(delta.Removed)
		if resultErr != nil {
			return
		}
	}

	if result.Requeue, resultErr = status.ManageDependenciesStatus(instance, r.client); resultErr != nil {
		return
	}

	if result.Requeue {
		log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = time.Second * 30
	}

	return
}

// updateBaseStatus updates the base status for the KogitoInfra instance
func (r *ReconcileKogitoInfra) updateBaseStatus(instance *appv1alpha1.KogitoInfra, err *error) {
	log.Info("Updating Kogito Infra status")
	if *err != nil {
		log.Warn("Seems that an error occurred, setting failure state: ", *err)
		if statusErr := status.SetResourceFailed(instance, r.client, *err); statusErr != nil {
			err = &statusErr
			log.Errorf("Error in setting status failes: %v", *err)
		}
	} else {
		log.Info("Kogito Infra successfully reconciled")
		if statusErr := status.SetResourceSuccess(instance, r.client); statusErr != nil {
			err = &statusErr
			log.Errorf("Error in setting status failes: %v", *err)
		}
	}
}
