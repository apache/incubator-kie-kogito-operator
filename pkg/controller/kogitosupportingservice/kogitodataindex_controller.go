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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	controller1 "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddDataIndex creates a new KogitoDataIndex Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddDataIndex(mgr manager.Manager) error {
	return addDataIndex(mgr, newDataIndexReconciler(mgr))
}

// newDataIndexReconciler returns a new reconcile.Reconciler
func newDataIndexReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoDataIndex{client: client.NewForController(mgr.GetConfig()), scheme: mgr.GetScheme()}
}

// addDataIndex adds a new Controller to mgr with r as the reconcile.Reconciler
func addDataIndex(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for DataIndex controller")
	// Create a new controller
	c, err := controller.New("DataIndex-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoSupportingService
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			supportingService := e.Object.(*appv1alpha1.KogitoSupportingService)
			return supportingService.Spec.ServiceType == appv1alpha1.DataIndex
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			supportingService := e.ObjectNew.(*appv1alpha1.KogitoSupportingService)
			return supportingService.Spec.ServiceType == appv1alpha1.DataIndex
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoSupportingService{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoDataIndex).client, mgr, c, &appv1alpha1.KogitoSupportingService{})
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
			GroupVersion: corev1.SchemeGroupVersion,
			Objects:      []runtime.Object{&corev1.ConfigMap{}},
			Owner:        &appv1alpha1.KogitoRuntime{},
			// Filter kogitoRuntime Config Map
			Predicate: predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					cm := e.Object.(*corev1.ConfigMap)
					if cm.Labels != nil {
						if cm.Labels[infrastructure.ConfigMapProtoBufEnabledLabelKey] == "true" {
							return true
						}
					}
					return false
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					cm := e.ObjectNew.(*corev1.ConfigMap)
					if cm.Labels != nil {
						if cm.Labels[infrastructure.ConfigMapProtoBufEnabledLabelKey] == "true" {
							return true
						}
					}
					return false
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					cm := e.Object.(*corev1.ConfigMap)
					if cm.Labels != nil {
						if cm.Labels[infrastructure.ConfigMapProtoBufEnabledLabelKey] == "true" {
							return true
						}
					}
					return false
				},
			},
		},
		{
			Objects:      []runtime.Object{&appv1alpha1.KogitoInfra{}},
			Eventhandler: &handler.EnqueueRequestForOwner{IsController: false, OwnerType: &appv1alpha1.KogitoSupportingService{}},
			Predicate: predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					return isKogitoInfraUpdated(e.ObjectOld.(*appv1alpha1.KogitoInfra), e.ObjectNew.(*appv1alpha1.KogitoInfra))
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

// blank assignment to verify that ReconcileKogitoDataIndex implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoDataIndex{}

// ReconcileKogitoDataIndex reconciles a KogitoSupportingService object
type ReconcileKogitoDataIndex struct {
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
func (r *ReconcileKogitoDataIndex) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	instance, resultErr := fetchKogitoSupportingService(r.client, request.Name, request.Namespace)
	if resultErr != nil {
		return
	}
	// Ignore reconcile of other services triggered when change in kogitoInfra object
	if appv1alpha1.DataIndex != instance.Spec.ServiceType {
		return
	}
	log.Infof("Reconciling KogitoDataIndex for %s in %s", request.Name, request.Namespace)
	if resultErr = ensureSingletonService(r.client, request.Namespace, instance.Spec.ServiceType); resultErr != nil {
		return
	}

	log.Infof("Injecting Data Index URL into KogitoRuntime services in the namespace '%s'", instance.Namespace)
	if err := infrastructure.InjectDataIndexURLIntoKogitoRuntimeServices(r.client, instance.Namespace); err != nil {
		return
	}
	if err := infrastructure.InjectDataIndexURLIntoMgmtConsole(r.client, instance.Namespace); err != nil {
		return
	}

	definition := services.ServiceDefinition{
		DefaultImageName:   infrastructure.DefaultDataIndexImageName,
		OnDeploymentCreate: dataIndexOnDeploymentCreate,
		KafkaTopics:        dataIndexKafkaTopics,
		Request:            controller1.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}},
		HealthCheckProbe:   services.QuarkusHealthCheckProbe,
	}
	requeue, resultErr := services.NewServiceDeployer(definition, instance, r.client, r.scheme).Deploy()

	if resultErr != nil {
		return
	}

	if requeue {
		log.Infof("Waiting for all resources to be created, scheduling for 30 seconds from now")
		result.RequeueAfter = time.Second * 30
		result.Requeue = true
	}
	return
}

// Collection of kafka topics that should be handled by the Data-Index service
var dataIndexKafkaTopics = []string{
	"kogito-processinstances-events",
	"kogito-usertaskinstances-events",
	"kogito-processdomain-events",
	"kogito-usertaskdomain-events",
	"kogito-jobs-events",
}

const (
	defaultProtobufMountPath                  = "/home/kogito/data/protobufs"
	protoBufKeyFolder                  string = "KOGITO_PROTOBUF_FOLDER"
	protoBufKeyWatch                   string = "KOGITO_PROTOBUF_WATCH"
	protoBufConfigMapVolumeDefaultMode int32  = 420
)

func dataIndexOnDeploymentCreate(cli *client.Client, deployment *appsv1.Deployment, kogitoService appv1alpha1.KogitoService) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if err := mountProtoBufConfigMaps(deployment, cli); err != nil {
			return err
		}
	} else {
		log.Warnf("No container definition for service %s. Skipping applying custom Data Index deployment configuration", kogitoService.GetName())
	}

	return nil
}

// mountProtoBufConfigMaps mounts protobuf configMaps from KogitoRuntime services into the given deployment
func mountProtoBufConfigMaps(deployment *appsv1.Deployment, client *client.Client) (err error) {
	var cms *corev1.ConfigMapList
	configMapDefaultMode := protoBufConfigMapVolumeDefaultMode
	if cms, err = infrastructure.GetProtoBufConfigMaps(deployment.Namespace, client); err != nil {
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
