package kogitoapp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	defs "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/definitions"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/logs"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/status"
	oappsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	oimagev1 "github.com/openshift/api/image/v1"
	oroutev1 "github.com/openshift/api/route/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	imageClient *imagev1.ImageV1Client
	buildClient *buildv1.BuildV1Client
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

	// TODO: move fetch objects to somewhere else
	// TODO: move object verification to somewhere else
	// Check if the SA already exists
	sa := defs.NewServiceAccount(instance)
	log.Info("Creating the ServiceAccount ", sa.Name, " in namespace ", sa.Namespace)
	rResult, err := r.createObj(&sa,
		r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, &corev1.ServiceAccount{}))
	if err != nil {
		return rResult, err
	}

	roleBinding := defs.NewRoleBinding(instance, &sa)
	log.Info("Creating the RoleBinding ", roleBinding.Name, " in namespace ", roleBinding.Namespace)
	rResult, err = r.createObj(&roleBinding,
		r.client.Get(context.TODO(), types.NamespacedName{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, &rbacv1.RoleBinding{}))
	if err != nil {
		return rResult, err
	}

	// Define new BuildConfig objects
	buildConfigs, err := defs.NewBuildConfig(instance)
	if err != nil {
		return rResult, err
	}

	for buildType, buildConfig := range buildConfigs.AsMap {
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
			bc, err := r.buildClient.BuildConfigs(buildConfig.Namespace).Create(buildConfig)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Trigger first build of new builder BC
			if buildType == defs.S2IBuildType {
				if err = r.triggerBuild(*bc, instance); err != nil {
					return reconcile.Result{}, err
				}
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create new DeploymentConfig object
	depConfig, err := r.newDCForCR(instance, buildConfigs.AsMap[defs.RunnerBuildType], &sa)
	if err != nil {
		return reconcile.Result{}, err
	}
	rResult, err = r.createObj(
		&depConfig,
		r.client.Get(context.TODO(), types.NamespacedName{Name: depConfig.Name, Namespace: depConfig.Namespace}, &oappsv1.DeploymentConfig{}),
	)
	if err != nil {
		return rResult, err
	}

	dcUpdated, err := r.updateDeploymentConfigs(instance, depConfig)
	if err != nil {
		return reconcile.Result{}, err
	}
	if dcUpdated && status.SetProvisioning(instance) {
		return r.UpdateObj(instance)
	}

	// Expose DC with service and route
	service := defs.NewService(instance, &depConfig)
	if service != nil {
		err = controllerutil.SetControllerReference(instance, service, r.scheme)
		if err != nil {
			log.Error(err)
		}

		service.ResourceVersion = ""
		rResult, err = r.createObj(
			service,
			r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &corev1.Service{}),
		)
		if err != nil {
			return rResult, err
		}

		// Create route
		rt, err := defs.NewRoute(instance, service)
		if err != nil {
			return rResult, err
		}
		if serviceRoute := r.GetRouteHost(*rt, instance); serviceRoute != "" {
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

// newDCForCR returns a DeploymentConfig with the same name/namespace as the cr
func (r *ReconcileKogitoApp) newDCForCR(cr *v1alpha1.KogitoApp, serviceBC *obuildv1.BuildConfig, sa *corev1.ServiceAccount) (oappsv1.DeploymentConfig, error) {
	dockerImage := &dockerv10.DockerImage{}
	bcNamespace := cr.Namespace
	if serviceBC.Spec.Output.To.Namespace != "" {
		bcNamespace = serviceBC.Spec.Output.To.Namespace
	}

	isTag, err := r.imageClient.ImageStreamTags(bcNamespace).Get(serviceBC.Spec.Output.To.Name, metav1.GetOptions{})
	if err != nil {
		log.Warn(cr.Spec.Name, " DeploymentConfig will start when ImageStream build completes.")
	}
	if len(isTag.Image.DockerImageMetadata.Raw) != 0 {
		err = json.Unmarshal(isTag.Image.DockerImageMetadata.Raw, dockerImage)
		if err != nil {
			log.Error(err)
		}
	}

	depConfig, err := defs.NewDeploymentConfig(cr, serviceBC, sa, dockerImage)
	err = controllerutil.SetControllerReference(cr, depConfig, r.scheme)
	if err != nil {
		log.Error(err)
		return oappsv1.DeploymentConfig{}, err
	}

	return *depConfig, nil
}

// updateBuildConfigs ...
func (r *ReconcileKogitoApp) updateBuildConfigs(instance *v1alpha1.KogitoApp, bc *obuildv1.BuildConfig) (bool, error) {
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

// checkImageStreamTag checks for ImageStream
func (r *ReconcileKogitoApp) checkImageStreamTag(name, namespace string) bool {
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
func (r *ReconcileKogitoApp) ensureImageStream(name string, cr *v1alpha1.KogitoApp) (string, error) {
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
func (r *ReconcileKogitoApp) createLocalImageTag(tagRefName string, cr *v1alpha1.KogitoApp) error {
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
func (r *ReconcileKogitoApp) triggerBuild(bc obuildv1.BuildConfig, cr *v1alpha1.KogitoApp) error {
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

// createObj creates an object based on the error passed in from a `client.Get`
func (r *ReconcileKogitoApp) createObj(obj v1alpha1.OpenShiftObject, err error) (reconcile.Result, error) {
	log := log.With("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())

	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		log.Info("Creating")
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			log.Warn("Failed to create object. ", err)
			return reconcile.Result{}, err
		}
		// Object created successfully - return and requeue
		return reconcile.Result{RequeueAfter: time.Duration(200) * time.Millisecond}, nil
	} else if err != nil {
		log.Error("Failed to get object. ", err)
		return reconcile.Result{}, err
	}
	log.Debug("Skip reconcile - object already exists")
	return reconcile.Result{}, nil
}

func (r *ReconcileKogitoApp) updateDeploymentConfigs(instance *v1alpha1.KogitoApp, depConfig oappsv1.DeploymentConfig) (bool, error) {
	log := log.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)
	listOps := &client.ListOptions{Namespace: instance.Namespace}
	dcList := &oappsv1.DeploymentConfigList{}
	err := r.client.List(context.TODO(), listOps, dcList)
	if err != nil {
		log.Warn("Failed to list dc's. ", err)
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
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, &oroutev1.Route{})
	if err != nil && errors.IsNotFound(err) {
		route.ResourceVersion = ""
		_, err = r.createObj(
			&route,
			err,
		)
		if err != nil {
			log.Error("Error creating Route. ", err)
		}
	}

	found := &oroutev1.Route{}
	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, found)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Error("Error getting Route. ", err)
	}

	return found.Spec.Host
}
