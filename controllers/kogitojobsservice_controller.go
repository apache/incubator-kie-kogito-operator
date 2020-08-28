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
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
)

// KogitoJobsServiceReconciler reconciles a KogitoJobsService object
type KogitoJobsServiceReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
}

func (r *KogitoJobsServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoJobsService controller")
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

	controllerWatcher := framework.NewControllerWatcher(r.Client, mgr, c, &appv1alpha1.KogitoJobsService{})
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
var _ reconcile.Reconciler = &KogitoJobsServiceReconciler{}

// Reconcile reads that state of the cluster for a KogitoJobsService object and makes changes based on the state read
// and what is in the KogitoJobsService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *KogitoJobsServiceReconciler) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	r.Log.Infof("Reconciling KogitoJobsService for %s in %s", request.Name, request.Namespace)

	// clean up variables if needed
	if err := infrastructure.InjectJobsServicesURLIntoKogitoRuntimeServices(r.Client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	instances := &appv1alpha1.KogitoJobsServiceList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultJobsServiceImageName,
		Request:             request,
		OnDeploymentCreate:  onDeploymentCreateJobsService,
		SingleReplica:       true,
		RequiresPersistence: false,
		RequiresMessaging:   false,
		HealthCheckProbe:    services.QuarkusHealthCheckProbe,
		KafkaTopics:         kafkaTopicsJobsService,
	}
	if requeueAfter, err := services.NewSingletonServiceDeployer(definition, instances, r.Client, r.Scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

const (
	backOffRetryEnvKey                  = "BACKOFF_RETRY"
	maxIntervalLimitRetryEnvKey         = "MAX_INTERVAL_LIMIT_RETRY"
	backOffRetryDefaultValue            = 1000
	maxIntervalLimitRetryDefaultValue   = 60000
	kafkaTopicNameJobsEventsJobsService = "kogito-job-service-job-status-events"
)

var kafkaTopicsJobsService = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameJobsEventsJobsService, MessagingType: services.KafkaTopicOutgoing},
}

func onDeploymentCreateJobsService(cli *kogitocli.Client, deployment *appsv1.Deployment, service appv1alpha1.KogitoService) error {
	jobService := service.(*appv1alpha1.KogitoJobsService)
	if jobService.Spec.BackOffRetryMillis <= 0 {
		jobService.Spec.BackOffRetryMillis = backOffRetryDefaultValue
	}
	framework.SetEnvVar(backOffRetryEnvKey, strconv.FormatInt(jobService.Spec.BackOffRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])
	if jobService.Spec.MaxIntervalLimitToRetryMillis <= 0 {
		jobService.Spec.MaxIntervalLimitToRetryMillis = maxIntervalLimitRetryDefaultValue
	}
	framework.SetEnvVar(maxIntervalLimitRetryEnvKey, strconv.FormatInt(jobService.Spec.MaxIntervalLimitToRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])

	return nil
}
