/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/pkg/external/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/services/kogitoinfra"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
)

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

// KogitoInfraReconciler reconciles a KogitoInfra object
type KogitoInfraReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
}

// blank assignment to verify that ReconcileKogitoInfra implements reconcile.Reconciler
var _ reconcile.Reconciler = &KogitoInfraReconciler{}

func (r *KogitoInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoInfra controller")
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

	controllerWatcher = framework.NewControllerWatcher(r.Client, mgr, c, &appv1alpha1.KogitoInfra{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}

	return nil
}

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *KogitoInfraReconciler) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	r.Log.Infof("Reconciling KogitoInfra for %s in %s", request.Name, request.Namespace)

	// Requires only one KogitoInfra instance in this namespace
	instances := &appv1alpha1.KogitoInfraList{}
	if err := kubernetes.ResourceC(r.Client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}
	if len(instances.Items) > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(5) * time.Minute},
			fmt.Errorf("There's more than one KogitoInfra resource in this namespace, please delete one of them ")
	}

	// Fetch the KogitoInfra instance
	instance := &appv1alpha1.KogitoInfra{}
	if exists, err := kubernetes.ResourceC(r.Client).FetchWithKey(request.NamespacedName, instance); err != nil {
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
				r.Log.Warn("Dependencies not found for KogitoInfra, please install them.")
				return reconcile.Result{RequeueAfter: time.Minute * 5}, nil
			}
		}
	}

	defer r.updateBaseStatus(instance, &resultErr)

	// Verify Infinispan
	infinispanAvailable, resultErr := infrastructure.IsInfinispanOperatorAvailable(r.Client, instance.Namespace)
	if resultErr != nil {
		return
	}
	if instance.Spec.InstallInfinispan && !infinispanAvailable {
		resultErr = fmt.Errorf("Infinispan is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	// Verify Kafka
	if instance.Spec.InstallKafka && !infrastructure.IsStrimziAvailable(r.Client) {
		resultErr = fmt.Errorf("Kafka is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	// Verify Keycloak
	if instance.Spec.InstallKeycloak && !infrastructure.IsKeycloakAvailable(r.Client) {
		resultErr = fmt.Errorf("Keycloak is not available in the namespace %s, impossible to continue ", instance.Namespace)
		return
	}

	requestedResources, resultErr := kogitoinfra.CreateRequiredResources(r.Client, instance)
	if resultErr != nil {
		return
	}
	deployedResources, resultErr := kogitoinfra.GetDeployedResources(r.Client, instance)
	if resultErr != nil {
		return
	}
	comparator := kogitoinfra.GetComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)

	writer := write.New(r.Client.ControlCli).WithOwnerController(instance, r.Scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		r.Log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
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

	if result.Requeue, resultErr = kogitoinfra.ManageDependenciesStatus(instance, r.Client); resultErr != nil {
		return
	}

	if result.Requeue {
		r.Log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = time.Second * 30
	}

	return
}

// updateBaseStatus updates the base status for the KogitoInfra instance
func (r *KogitoInfraReconciler) updateBaseStatus(instance *appv1alpha1.KogitoInfra, err *error) {
	r.Log.Info("Updating Kogito Infra status")
	if *err != nil {
		r.Log.Warn("Seems that an error occurred, setting failure state: ", *err)
		if statusErr := kogitoinfra.SetResourceFailed(instance, r.Client, *err); statusErr != nil {
			err = &statusErr
			r.Log.Errorf("Error in setting status failes: %v", *err)
		}
	} else {
		r.Log.Info("Kogito Infra successfully reconciled")
		if statusErr := kogitoinfra.SetResourceSuccess(instance, r.Client); statusErr != nil {
			err = &statusErr
			r.Log.Errorf("Error in setting status failes: %v", *err)
		}
	}
}
