package subapp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/kiegroup/submarine-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/constants"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/logs"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/shared"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/status"
	obuildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	client      client.Client
	scheme      *runtime.Scheme
	cache       cachev1.Cache
	imageClient *imagev1.ImageV1Client
	buildClient *buildv1.BuildV1Client
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

	// Set some CR defaults
	if len(instance.Spec.Name) == 0 {
		instance.Spec.Name = instance.Name
	}
	if instance.Spec.Runtime == "" || (instance.Spec.Runtime != v1alpha1.QuarkusRuntimeType && instance.Spec.Runtime != v1alpha1.SpringbootRuntimeType) {
		instance.Spec.Runtime = v1alpha1.QuarkusRuntimeType
	}

	// Define new BuildConfig objects
	buildConfigs := newBCsForCR(instance)
	for imageType, buildConfig := range buildConfigs {
		if _, err := r.ensureImageStream(
			buildConfig.Name,
			instance,
		); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this BC already exists
		_, err = r.buildClient.BuildConfigs(buildConfig.Namespace).Get(buildConfig.Name, metav1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating a new BuildConfig ", buildConfig.Name, " in namespace ", buildConfig.Namespace)
			bc, err := r.buildClient.BuildConfigs(buildConfig.Namespace).Create(&buildConfig)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Trigger first build of new builder BC
			if imageType == "builder" {
				if err = r.triggerBuild(*bc, instance); err != nil {
					return reconcile.Result{}, err
				}
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create new DeploymentConfig object
	/*
		depConfig := oappsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Spec.Name,
				Namespace: instance.Namespace,
				Labels: map[string]string{
					"app": instance.Name,
				},
			},
		}

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
func newBCsForCR(cr *v1alpha1.SubApp) map[string]obuildv1.BuildConfig {
	buildConfigs := map[string]obuildv1.BuildConfig{}
	serviceBC := obuildv1.BuildConfig{}
	images := constants.RuntimeImageDefaults[cr.Spec.Runtime]

	for _, imageDefaults := range images {
		if imageDefaults.BuilderImage {
			builderName := strings.Join([]string{cr.Spec.Name, "builder"}, "-")
			builderBC := obuildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      builderName,
					Namespace: cr.Namespace,
					Labels: map[string]string{
						"app": cr.Name,
					},
				},
			}
			builderBC.SetGroupVersionKind(obuildv1.SchemeGroupVersion.WithKind("BuildConfig"))
			builderBC.Spec.Source.Git = &obuildv1.GitBuildSource{
				URI: cr.Spec.Build.GitSource.URI,
				Ref: cr.Spec.Build.GitSource.Reference,
			}
			builderBC.Spec.Source.ContextDir = cr.Spec.Build.GitSource.ContextDir
			builderBC.Spec.Output.To = &corev1.ObjectReference{Name: strings.Join([]string{builderName, "latest"}, ":"), Kind: "ImageStreamTag"}
			builderBC.Spec.Strategy.Type = obuildv1.SourceBuildStrategyType
			builderBC.Spec.Strategy.SourceStrategy = &obuildv1.SourceBuildStrategy{
				Incremental: &cr.Spec.Build.Incremental,
				Env:         cr.Spec.Build.Env,
				From: corev1.ObjectReference{
					Name:      fmt.Sprintf("%s:%s", imageDefaults.ImageStreamName, imageDefaults.ImageStreamTag),
					Namespace: imageDefaults.ImageStreamNamespace,
					Kind:      "ImageStreamTag",
				},
			}

			buildConfigs["builder"] = builderBC
		} else {
			serviceBC = obuildv1.BuildConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Spec.Name,
					Namespace: cr.Namespace,
					Labels: map[string]string{
						"app": cr.Name,
					},
				},
			}
			serviceBC.SetGroupVersionKind(obuildv1.SchemeGroupVersion.WithKind("BuildConfig"))
			serviceBC.Spec.Source.Type = obuildv1.BuildSourceImage
			serviceBC.Spec.Output.To = &corev1.ObjectReference{Name: strings.Join([]string{cr.Spec.Name, "latest"}, ":"), Kind: "ImageStreamTag"}
			serviceBC.Spec.Strategy.Type = obuildv1.SourceBuildStrategyType
			serviceBC.Spec.Strategy.SourceStrategy = &obuildv1.SourceBuildStrategy{
				From: corev1.ObjectReference{
					Name:      fmt.Sprintf("%s:%s", imageDefaults.ImageStreamName, imageDefaults.ImageStreamTag),
					Namespace: imageDefaults.ImageStreamNamespace,
					Kind:      "ImageStreamTag",
				},
			}
		}
	}

	serviceBC.Spec.Source.Images = []obuildv1.ImageSource{
		{
			From: *buildConfigs["builder"].Spec.Output.To,
			Paths: []obuildv1.ImageSourcePath{
				{
					DestinationDir: ".",
					SourcePath:     "/home/submarine/bin",
				},
			},
		},
	}
	serviceBC.Spec.Triggers = []obuildv1.BuildTriggerPolicy{
		{
			Type:        obuildv1.ImageChangeBuildTriggerType,
			ImageChange: &obuildv1.ImageChangeTrigger{From: buildConfigs["builder"].Spec.Output.To},
		},
	}
	buildConfigs["service"] = serviceBC

	return buildConfigs
}

// updateBuildConfigs ...
func (r *ReconcileSubApp) updateBuildConfigs(instance *v1alpha1.SubApp, bc *obuildv1.BuildConfig) (bool, error) {
	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)
	listOps := &client.ListOptions{Namespace: instance.Namespace}
	bcList := &obuildv1.BuildConfigList{}
	err := r.client.List(context.TODO(), listOps, bcList)
	if err != nil {
		log.Warn("Failed to list bc's. ", err)
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

func (r *ReconcileSubApp) bcUpdateCheck(current, new obuildv1.BuildConfig, bcUpdates []obuildv1.BuildConfig, cr *v1alpha1.SubApp) []obuildv1.BuildConfig {
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

// checkImageStreamTag checks for ImageStream
func (r *ReconcileSubApp) checkImageStreamTag(name, namespace string) bool {
	log := log.With("kind", "ImageStreamTag", "name", name, "namespace", namespace)
	result := strings.Split(name, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}
	tagName := fmt.Sprintf("%s:%s", result[0], result[1])
	_, err := r.imageClient.ImageStreamTags(namespace).Get(tagName, metav1.GetOptions{})
	if err != nil {
		log.Debug("Object does not exist")
		return false
	}
	return true
}

// ensureImageStream ...
func (r *ReconcileSubApp) ensureImageStream(name string, cr *v1alpha1.SubApp) (string, error) {
	if r.checkImageStreamTag(name, cr.Namespace) {
		return cr.Namespace, nil
	}
	err := r.createLocalImageTag(name, cr)
	if err != nil {
		log.Error(err)
		return cr.Namespace, err
	}
	return cr.Namespace, nil
}

// createLocalImageTag creates local ImageStreamTag
func (r *ReconcileSubApp) createLocalImageTag(tagRefName string, cr *v1alpha1.SubApp) error {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, "latest")
	}

	isnew := &oimagev1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", result[0], result[1]),
			Namespace: cr.Namespace,
		},
		Tag: &oimagev1.TagReference{
			Name: result[1],
			ReferencePolicy: oimagev1.TagReferencePolicy{
				Type: oimagev1.LocalTagReferencePolicy,
			},
			// From: cr.Spec.Build.From,
		},
	}
	isnew.SetGroupVersionKind(oimagev1.SchemeGroupVersion.WithKind("ImageStreamTag"))

	log := log.With("kind", isnew.GetObjectKind().GroupVersionKind().Kind, "name", isnew.Name, "namespace", isnew.Namespace)
	log.Info("Creating")

	_, err := r.imageClient.ImageStreamTags(isnew.Namespace).Create(isnew)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error("Issue creating object. ", err)
		return err
	}
	return nil
}

// triggerBuild triggers a BuildConfig to start a new build
func (r *ReconcileSubApp) triggerBuild(bc obuildv1.BuildConfig, cr *v1alpha1.SubApp) error {
	buildConfig, err := r.buildClient.BuildConfigs(bc.Namespace).Get(bc.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	buildRequest := obuildv1.BuildRequest{ObjectMeta: metav1.ObjectMeta{Name: buildConfig.Name}}
	buildRequest.SetGroupVersionKind(obuildv1.SchemeGroupVersion.WithKind("BuildRequest"))
	buildRequest.TriggeredBy = []obuildv1.BuildTriggerCause{{Message: fmt.Sprintf("Triggered by %s operator", cr.Kind)}}
	build, err := r.buildClient.BuildConfigs(buildConfig.Namespace).Instantiate(buildConfig.Name, &buildRequest)
	if err != nil {
		return err
	}

	log.Info("Name of the triggered build is ", build.Name)
	return nil
}
