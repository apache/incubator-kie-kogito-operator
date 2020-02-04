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
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"

	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"

	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	utilsres "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
)

var log = logger.GetLogger("controller_kogitodataindex")

// Add creates a new KogitoDataIndex Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoDataIndex{
		client: client.NewForController(mgr.GetConfig(), mgr.GetClient()),
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
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appv1alpha1.KogitoApp{}})
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
		&kafkabetav1.KafkaTopic{},
		&imgv1.ImageStream{},
		&keycloakv1alpha1.KeycloakUser{},
		&keycloakv1alpha1.KeycloakClient{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.KogitoDataIndex{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			if framework.IsNoKindMatchError(kafkabetav1.SchemeGroupVersion.Group, err) ||
				framework.IsNoKindMatchError(keycloakv1alpha1.SchemeGroupVersion.Group, err) {
				log.Warn("Tried to watch Kafka CRD, but failed. Maybe related Operators are not installed?")
				continue
			}
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
func (r *ReconcileKogitoDataIndex) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoDataIndex")

	// TODO: move to finalizers the exclusion use case
	// If it's an exclusion, the Data Index won't exist anymore. Routes need to be cleaned.
	reqLogger.Infof("Injecting Data Index URL into KogitoApps in the namespace '%s'", request.Namespace)
	if err := infrastructure.InjectDataIndexURLIntoKogitoApps(r.client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	// Requires only one Data Index.
	// The request might be coming from another source within this namespace, and since we only have one deployment, we
	// can safely use this instance during the reconciliation phase
	instances := &appv1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(r.client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}
	instancesCount := len(instances.Items)
	if instancesCount > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(5) * time.Minute},
			fmt.Errorf("There's more than one KogitoDataIndex resource in %s namespace, please delete one of them ", request.Namespace)
	} else if instancesCount == 0 {
		return reconcile.Result{}, nil
	}

	// Fetch the KogitoDataIndex instance
	instance := &instances.Items[0]

	// Deploy infra dependencies
	if result, err := r.ensureKogitoInfra(instance); err != nil {
		return reconcile.Result{}, err
	} else if result != nil {
		return *result, nil
	}

	var hasUpdates bool

	// Create our inventory
	reqLogger.Infof("Ensure Kogito Data Index '%s' resources are created", instance.Name)
	resources, err := resource.GetRequestedResources(instance, r.client)

	result = reconcile.Result{}

	defer r.updateStatus(&request, instance, resources, &hasUpdates, &resultErr)

	if err != nil {
		resultErr = err
		return
	}

	deployedResources, err := resource.GetDeployedResources(instance, r.client)
	if err != nil {
		resultErr = err
		return
	}

	requestedResources := compare.NewMapBuilder().Add(getKubernetesResources(resources)...).ResourceMap()

	comparator := resource.GetComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)

	writer := write.New(r.client.ControlCli).WithOwnerController(instance, r.scheme)

	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
		added, err := writer.AddResources(delta.Added)
		if err != nil {
			resultErr = err
			return
		}
		updated, err := writer.UpdateResources(deployedResources[resourceType], delta.Updated)
		if err != nil {
			resultErr = err
			return
		}
		removed, err := writer.RemoveResources(delta.Removed)
		if err != nil {
			resultErr = err
			return
		}
		hasUpdates = hasUpdates || added || updated || removed
	}

	return
}

// ensureKogitoInfra will deploy a new Kogito Infra if needed based on Data Index instance requirements.
// returns result not nil if needs reconciliation
func (r *ReconcileKogitoDataIndex) ensureKogitoInfra(instance *appv1alpha1.KogitoDataIndex) (result *reconcile.Result, err error) {
	log.Debug("Verify if we need to deploy Infinispan")

	var updateForInfinispan, updateForKafka, updateForKeycloak bool
	var requeueForInfinispan, requeueForKafka, requeueForKeycloak time.Duration

	if updateForInfinispan, requeueForInfinispan, err = infrastructure.DeployInfinispanWithKogitoInfra(&instance.Spec, instance.Namespace, r.client); err != nil {
		return nil, err
	}

	if updateForKafka, requeueForKafka, err = infrastructure.DeployKafkaWithKogitoInfra(&instance.Spec, instance.Namespace, r.client); err != nil {
		return nil, err
	}

	if instance.Spec.EnableSecurity {
		if updateForKeycloak, requeueForKeycloak, err = infrastructure.DeployKeycloakWithKogitoInfra(&instance.Spec, instance.Namespace, r.client); err != nil {
			return nil, err
		}
	}

	if updateForInfinispan || updateForKafka || updateForKeycloak {
		if err := kubernetes.ResourceC(r.client).Update(instance); err != nil {
			return nil, err
		}
		return &reconcile.Result{}, nil
	} else if requeueForInfinispan > 0 {
		return &reconcile.Result{RequeueAfter: requeueForInfinispan}, nil
	} else if requeueForKafka > 0 {
		return &reconcile.Result{RequeueAfter: requeueForKafka}, nil
	} else if requeueForKeycloak > 0 {
		return &reconcile.Result{RequeueAfter: requeueForKeycloak}, nil
	}

	return nil, nil
}

func getKubernetesResources(resources *resource.KogitoDataIndexResources) []utilsres.KubernetesResource {
	var k8sRes []utilsres.KubernetesResource

	if resources.Deployment != nil {
		k8sRes = append(k8sRes, resources.Deployment)
	}
	if resources.Service != nil {
		k8sRes = append(k8sRes, resources.Service)
	}
	if resources.Route != nil {
		k8sRes = append(k8sRes, resources.Route)
	}
	if resources.KafkaTopics != nil && len(resources.KafkaTopics) > 0 {
		for _, r := range resources.KafkaTopics {
			k8sRes = append(k8sRes, r)
		}
	}
	if resources.ImageStream != nil {
		k8sRes = append(k8sRes, resources.ImageStream)
	}
	if resources.KeycloakUsers != nil && len(resources.KeycloakUsers) > 0 {
		for _, u := range resources.KeycloakUsers {
			k8sRes = append(k8sRes, u)
		}
	}
	if resources.KeycloakClients != nil && len(resources.KeycloakClients) > 0 {
		for _, c := range resources.KeycloakClients {
			k8sRes = append(k8sRes, c)
		}
	}

	return k8sRes
}

func (r *ReconcileKogitoDataIndex) updateStatus(request *reconcile.Request, instance *appv1alpha1.KogitoDataIndex,
	resources *resource.KogitoDataIndexResources, resourcesUpdate *bool, err *error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	reconcileError := *err != nil

	reqLogger.Infof("Handling Status updates on '%s'", instance.Name)
	if errStatus := status.ManageStatus(instance, resources, *resourcesUpdate, reconcileError, r.client); errStatus != nil {
		if !reconcileError {
			*err = errStatus
		}
		return
	}
}
