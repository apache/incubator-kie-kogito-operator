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

package trusty

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

var log = logger.GetLogger("controller_kogitotrusty")

// Add creates a new KogitoDataIndex Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoTrusty{
		client: client.NewForController(mgr.GetConfig(), mgr.GetClient()),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoTrusty controller")
	// Create a new controller
	c, err := controller.New("kogitotrusty-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoDataIndex
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoTrusty{}}, &handler.EnqueueRequestForObject{})
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
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.ConfigMap{}}},
	}
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoTrusty).client, mgr, c, &appv1alpha1.KogitoTrusty{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKogitoTrusty implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoTrusty{}

// ReconcileKogitoTrusty reconciles a KogitoTrusty object
type ReconcileKogitoTrusty struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoTrusty object and makes changes based on the state read
// and what is in the KogitoTrusty.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoTrusty) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoTrusty")

	reqLogger.Infof("Injecting Trusty Index URL into KogitoApps in the namespace '%s'", request.Namespace)
	if err := infrastructure.InjectTrustyURLIntoKogitoApps(r.client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	instances := &appv1alpha1.KogitoTrustyList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultTrustyImageName,
		Request:             request,
		OnDeploymentCreate:  r.onDeploymentCreate,
		KafkaTopics:         kafkaTopics,
		RequiresPersistence: true,
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
	defaultProtobufMountPath                  = "/home/kogito/data/protobufs"
	protoBufKeyFolder                  string = "KOGITO_PROTOBUF_FOLDER"
	protoBufKeyWatch                   string = "KOGITO_PROTOBUF_WATCH"
	protoBufConfigMapVolumeDefaultMode int32  = 420

	// Collection of kafka topics that should be handled by the Trusty
	kafkaTopicNameTraceEvents string = "kogito-jobs-events"
)

var kafkaTopics = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameTraceEvents, MessagingType: services.KafkaTopicIncoming},
}

func (r *ReconcileKogitoTrusty) onDeploymentCreate(deployment *appsv1.Deployment, kogitoService appv1alpha1.KogitoService) error {
	return nil
}
