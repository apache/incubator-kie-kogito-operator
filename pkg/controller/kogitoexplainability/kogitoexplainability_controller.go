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

package kogitoexplainability

import (
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
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
)

var log = logger.GetLogger("controller_kogitoexplainability")

// Add creates a new KogitoExplainability Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoExplainability{
		client: client.NewForController(mgr.GetConfig()),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoExplainability controller")
	// Create a new controller
	c, err := controller.New("kogitoexplainability-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoExplainability
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoExplainability{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KogitoExplainability
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.KogitoTrusty{},
	})
	if err != nil {
		return err
	}

	// We also watch for any resources regarding infra to recreate it in case is deleted and we depend on them
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appv1alpha1.KogitoInfra{}})
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
			GroupVersion: imgv1.GroupVersion,
			AddToScheme:  imgv1.Install,
			Objects:      []runtime.Object{&imgv1.ImageStream{}},
		},
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}}},
	}
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoExplainability).client, mgr, c, &appv1alpha1.KogitoExplainability{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKogitoExplainability implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoExplainability{}

// ReconcileKogitoExplainability reconciles a KogitoExplainability object
type ReconcileKogitoExplainability struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoExplainability object and makes changes based on the state read
// and what is in the KogitoExplainability.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoExplainability) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoExplainability")

	instances := &appv1alpha1.KogitoExplainabilityList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultExplainabilityImageName,
		Request:             request,
		KafkaTopics:         kafkaTopics,
		RequiresPersistence: false,
		RequiresMessaging:   true,
		HealthCheckProbe:    services.QuarkusHealthCheckProbe,
	}
	if requeueAfter, err := services.NewSingletonServiceDeployer(definition, instances, r.client, r.scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

const (
	// Collection of kafka topics that should be handled by the Explainability service
	kafkaTopicNameExplainabilityRequest string = "trusty-explainability-request"
	kafkaTopicNameExplainabilityResult  string = "trusty-explainability-result"
)

var kafkaTopics = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameExplainabilityRequest, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameExplainabilityResult, MessagingType: services.KafkaTopicOutgoing},
}
