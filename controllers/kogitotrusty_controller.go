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
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// KogitoTrustyReconciler reconciles a KogitoTrusty object
type KogitoTrustyReconciler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
}

func (r *KogitoTrustyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoTrusty controller")
	// Create a new controller
	c, err := controller.New("kogitotrusty-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoTrusty
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoTrusty{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KogitoTrusty
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
	controllerWatcher := framework.NewControllerWatcher(r.Client, mgr, c, &appv1alpha1.KogitoTrusty{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKogitoTrusty implements reconcile.Reconciler
var _ reconcile.Reconciler = &KogitoTrustyReconciler{}

// Reconcile reads that state of the cluster for a KogitoTrusty object and makes changes based on the state read
// and what is in the KogitoTrusty.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *KogitoTrustyReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := r.Log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoTrusty")

	reqLogger.Infof("Injecting Trusty Index URL into KogitoApps in the namespace '%s'", request.Namespace)
	if err := infrastructure.InjectTrustyURLIntoKogitoApps(r.Client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	instances := &appv1alpha1.KogitoTrustyList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultTrustyImageName,
		Request:             request,
		KafkaTopics:         kafkaTopics,
		RequiresPersistence: true,
		RequiresMessaging:   true,
		HealthCheckProbe:    services.QuarkusHealthCheckProbe,
	}
	if requeueAfter, err := services.NewSingletonServiceDeployer(definition, instances, r.Client, r.Scheme).Deploy(); err != nil {
		return reconcile.Result{}, err
	} else if requeueAfter > 0 {
		return reconcile.Result{RequeueAfter: requeueAfter, Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

const (
	// Collection of kafka topics that should be handled by the Trusty
	kafkaTopicNameTraceEvents                 string = "kogito-tracing-decision"
	kafkaTopicNameModelEvents                 string = "kogito-tracing-model"
	kafkaTopicNameExplainabilityResultEvents  string = "trusty-explainability-result"
	kafkaTopicNameExplainabilityRequestEvents string = "trusty-explainability-request"
)

var kafkaTopics = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameTraceEvents, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameModelEvents, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameExplainabilityResultEvents, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameExplainabilityRequestEvents, MessagingType: services.KafkaTopicOutgoing},
}
