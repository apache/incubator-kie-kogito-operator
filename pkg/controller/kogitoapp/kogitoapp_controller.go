package kogitoapp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	inventory "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/inventory"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/logs"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/status"
	oappsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	oroutev1 "github.com/openshift/api/route/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logs.GetLogger("controller_kogitoapp")
var _ reconcile.Reconciler = &ReconcileKogitoApp{}

// ReconcileKogitoApp reconciles a KogitoApp object
type ReconcileKogitoApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	cache       cachev1.Cache
	imageClient *imagev1.ImageV1Interface
	buildClient *buildv1.BuildV1Interface
}

// Reconcile reads that state of the cluster for a KogitoApp object and makes changes based on the state read
// and what is in the KogitoApp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling KogitoApp")

	// Fetch the KogitoApp instance
	instance := &v1alpha1.KogitoApp{}
	if exists, err := inventory.FetchResourceWithKey(r.client, request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, err
	} else if !exists {
		return reconcile.Result{}, nil
	}

	// Set some CR defaults
	if len(instance.Spec.Name) == 0 {
		instance.Spec.Name = instance.Name
	}
	if instance.Spec.Runtime == "" || (instance.Spec.Runtime != v1alpha1.QuarkusRuntimeType && instance.Spec.Runtime != v1alpha1.SpringbootRuntimeType) {
		instance.Spec.Runtime = v1alpha1.QuarkusRuntimeType
	}

	// create resources in the cluster that not exists
	kogitoInv, err := inventory.CreateResources(&inventory.BuilderContext{
		Client:      r.client,
		ImageClient: *r.imageClient,
		KogitoApp:   instance,
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// ensure builds
	if imageExists, err := r.ensureApplicationImageExists(kogitoInv, instance); err != nil {
		return reconcile.Result{}, err
	} else if !imageExists {
		// let's wait for the build to finish
		return reconcile.Result{RequeueAfter: time.Duration(30) * time.Second}, nil
	}

	// checks for dc updates
	if kogitoInv.DeploymentConfig != nil {
		if err := controllerutil.SetControllerReference(instance, kogitoInv.DeploymentConfig, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		if dcUpdated, err := r.updateDeploymentConfigs(instance, *kogitoInv.DeploymentConfig); err != nil {
			return reconcile.Result{}, err
		} else if dcUpdated && status.SetProvisioning(instance) {
			return r.UpdateObj(instance)
		}
	}

	// Expose DC with service and route
	if kogitoInv.Service != nil {
		err = controllerutil.SetControllerReference(instance, kogitoInv.Service, r.scheme)
		if err != nil {
			log.Error(err)
		}
		// TODO: this is really needed?
		if serviceRoute := r.GetRouteHost(*kogitoInv.Route, instance); serviceRoute != "" {
			instance.Status.Route = fmt.Sprintf("http://%s", serviceRoute)
		}
	}

	/*

		bcUpdated, err := r.updateBuildConfigs(instance, buildConfig)
		if err != nil {
			return reconcile.Result{}, err
		}
		if bcUpdated && status.SetProvisioning(instance) {
			return r.UpdateObj(instance)
		}
	*/

	// Fetch the cached KogitoApp instance
	cachedInstance := &v1alpha1.KogitoApp{}
	err = r.cache.Get(context.TODO(), request.NamespacedName, cachedInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.setFailedStatus(instance, v1alpha1.UnknownReason, err)
		return reconcile.Result{}, err
	}

	// Update CR if needed
	if r.hasSpecChanges(instance, cachedInstance) {
		if status.SetProvisioning(instance) && instance.ResourceVersion == cachedInstance.ResourceVersion {
			return r.UpdateObj(instance)
		}
		return reconcile.Result{Requeue: true}, nil
	}
	if r.hasStatusChanges(instance, cachedInstance) {
		if instance.ResourceVersion == cachedInstance.ResourceVersion {
			return r.UpdateObj(instance)
		}
		return reconcile.Result{Requeue: true}, nil
	}
	if status.SetDeployed(instance) {
		if instance.ResourceVersion == cachedInstance.ResourceVersion {
			return r.UpdateObj(instance)
		}
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileKogitoApp) ensureApplicationImageExists(inv *inventory.KogitoAppInventory, instance *v1alpha1.KogitoApp) (bool, error) {
	buildServiceState, err := inventory.EnsureImageBuild(*r.buildClient, *r.imageClient, inv.BuildConfigService)
	if err != nil {
		return false, err
	}

	// we have the final image built, there's no need to proceed
	if buildServiceState.ImageExists {
		log.Debugf("Final application image exists there's no need to trigger any build")
		return true, nil
	}

	log.Infof("No image found for the application %s. Trying to trigger a new build.", instance.Spec.Name)

	// verify s2i build and image
	if state, err := inventory.EnsureImageBuild(*r.buildClient, *r.imageClient, inv.BuildConfigS2I); err != nil {
		return false, err
	} else if state.BuildRunning {
		// build is running, nothing to do
		return false, nil
	} else if !state.ImageExists && !state.BuildRunning {
		log.Infof("There's no image nor build for %s running, triggering build", inv.BuildConfigS2I.Name)
		_, err := inventory.TriggerBuild(*r.buildClient, inv.BuildConfigS2I, instance)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	// verify service build and image
	if !buildServiceState.ImageExists {
		log.Warnf("There's no image for %s.", inv.BuildConfigService.Name)
		return false, nil
	}

	log.Debugf("There are images for both builds, nothing to do")
	return true, nil
}

// updateBuildConfigs ...
func (r *ReconcileKogitoApp) updateBuildConfigs(instance *v1alpha1.KogitoApp, bc *obuildv1.BuildConfig) (bool, error) {
	bcList := &obuildv1.BuildConfigList{}
	err := inventory.ListResourceWithNamespace(r.client, instance.Namespace, bcList)
	if err != nil {
		r.setFailedStatus(instance, v1alpha1.UnknownReason, err)
		return false, err
	}

	var bcUpdates []obuildv1.BuildConfig
	for _, lbc := range bcList.Items {
		if bc.Name == lbc.Name {
			bcUpdates = r.bcUpdateCheck(*bc, lbc, bcUpdates, instance)
		}
	}
	if len(bcUpdates) > 0 {
		for _, uBc := range bcUpdates {
			fmt.Println(uBc)
			_, err := r.UpdateObj(&uBc)
			if err != nil {
				r.setFailedStatus(instance, v1alpha1.DeploymentFailedReason, err)
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

// UpdateObj reconciles the given object
func (r *ReconcileKogitoApp) UpdateObj(obj v1alpha1.OpenShiftObject) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	log.Info("Updating")
	err := r.client.Update(context.TODO(), obj)
	if err != nil {
		log.Warn("Failed to update object. ", err)
		return reconcile.Result{}, err
	}
	// Object updated - return and requeue
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileKogitoApp) setFailedStatus(instance *v1alpha1.KogitoApp, reason v1alpha1.ReasonType, err error) {
	status.SetFailed(instance, reason, err)
	_, updateError := r.UpdateObj(instance)
	if updateError != nil {
		log.Warn("Unable to update object after receiving failed status. ", err)
	}
}

func (r *ReconcileKogitoApp) bcUpdateCheck(current, new obuildv1.BuildConfig, bcUpdates []obuildv1.BuildConfig, cr *v1alpha1.KogitoApp) []obuildv1.BuildConfig {
	log := log.With("kind", current.GetObjectKind().GroupVersionKind().Kind, "name", current.Name, "namespace", current.Namespace)
	update := false

	if !reflect.DeepEqual(current.Spec.Source, new.Spec.Source) {
		log.Debug("Changes detected in 'Source' config.", " OLD - ", current.Spec.Source, " NEW - ", new.Spec.Source)
		update = true
	}
	if !shared.EnvVarCheck(current.Spec.Strategy.SourceStrategy.Env, new.Spec.Strategy.SourceStrategy.Env) {
		log.Debug("Changes detected in 'Env' config.", " OLD - ", current.Spec.Strategy.SourceStrategy.Env, " NEW - ", new.Spec.Strategy.SourceStrategy.Env)
		update = true
	}
	if !reflect.DeepEqual(current.Spec.Resources, new.Spec.Resources) {
		log.Debug("Changes detected in 'Resource' config.", " OLD - ", current.Spec.Resources, " NEW - ", new.Spec.Resources)
		update = true
	}

	if update {
		bcnew := new
		err := controllerutil.SetControllerReference(cr, &bcnew, r.scheme)
		if err != nil {
			log.Error("Error setting controller reference for bc. ", err)
		}
		bcnew.SetNamespace(current.Namespace)
		bcnew.SetResourceVersion(current.ResourceVersion)
		bcnew.SetGroupVersionKind(obuildv1.SchemeGroupVersion.WithKind("BuildConfig"))

		bcUpdates = append(bcUpdates, bcnew)
	}
	return bcUpdates
}

func (r *ReconcileKogitoApp) hasSpecChanges(instance, cached *v1alpha1.KogitoApp) bool {
	if !reflect.DeepEqual(instance.Spec, cached.Spec) {
		return true
	}
	return false
}

func (r *ReconcileKogitoApp) hasStatusChanges(instance, cached *v1alpha1.KogitoApp) bool {
	if !reflect.DeepEqual(instance.Status, cached.Status) {
		return true
	}
	return false
}

func (r *ReconcileKogitoApp) updateDeploymentConfigs(instance *v1alpha1.KogitoApp, depConfig oappsv1.DeploymentConfig) (bool, error) {
	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)
	dcList := &oappsv1.DeploymentConfigList{}
	err := inventory.ListResourceWithNamespace(r.client, instance.Namespace, dcList)
	if err != nil {
		r.setFailedStatus(instance, v1alpha1.UnknownReason, err)
		return false, err
	}
	instance.Status.Deployments = getDeploymentsStatuses(dcList.Items, instance)

	var dcUpdates []oappsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		if dc.Name == depConfig.Name {
			dcUpdates = r.dcUpdateCheck(dc, depConfig, dcUpdates, instance)
		}
	}
	log.Debugf("There are %d updated DCs", len(dcUpdates))
	if len(dcUpdates) > 0 {
		for _, uDc := range dcUpdates {
			_, err := r.UpdateObj(&uDc)
			if err != nil {
				r.setFailedStatus(instance, v1alpha1.DeploymentFailedReason, err)
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (r *ReconcileKogitoApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1alpha1.KogitoApp) []oappsv1.DeploymentConfig {
	log := log.With("kind", new.GetObjectKind().GroupVersionKind().Kind, "name", current.Name, "namespace", current.Namespace)
	update := false
	if !reflect.DeepEqual(current.Spec.Template.Labels, new.Spec.Template.Labels) {
		log.Debug("Changes detected in labels.", " OLD - ", current.Spec.Template.Labels, " NEW - ", new.Spec.Template.Labels)
		update = true
	}
	if current.Spec.Replicas != new.Spec.Replicas {
		log.Debug("Changes detected in replicas.", " OLD - ", current.Spec.Replicas, " NEW - ", new.Spec.Replicas)
		update = true
	}

	cContainer := current.Spec.Template.Spec.Containers[0]
	nContainer := new.Spec.Template.Spec.Containers[0]
	if !shared.EnvVarCheck(cContainer.Env, nContainer.Env) {
		log.Debug("Changes detected in 'Env' config.", " OLD - ", cContainer.Env, " NEW - ", nContainer.Env)
		update = true
	}
	if !reflect.DeepEqual(cContainer.Resources, nContainer.Resources) {
		log.Debug("Changes detected in 'Resource' config.", " OLD - ", cContainer.Resources, " NEW - ", nContainer.Resources)
		update = true
	}
	if !reflect.DeepEqual(cContainer.Ports, nContainer.Ports) {
		log.Debug("Changes detected in 'Ports' config.", " OLD - ", cContainer.Ports, " NEW - ", nContainer.Ports)
		update = true
	}
	if update {
		dcnew := new
		err := controllerutil.SetControllerReference(cr, &dcnew, r.scheme)
		if err != nil {
			log.Error("Error setting controller reference for dc. ", err)
		}
		dcnew.SetNamespace(current.Namespace)
		dcnew.SetResourceVersion(current.ResourceVersion)
		dcnew.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))

		dcUpdates = append(dcUpdates, dcnew)
	}
	return dcUpdates
}

func getDeploymentsStatuses(dcs []oappsv1.DeploymentConfig, cr *v1alpha1.KogitoApp) v1alpha1.Deployments {
	var ready, starting, stopped []string
	for _, dc := range dcs {
		for _, ownerRef := range dc.GetOwnerReferences() {
			if ownerRef.UID == cr.UID {
				if dc.Spec.Replicas == 0 {
					stopped = append(stopped, dc.Name)
				} else if dc.Status.Replicas == 0 {
					stopped = append(stopped, dc.Name)
				} else if dc.Status.ReadyReplicas < dc.Status.Replicas {
					starting = append(starting, dc.Name)
				} else {
					ready = append(ready, dc.Name)
				}
			}
		}
	}
	log.Debugf("Found DCs with status stopped [%s], starting [%s], and ready [%s]", stopped, starting, ready)
	return v1alpha1.Deployments{
		Stopped:  stopped,
		Starting: starting,
		Ready:    ready,
	}
}

// GetRouteHost returns the Hostname of the route provided
func (r *ReconcileKogitoApp) GetRouteHost(route oroutev1.Route, cr *v1alpha1.KogitoApp) string {
	log := log.With("kind", route.GetObjectKind().GroupVersionKind().Kind, "name", route.Name, "namespace", route.Namespace)
	err := controllerutil.SetControllerReference(cr, &route, r.scheme)
	if err != nil {
		log.Error("Error setting controller reference. ", err)
	}

	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		if exists, err := inventory.FetchResourceWithKey(r.client, types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, &route); exists {
			break
		} else if err != nil {
			log.Error("Error getting Route. ", err)
		}
	}

	return route.Spec.Host
}
