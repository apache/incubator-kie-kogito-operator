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

package kogitojobsservice

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitojobsservice/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitojobsservice/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
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

var log = logger.GetLogger("jobsservice_controller")

// Add creates a new KogitoJobsService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoJobsService{client: kogitocli.NewForController(mgr.GetConfig(), mgr.GetClient()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kogitojobsservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoJobsService
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoJobsService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// We also watch for any resources regarding infra to recreate it in case is deleted and we depend on them
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appv1alpha1.KogitoInfra{}})
	if err != nil {
		return err
	}

	watchOwnedObjects := []runtime.Object{
		&corev1.Service{},
		&appsv1.Deployment{},
		&routev1.Route{},
		&imagev1.ImageStream{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KogitoJobsService{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			// Kubernetes clusters doesn't have routes or imageStream APIs
			if framework.IsNoKindMatchError(routev1.GroupName, err) ||
				framework.IsNoKindMatchError(imagev1.GroupName, err) {
				log.Info("Ignoring specific group to be watched, APIs not found in the current cluster")
				continue
			}
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileKogitoJobsService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoJobsService{}

// ReconcileKogitoJobsService reconciles a KogitoJobsService object
type ReconcileKogitoJobsService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoJobsService object and makes changes based on the state read
// and what is in the KogitoJobsService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoJobsService) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoJobsService for %s in %s", request.Name, request.Namespace)

	// clean up variables if needed
	if err := infrastructure.InjectJobsServicesURLIntoKogitoApps(r.client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	// Requires only one KogitoJobsService instance in this namespace
	instances := &appv1alpha1.KogitoJobsServiceList{}
	if err := kubernetes.ResourceC(r.client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}
	if len(instances.Items) > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(1) * time.Minute},
			fmt.Errorf("There's more than one KogitoJobsService resource in this namespace, please delete one of them ")
	} else if len(instances.Items) == 0 {
		// jobs service being deleted
		return reconcile.Result{}, nil
	}

	// Fetch the KogitoJobsService instance
	instance := &instances.Items[0]
	result = reconcile.Result{}

	defer r.updateStatus(instance, &resultErr)

	requeueAfter, resultErr := r.deployInfinispanIfNeeded(instance)
	if resultErr != nil {
		return
	} else if requeueAfter > 0 {
		result.RequeueAfter = requeueAfter
		result.Requeue = true
		return
	}

	requestedResources, resultErr := resource.CreateRequiredResources(instance, r.client)
	if resultErr != nil {
		return
	}

	deployedResources, resultErr := resource.GetDeployedResources(instance, r.client)
	if resultErr != nil {
		return
	}
	comparator := resource.GetComparator()
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
	return
}

func (r *ReconcileKogitoJobsService) updateStatus(instance *appv1alpha1.KogitoJobsService, err *error) {
	log.Info("Updating status for Job Service")
	if statusErr := status.ManageStatus(instance, r.client, *err); statusErr != nil {
		// this error will return to the operator console
		err = &statusErr
	}
	log.Infof("Successfully reconciled Job Service %s", instance.Name)
}

func (r *ReconcileKogitoJobsService) deployInfinispanIfNeeded(instance *appv1alpha1.KogitoJobsService) (requeueAfter time.Duration, err error) {
	requeueAfter = 0
	if !instance.Spec.InfinispanProperties.UseKogitoInfra {
		return
	}
	update := false
	if update, requeueAfter, err = infrastructure.DeployInfinispanWithKogitoInfra(&instance.Spec, instance.Namespace, r.client); err != nil {
		return
	} else if update {
		if err = kubernetes.ResourceC(r.client).Update(instance); err != nil {
			return
		}
	}

	return
}
