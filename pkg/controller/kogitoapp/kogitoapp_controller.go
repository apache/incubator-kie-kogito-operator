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

package kogitoapp

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/openshift/api/image/docker10"
	"reflect"
	"time"

	utilsres "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	kogitores "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	oappsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("controller_kogitoapp")

// Add creates a new KogitoApp Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoApp{
		client: kogitocli.NewForController(mgr.GetConfig(), mgr.GetClient()),
		scheme: mgr.GetScheme(),
		cache:  mgr.GetCache(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kogitoapp-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoApp
	err = c.Watch(&source.Kind{Type: &v1alpha1.KogitoApp{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = addWatchObjects(c,
		&handler.EnqueueRequestForOwner{IsController: true, OwnerType: &v1alpha1.KogitoApp{}},
		&oappsv1.DeploymentConfig{},
		&corev1.Service{},
		&routev1.Route{},
		&obuildv1.BuildConfig{},
		&oimagev1.ImageStream{},
		&corev1.ConfigMap{})
	if err != nil {
		return err
	}

	return nil
}

// addWatchObjects add an object batch to the watch list
func addWatchObjects(c controller.Controller, eventHandler handler.EventHandler, objects ...runtime.Object) error {
	for _, watchObject := range objects {
		err := c.Watch(&source.Kind{Type: watchObject}, eventHandler)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileKogitoApp{}

// ReconcileKogitoApp reconciles a KogitoApp object
type ReconcileKogitoApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *kogitocli.Client
	scheme *runtime.Scheme
	cache  cachev1.Cache
}

// Reconcile reads that state of the cluster for a KogitoApp object and makes changes based on the state read
// and what is in the KogitoApp.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoApp) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Info("Reconciling KogitoApp")

	// Fetch the KogitoApp instance
	instance := &v1alpha1.KogitoApp{}
	if exists, err := kubernetes.ResourceC(r.client).FetchWithKey(request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, err
	} else if !exists {
		return reconcile.Result{}, nil
	}

	if instance.Spec.Runtime != v1alpha1.SpringbootRuntimeType {
		instance.Spec.Runtime = v1alpha1.QuarkusRuntimeType
	}

	// infra defaults
	if &instance.Spec.Infra == nil {
		instance.Spec.Infra = v1alpha1.KogitoAppInfra{}
	}
	if len(instance.Spec.Infra.InstallKafka) == 0 {
		instance.Spec.Infra.InstallKafka = v1alpha1.KogitoAppInfraInstallKafkaNever
	}
	if len(instance.Spec.Infra.InstallInfinispan) == 0 {
		instance.Spec.Infra.InstallInfinispan = v1alpha1.KogitoAppInfraInstallInfinispanAuto
	}

	requeue, err := r.ensureKogitoImageStream(instance)
	if err != nil {
		return reconcile.Result{}, err
	} else if requeue {
		return reconcile.Result{RequeueAfter: time.Duration(500) * time.Millisecond}, nil
	}

	updateResourceResult := &status.UpdateResourcesResult{ErrorReason: v1alpha1.ReasonType("")}

	log.Infof("Checking if all resources for '%s' are created", instance.Name)
	// create resources in the cluster that do not exist
	kogitoResources, err := kogitores.GetRequestedResources(&kogitores.Context{
		KogitoApp: instance,
		FactoryContext: framework.FactoryContext{
			Client: r.client,
		},
	})

	defer r.updateKogitoAppStatus(&request, instance, kogitoResources, updateResourceResult, &result, &resultErr)

	if err != nil {
		updateResourceResult.Err = err
		updateResourceResult.ErrorReason = v1alpha1.ParseCRRequestFailedReason
		return
	}

	deployedRes, err := kogitores.GetDeployedResources(instance, r.client)
	if err != nil {
		updateResourceResult.Err = err
		updateResourceResult.ErrorReason = v1alpha1.RetrieveDeployedResourceFailedReason
		return
	}

	if len(deployedRes) > 0 {
		if requeue, err = r.ensureKogitoInfra(instance, kogitoResources.RuntimeImage, kogitoResources.DeploymentConfig); err != nil {
			updateResourceResult.Err = err
			updateResourceResult.ErrorReason = v1alpha1.DeployKogitoInfraFailedReason
			return
		} else if requeue {
			result.Requeue = true
			result.RequeueAfter = 5 * time.Second
			return
		}

		if err = r.injectExternalVariables(instance, kogitoResources.DeploymentConfig); err != nil {
			updateResourceResult.Err = err
			updateResourceResult.ErrorReason = v1alpha1.ServicesIntegrationFailedReason
			return
		}
	}

	// Verifies if deployed configMap has integrity, roll out if it does not and let the application update it
	if err := r.rollOutDeploymentIfConfigMapBroken(instance, deployedRes); err != nil {
		updateResourceResult.Err = err
		updateResourceResult.ErrorReason = v1alpha1.RolloutDeploymentFailedReason
		return
	}

	{
		requestedRes := compare.NewMapBuilder().Add(getKubernetesResources(kogitoResources)...).ResourceMap()

		comparator := kogitores.GetComparator()
		deltas := comparator.Compare(deployedRes, requestedRes)

		writer := write.New(r.client.ControlCli).WithOwnerController(instance, r.scheme)

		var hasUpdates bool
		for resourceType, delta := range deltas {
			if !delta.HasChanges() {
				continue
			}
			log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
			added, err := writer.AddResources(delta.Added)
			if err != nil {
				updateResourceResult.Err = err
				updateResourceResult.ErrorReason = v1alpha1.CreateResourceFailedReason
				return
			}
			updated, err := writer.UpdateResources(deployedRes[resourceType], delta.Updated)
			if err != nil {
				updateResourceResult.Err = err
				updateResourceResult.ErrorReason = v1alpha1.UpdateResourceFailedReason
				return
			}
			removed, err := writer.RemoveResources(delta.Removed)
			if err != nil {
				updateResourceResult.Err = err
				updateResourceResult.ErrorReason = v1alpha1.RemoveResourceFailedReason
				return
			}
			hasUpdates = hasUpdates || added || updated || removed
		}

		updateResourceResult.Updated = hasUpdates

		bcDelta := deltas[reflect.TypeOf(obuildv1.BuildConfig{})]
		if bcDelta.HasChanges() {
			var bcs []*obuildv1.BuildConfig
			for _, bc := range bcDelta.Added {
				bcs = append(bcs, bc.(*obuildv1.BuildConfig))
			}
			for _, bc := range bcDelta.Updated {
				bcs = append(bcs, bc.(*obuildv1.BuildConfig))
			}

			if err = r.triggerBuilds(instance, bcs...); err != nil {
				updateResourceResult.Err = err
				updateResourceResult.ErrorReason = v1alpha1.TriggerBuildFailedReason
				return
			}
		}
	}

	return
}

func (r *ReconcileKogitoApp) updateKogitoAppStatus(request *reconcile.Request, instance *v1alpha1.KogitoApp,
	kogitoResources *kogitores.KogitoAppResources, updateResourceResult *status.UpdateResourcesResult, result *reconcile.Result, err *error) {

	log.Infof("Handling Status updates on '%s'", instance.Name)
	statusUpdateResult := status.ManageStatus(instance, kogitoResources, r.client, updateResourceResult, r.cache, request.NamespacedName)

	if statusUpdateResult.Err != nil {
		log.Warnf("Reconcile for '%s' finished with error", instance.Name)
		*err = statusUpdateResult.Err
	} else if statusUpdateResult.RequeueAfter > 0 {
		log.Infof("Reconcile for '%s' finished with requeue in the given time interval", instance.Name)
		result.RequeueAfter = statusUpdateResult.RequeueAfter
	} else if statusUpdateResult.Updated {
		// Updating the KogitoApp will trigger the execution of the reconcile loop, so it is not necessary to requeue the loop
		log.Infof("Reconcile for '%s' finished with the KogitoApp updated", instance.Name)
	} else {
		log.Infof("Reconcile for '%s' successfully finished", instance.Name)
	}
}

func (r *ReconcileKogitoApp) triggerBuilds(instance *v1alpha1.KogitoApp, bcs ...*obuildv1.BuildConfig) error {
	bcMap := map[string][]obuildv1.BuildConfig{}
	for _, bc := range bcs {
		bcMap[bc.Labels[kogitores.LabelKeyBuildType]] = append(bcMap[bc.Labels[kogitores.LabelKeyBuildType]], *bc)
	}

	// Trigger only the S2I builds, they will trigger runtime builds
	if bcMap[string(kogitores.BuildTypeS2I)] != nil {
		for _, bc := range bcMap[string(kogitores.BuildTypeS2I)] {
			log.Infof("Buildconfigs are created, triggering build %s", bc.GetName())
			if _, err := openshift.BuildConfigC(r.client).TriggerBuild(&bc, instance.Name); err != nil {
				return err
			}
		}
	} else {
		for _, bc := range bcs {
			log.Infof("Buildconfigs are created, triggering build %s", bc.GetName())
			if _, err := openshift.BuildConfigC(r.client).TriggerBuild(bc, instance.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Ensure that all Kogito images are available before the build.
func (r *ReconcileKogitoApp) ensureKogitoImageStream(instance *v1alpha1.KogitoApp) (requeue bool, err error) {
	isISPresentOnOpenshiftNamespace := true
	isImagestreamReady := true

	imageStreamTag := kogitores.ImageStreamTag
	if instance.Spec.Build.ImageS2I.ImageStreamTag != "" || instance.Spec.Build.ImageRuntime.ImageStreamTag != "" {
		if instance.Spec.Build.ImageS2I.ImageStreamTag == instance.Spec.Build.ImageRuntime.ImageStreamTag {
			imageStreamTag = instance.Spec.Build.ImageS2I.ImageStreamTag
			log.Infof("Using ImageS2I and RuntimeImage Tag from CR: %s", imageStreamTag)
		} else {
			log.Infof("ImageS2I and ImageRuntime are set and are different, using the default tag %s.", imageStreamTag)
		}
	}

	kogitoRequiredIS := kogitores.KogitoImageStream(instance.Namespace, imageStreamTag, instance.Spec.Runtime, instance.Spec.Build.Native)

	// First look on the current namespace, if the imagestreams are available, return
	// and don't check for anything else.
	// look for the Kogito image stream on the default namespace (openshift)
	// All needed images must be present, if not create it on current namespace
	for _, is := range kogitoRequiredIS.Items {
		isCurrenctNs, _ := openshift.ImageStreamC(r.client).FetchTag(types.NamespacedName{Name: is.Name, Namespace: instance.Namespace}, imageStreamTag)
		isOcpNs, _ := openshift.ImageStreamC(r.client).FetchTag(types.NamespacedName{Name: is.Name, Namespace: kogitores.ImageStreamNamespace}, imageStreamTag)
		if isCurrenctNs == nil {
			isImagestreamReady = false
		}
		if isOcpNs == nil {
			isISPresentOnOpenshiftNamespace = false
		}
	}

	if !isISPresentOnOpenshiftNamespace && !isImagestreamReady {

		for _, is := range kogitoRequiredIS.Items {
			is.Namespace = instance.Namespace
			hasIs, err := openshift.ImageStreamC(r.client).FetchTag(types.NamespacedName{Name: is.Name, Namespace: instance.Namespace}, imageStreamTag)
			if err != nil {
				log.Error(err.Error())
			}

			if hasIs == nil {
				_, err := openshift.ImageStreamC(r.client).CreateImageStream(&is)
				if err != nil {
					return false, err
				}

				tagRef := fmt.Sprintf("%s:%s", is.Name, imageStreamTag)
				imageTag, getISerr := openshift.ImageStreamC(r.client).FetchTag(types.NamespacedName{Name: is.Name, Namespace: is.Namespace}, imageStreamTag)
				if getISerr != nil {
					log.Error(getISerr)
				}
				if imageTag != nil && imageTag.Image.Name != "" {
					log.Infof("ImageStream %s ready.", imageTag.Name)
				} else {
					log.Warnf("ImageStream %s not ready/found, scheduling reconcile", tagRef)
					return true, nil
				}
			}
		}

		//set the isImagestreamReady to true to avoid it to be checked every time the reconcile runs
	} else if isISPresentOnOpenshiftNamespace {
		// if all imagestreams are present on the openshift namespace, set it as default
		instance.Spec.Build.ImageS2I.ImageStreamNamespace = kogitores.ImageStreamNamespace
		instance.Spec.Build.ImageRuntime.ImageStreamNamespace = kogitores.ImageStreamNamespace
	}
	return false, nil
}

func (r *ReconcileKogitoApp) ensureKogitoInfra(instance *v1alpha1.KogitoApp, runtimeImage *docker10.DockerImage, requestedDeployment *oappsv1.DeploymentConfig) (requeue bool, err error) {
	requeueInfinispan, err := r.ensureInfinispan(instance, runtimeImage, requestedDeployment)
	if err != nil {
		return false, err
	}

	requeueKafka, err := r.ensureKafka(instance, requestedDeployment)
	if err != nil {
		return false, err
	}

	return requeueInfinispan || requeueKafka, nil
}

func (r *ReconcileKogitoApp) ensureInfinispan(instance *v1alpha1.KogitoApp, runtimeImage *docker10.DockerImage, requestedDeployment *oappsv1.DeploymentConfig) (requeue bool, err error) {
	log.Debug("Verify if we need to deploy Infinispan")
	if instance.Spec.Infra.InstallInfinispan == v1alpha1.KogitoAppInfraInstallInfinispanAlways ||
		(instance.Spec.Infra.InstallInfinispan == v1alpha1.KogitoAppInfraInstallInfinispanAuto && framework.IsPersistenceEnabled(runtimeImage)) {
		infra, created, ready, err := infrastructure.EnsureKogitoInfra(instance.Namespace, r.client).WithInfinispan()
		if err != nil {
			return true, err
		}
		if created {
			// since we just created a new Infra instance, let's wait for it to provision everything before proceeding
			log.Debug("Returning to reconcile phase to give some time for the Infinispan Operator to deploy")
			return true, nil
		}
		if ready {
			if err := kogitores.SetInfinispanEnvVars(r.client, infra, instance, requestedDeployment); err != nil {
				return true, err
			}
			log.Debug("KogitoInfra is ready, proceed!")
		} else if !ready {
			log.Debug("KogitoInfra is not ready, requeue")
			return true, nil
		}
	}
	return false, nil
}

func (r *ReconcileKogitoApp) ensureKafka(instance *v1alpha1.KogitoApp, requestedDeployment *oappsv1.DeploymentConfig) (requeue bool, err error) {
	log.Debug("Verify if we need to deploy Kafka")
	if instance.Spec.Infra.InstallKafka == v1alpha1.KogitoAppInfraInstallKafkaAlways {
		infra, created, ready, err := infrastructure.EnsureKogitoInfra(instance.Namespace, r.client).WithKafka()
		if err != nil {
			return true, err
		}
		if created {
			// since we just created a new Infra instance, let's wait for it to provision everything before proceeding
			log.Debug("Returning to reconcile phase to give some time for the Kafka Operator to deploy")
			return true, nil
		}
		if ready {
			if err := kogitores.SetKafkaEnvVars(r.client, infra, instance, requestedDeployment); err != nil {
				return true, err
			}
			log.Debug("KogitoInfra is ready, proceed!")
			return false, nil
		} else if !ready {
			log.Debug("KogitoInfra is not ready, requeue")
			return true, nil
		}
	}
	return false, nil
}

func (r *ReconcileKogitoApp) injectExternalVariables(instance *v1alpha1.KogitoApp, requestedDeployment *oappsv1.DeploymentConfig) error {
	if requestedDeployment == nil {
		return nil
	}

	log.Infof("Injecting external route URL to '%s'", instance.Name)
	if err := kogitores.SetExternalRouteEnvVar(r.client, instance, requestedDeployment); err != nil {
		return err
	}

	if len(requestedDeployment.Spec.Template.Spec.Containers) > 0 {
		log.Infof("Injecting external services references to '%s'", instance.Name)
		if err := infrastructure.InjectEnvVarsFromExternalServices(instance, &requestedDeployment.Spec.Template.Spec.Containers[0], r.client); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileKogitoApp) rollOutDeploymentIfConfigMapBroken(instance *v1alpha1.KogitoApp, deployed map[reflect.Type][]utilsres.KubernetesResource) (err error) {
	deployedDeployment := framework.GetResource(reflect.TypeOf(oappsv1.DeploymentConfig{}), instance.Name, deployed)
	deployedConfigMap := framework.GetResource(reflect.TypeOf(corev1.ConfigMap{}), kogitores.GenerateProtoBufConfigMapName(instance), deployed)

	if deployedDeployment == nil {
		return nil
	}

	// only rolls out the dc if all replicas are available
	if !framework.IsSafeToRollOutDeploymentConfig(deployedDeployment.(*oappsv1.DeploymentConfig)) {
		return nil
	}

	if deployedConfigMap == nil || !kogitores.CheckProtoBufConfigMapIntegrity(deployedConfigMap.(*corev1.ConfigMap)) {
		if _, err := openshift.DeploymentC(r.client).RolloutLatest(deployedDeployment.GetName(), deployedDeployment.GetNamespace()); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func getKubernetesResources(kogitoRes *kogitores.KogitoAppResources) []utilsres.KubernetesResource {
	k8sRes := make([]utilsres.KubernetesResource, 7)

	if kogitoRes.BuildConfigS2I != nil {
		k8sRes = append(k8sRes, kogitoRes.BuildConfigS2I)
	}
	if kogitoRes.BuildConfigRuntime != nil {
		k8sRes = append(k8sRes, kogitoRes.BuildConfigRuntime)
	}
	if kogitoRes.ImageStreamS2I != nil {
		k8sRes = append(k8sRes, kogitoRes.ImageStreamS2I)
	}
	if kogitoRes.ImageStreamRuntime != nil {
		k8sRes = append(k8sRes, kogitoRes.ImageStreamRuntime)
	}
	if kogitoRes.DeploymentConfig != nil {
		k8sRes = append(k8sRes, kogitoRes.DeploymentConfig)
	}
	if kogitoRes.Service != nil {
		k8sRes = append(k8sRes, kogitoRes.Service)
	}
	if kogitoRes.Route != nil {
		k8sRes = append(k8sRes, kogitoRes.Route)
	}
	if kogitoRes.ServiceMonitor != nil {
		k8sRes = append(k8sRes, kogitoRes.ServiceMonitor)
	}
	if kogitoRes.ProtoBufCM != nil {
		k8sRes = append(k8sRes, kogitoRes.ProtoBufCM)
	}

	return k8sRes
}
