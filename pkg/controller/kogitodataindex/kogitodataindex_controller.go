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
	"path"

	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	log.Debug("Adding watched objects for KogitoDataIndex controller")
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

	// Watch for changes to Kogito Runtime since we need their runtime images to check for labels, persistence and so on
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &appv1alpha1.KogitoRuntime{}})
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
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoDataIndex).client, mgr, c, &appv1alpha1.KogitoDataIndex{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
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

	reqLogger.Infof("Injecting Data Index URL into KogitoRuntime services in the namespace '%s'", request.Namespace)
	if err := infrastructure.InjectDataIndexURLIntoKogitoRuntimeServices(r.client, request.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	instances := &appv1alpha1.KogitoDataIndexList{}
	definition := services.ServiceDefinition{
		DefaultImageName:    infrastructure.DefaultDataIndexImageName,
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

	// Collection of kafka topics that should be handled by the Data Index
	kafkaTopicNameProcessInstances  string = "kogito-processinstances-events"
	kafkaTopicNameUserTaskInstances string = "kogito-usertaskinstances-events"
	kafkaTopicNameProcessDomain     string = "kogito-processdomain-events"
	kafkaTopicNameUserTaskDomain    string = "kogito-usertaskdomain-events"
	kafkaTopicNameJobsEvents        string = "kogito-jobs-events"
)

var kafkaTopics = []services.KafkaTopicDefinition{
	{TopicName: kafkaTopicNameProcessInstances, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameUserTaskInstances, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameProcessDomain, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameUserTaskDomain, MessagingType: services.KafkaTopicIncoming},
	{TopicName: kafkaTopicNameJobsEvents, MessagingType: services.KafkaTopicIncoming},
}

func (r *ReconcileKogitoDataIndex) onDeploymentCreate(deployment *appsv1.Deployment, kogitoService appv1alpha1.KogitoService) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if err := r.mountProtoBufConfigMaps(deployment); err != nil {
			return err
		}
	} else {
		log.Warnf("No container definition for service %s. Skipping applying custom Data Index deployment configuration", kogitoService.GetName())
	}

	return nil
}

// mountProtoBufConfigMaps mounts protobuf configMaps from KogitoRuntime services into the given deployment
func (r *ReconcileKogitoDataIndex) mountProtoBufConfigMaps(deployment *appsv1.Deployment) (err error) {
	var cms *corev1.ConfigMapList
	configMapDefaultMode := protoBufConfigMapVolumeDefaultMode
	if cms, err = infrastructure.GetProtoBufConfigMaps(deployment.Namespace, r.client); err != nil {
		return err
	}
	for _, cm := range cms.Items {
		deployment.Spec.Template.Spec.Volumes =
			append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: cm.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &configMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cm.Name,
						},
					},
				},
			})
		for fileName := range cm.Data {
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
				append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
					corev1.VolumeMount{Name: cm.Name, MountPath: path.Join(defaultProtobufMountPath, cm.Labels["app"], fileName), SubPath: fileName})
		}
	}

	if len(deployment.Spec.Template.Spec.Volumes) > 0 {
		framework.SetEnvVar(protoBufKeyWatch, "true", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, defaultProtobufMountPath, &deployment.Spec.Template.Spec.Containers[0])
	} else {
		framework.SetEnvVar(protoBufKeyWatch, "false", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, "", &deployment.Spec.Template.Spec.Containers[0])
	}

	return nil
}
