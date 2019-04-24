package subapp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kiegroup/submarine-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/logs"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/shared"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/status"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logs.GetLogger("controller_subapp")
var _ reconcile.Reconciler = &ReconcileSubApp{}

// ReconcileSubApp reconciles a SubApp object
type ReconcileSubApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	cache  cachev1.Cache
}

// Reconcile reads that state of the cluster for a SubApp object and makes changes based on the state read
// and what is in the SubApp.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSubApp) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling SubApp")

	// Fetch the SubApp instance
	instance := &v1alpha1.SubApp{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new BuildConfig object
	buildConfig := newBCForCR(instance)
	// Set SubApp instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, buildConfig, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this BC already exists
	found := &buildv1.BuildConfig{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: buildConfig.Name, Namespace: buildConfig.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new BuildConfig", "BuildConfig.Namespace", buildConfig.Namespace, "BuildConfig.Name", buildConfig.Name)
		err = r.client.Create(context.TODO(), buildConfig)
		if err != nil {
			return reconcile.Result{}, err
		}

		// BC created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
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

	// Fetch the cached SubApp instance
	cachedInstance := &v1alpha1.SubApp{}
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

// newBCForCR returns a BuildConfig with the same name/namespace as the cr
func newBCForCR(cr *v1alpha1.SubApp) *buildv1.BuildConfig {
	name := cr.Name
	if cr.Spec.Name != "" {
		name = cr.Spec.Name
	}
	if len(cr.Spec.Build.From.Namespace) == 0 {
		cr.Spec.Build.From.Namespace = "openshift"
	}

	builderName := strings.Join([]string{name, "builder"}, "-")
	bc := buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builderName,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
	}

	bc.Spec.Source.Git = &buildv1.GitBuildSource{
		URI: cr.Spec.Build.GitSource.URI,
		Ref: cr.Spec.Build.GitSource.Reference,
	}
	bc.Spec.Source.ContextDir = cr.Spec.Build.GitSource.ContextDir
	bc.Spec.Output.To = &corev1.ObjectReference{Name: strings.Join([]string{builderName, "latest"}, ":"), Kind: "ImageStreamTag"}
	bc.Spec.Strategy.Type = buildv1.SourceBuildStrategyType
	bc.Spec.Strategy.SourceStrategy = &buildv1.SourceBuildStrategy{
		Incremental: &cr.Spec.Build.Incremental,
		From:        corev1.ObjectReference{Name: cr.Spec.Build.From.Name, Namespace: cr.Spec.Build.From.Namespace, Kind: "ImageStreamTag"},
	}

	return &bc
}

func (r *ReconcileSubApp) updateBuildConfigs(instance *v1alpha1.SubApp, bc *buildv1.BuildConfig) (bool, error) {
	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)
	listOps := &client.ListOptions{Namespace: instance.Namespace}
	bcList := &buildv1.BuildConfigList{}
	err := r.client.List(context.TODO(), listOps, bcList)
	if err != nil {
		log.Warn("Failed to list bc's. ", err)
		r.setFailedStatus(instance, v1alpha1.UnknownReason, err)
		return false, err
	}

	var bcUpdates []buildv1.BuildConfig
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
func (r *ReconcileSubApp) UpdateObj(obj v1alpha1.OpenShiftObject) (reconcile.Result, error) {
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

func (r *ReconcileSubApp) setFailedStatus(instance *v1alpha1.SubApp, reason v1alpha1.ReasonType, err error) {
	status.SetFailed(instance, reason, err)
	_, updateError := r.UpdateObj(instance)
	if updateError != nil {
		log.Warn("Unable to update object after receiving failed status. ", err)
	}
}

func (r *ReconcileSubApp) bcUpdateCheck(current, new buildv1.BuildConfig, bcUpdates []buildv1.BuildConfig, cr *v1alpha1.SubApp) []buildv1.BuildConfig {
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
		bcnew.SetGroupVersionKind(buildv1.SchemeGroupVersion.WithKind("BuildConfig"))

		bcUpdates = append(bcUpdates, bcnew)
	}
	return bcUpdates
}

func (r *ReconcileSubApp) hasSpecChanges(instance, cached *v1alpha1.SubApp) bool {
	if !reflect.DeepEqual(instance.Spec, cached.Spec) {
		return true
	}
	return false
}

func (r *ReconcileSubApp) hasStatusChanges(instance, cached *v1alpha1.SubApp) bool {
	if !reflect.DeepEqual(instance.Status, cached.Status) {
		return true
	}
	return false
}
