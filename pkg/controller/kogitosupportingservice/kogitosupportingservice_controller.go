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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("kogitosupportingservice_controller")

// Add creates a new KogitoSupportingService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoSupportingService{client: client.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoSupportingService controller")
	// Create a new controller
	c, err := controller.New("kogitosupportingservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1beta1.KogitoSupportingService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoSupportingService).client, mgr, c, &v1beta1.KogitoSupportingService{})
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
		{
			Objects:      []runtime.Object{&v1beta1.KogitoInfra{}},
			EventHandler: &handler.EnqueueRequestForOwner{IsController: false, OwnerType: &v1beta1.KogitoSupportingService{}},
			Predicate: predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return isKogitoInfraUpdated(e.ObjectOld.(*v1beta1.KogitoInfra), e.ObjectNew.(*v1beta1.KogitoInfra))
				},
			},
		},
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.ConfigMap{}}},
	}
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoSupportingService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoSupportingService{}

// ReconcileKogitoSupportingService reconciles a KogitoSupportingService object
type ReconcileKogitoSupportingService struct {
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
func (r *ReconcileKogitoSupportingService) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	// Fetch the KogitoSupportingService instance
	log.Infof("Reconciling KogitoSupportingService for %s in %s", request.Name, request.Namespace)
	instance, resultErr := fetchKogitoSupportingService(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return
	}
	if instance == nil {
		log.Debugf("KogitoSupportingService instance with name %s not found in namespace %s. Going to return reconciliation request", request.Name, request.Namespace)
		return
	}

	log.Debugf("Going to reconcile service of type %s", instance.Spec.ServiceType)
	if resultErr = ensureSingletonService(r.client, request.Namespace, instance.Spec.ServiceType); resultErr != nil {
		return
	}

	supportingResource := getKogitoSupportingService(instance)

	requeueAfter, resultErr := supportingResource.Reconcile(r.client, instance, r.scheme)
	if resultErr != nil {
		return
	}

	if requeueAfter > 0 {
		log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = requeueAfter
		result.Requeue = true
	}
	return
}

func fetchKogitoSupportingService(client *client.Client, name string, namespace string) (*v1beta1.KogitoSupportingService, error) {
	log.Debugf("going to fetch deployed kogito supporting service instance %s in namespace %s", name, namespace)
	instance := &v1beta1.KogitoSupportingService{}
	if exists, resultErr := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, instance); resultErr != nil {
		log.Errorf("Error occurs while fetching deployed kogito supporting service instance %s", name)
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		log.Debugf("Successfully fetch deployed kogito supporting reference %s", name)
		return instance, nil
	}
}

// fetches all the supported services managed by kogitoSupportingService controller
func getKogitoSupportingService(instance *v1beta1.KogitoSupportingService) SupportingServiceResource {
	log.Debugf("going to fetch related kogito supporting Service resource for given instance : %s of type: %s", instance.Name, instance.Spec.ServiceType)
	return kogitoSupportingServices[instance.Spec.ServiceType]
}

func ensureSingletonService(client *client.Client, namespace string, resourceType v1beta1.ServiceType) error {
	supportingServiceList := &v1beta1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return err
	}

	var kogitoSupportingService []v1beta1.KogitoSupportingService
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
func contains(allServices []v1beta1.ServiceType, testService v1beta1.ServiceType) bool {
	for _, a := range allServices {
		if a == testService {
			return true
		}
	}
	return false
}

func isKogitoInfraUpdated(oldKogitoInfra, newKogitoInfra *v1beta1.KogitoInfra) bool {
	return !reflect.DeepEqual(oldKogitoInfra.Status.AppProps, newKogitoInfra.Status.AppProps) || !reflect.DeepEqual(oldKogitoInfra.Status.Env, newKogitoInfra.Status.Env)
}

// SupportingServiceResource Interface to represent type of kogito supporting service resources like JobsService & MgmtConcole
type SupportingServiceResource interface {
	Reconcile(client *client.Client, instance *v1beta1.KogitoSupportingService, scheme *runtime.Scheme) (reconcileAfter time.Duration, resultErr error)
}

// map of all the kogitoSupportingService
// Note: Data Index is not part of this map because it has it's own controller
var kogitoSupportingServices = map[v1beta1.ServiceType]SupportingServiceResource{
	v1beta1.DataIndex:      &DataIndexSupportingServiceResource{},
	v1beta1.Explainability: &ExplainabilitySupportingServiceResource{},
	v1beta1.JobsService:    &JobsServiceSupportingServiceResource{},
	v1beta1.MgmtConsole:    &MgmtConsoleSupportingServiceResource{},
	v1beta1.TaskConsole:    &TaskConsoleSupportingServiceResource{},
	v1beta1.TrustyAI:       &TrustyAISupportingServiceResource{},
	v1beta1.TrustyUI:       &TrustyUISupportingServiceResource{},
}
