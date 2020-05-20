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
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"strconv"

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
	log.Debug("Adding watched objects for KogitoJobsService controller")
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

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoJobsService).client, mgr, c, &appv1alpha1.KogitoJobsService{})
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
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
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

	instances := &appv1alpha1.KogitoJobsServiceList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultJobsServiceImageName,
		Request:             request,
		OnDeploymentCreate:  onDeploymentCreate,
		SingleReplica:       true,
		RequiresPersistence: false,
		RequiresMessaging:   false,
		HealthCheckProbe:    services.QuarkusHealthCheckProbe,
		KafkaTopics:         kafkaTopics,
	}
	if requeueAfter, err := services.NewSingletonServiceDeployer(definition, instances, r.client, r.scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

const (
	backOffRetryEnvKey                = "BACKOFF_RETRY"
	maxIntervalLimitRetryEnvKey       = "MAX_INTERVAL_LIMIT_RETRY"
	backOffRetryDefaultValue          = 1000
	maxIntervalLimitRetryDefaultValue = 60000
	kafkaTopicNameJobsEvents          = "kogito-job-service-job-status-events"
)

var kafkaTopics = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameJobsEvents, MessagingType: services.KafkaTopicOutgoing},
}

func onDeploymentCreate(deployment *appsv1.Deployment, service appv1alpha1.KogitoService) error {
	jobService := service.(*appv1alpha1.KogitoJobsService)
	if &jobService.Spec.BackOffRetryMillis != nil {
		if jobService.Spec.BackOffRetryMillis <= 0 {
			jobService.Spec.BackOffRetryMillis = backOffRetryDefaultValue
		}
		framework.SetEnvVar(backOffRetryEnvKey, strconv.FormatInt(jobService.Spec.BackOffRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])
	}
	if &jobService.Spec.MaxIntervalLimitToRetryMillis != nil {
		if jobService.Spec.MaxIntervalLimitToRetryMillis <= 0 {
			jobService.Spec.MaxIntervalLimitToRetryMillis = maxIntervalLimitRetryDefaultValue
		}
		framework.SetEnvVar(maxIntervalLimitRetryEnvKey, strconv.FormatInt(jobService.Spec.MaxIntervalLimitToRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])
	}
	return nil
}
