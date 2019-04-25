package subapp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kiegroup/submarine-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/constants"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/logs"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/shared"
	"github.com/kiegroup/submarine-cloud-operator/pkg/controller/subapp/status"
	oappsv1 "github.com/openshift/api/apps/v1"
	obuildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	oimagev1 "github.com/openshift/api/image/v1"
	oroutev1 "github.com/openshift/api/route/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	depConfig, err := r.newDCForCR(instance, buildConfigs["service"])
	if err != nil {
		return reconcile.Result{}, err
	}
	rResult, err := r.createObj(
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
	serviceRoute := ""
	if len(depConfig.Spec.Template.Spec.Containers[0].Ports) != 0 {
		servicePorts := []corev1.ServicePort{}
		for _, port := range depConfig.Spec.Template.Spec.Containers[0].Ports {
			servicePorts = append(servicePorts, corev1.ServicePort{
				Name:       port.Name,
				Protocol:   port.Protocol,
				Port:       port.ContainerPort,
				TargetPort: intstr.FromInt(int(port.ContainerPort)),
			},
			)
		}
		service := corev1.Service{
			ObjectMeta: depConfig.ObjectMeta,
			Spec: corev1.ServiceSpec{
				Selector: depConfig.Spec.Selector,
				Type:     corev1.ServiceTypeClusterIP,
				Ports:    servicePorts,
			},
		}
		service.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))
		err = controllerutil.SetControllerReference(instance, &service, r.scheme)
		if err != nil {
			log.Error(err)
		}

		service.ResourceVersion = ""
		rResult, err := r.createObj(
			&service,
			r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &corev1.Service{}),
		)
		if err != nil {
			return rResult, err
		}

		// Create route
		rt := oroutev1.Route{
			ObjectMeta: service.ObjectMeta,
			Spec: oroutev1.RouteSpec{
				Port: &oroutev1.RoutePort{
					TargetPort: intstr.FromString("http"),
				},
				To: oroutev1.RouteTargetReference{
					Kind: "Service",
					Name: service.Name,
				},
			},
		}
		if serviceRoute = r.GetRouteHost(rt, instance); serviceRoute != "" {
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

// newDCForCR returns a BuildConfig with the same name/namespace as the cr
func (r *ReconcileSubApp) newDCForCR(cr *v1alpha1.SubApp, serviceBC obuildv1.BuildConfig) (oappsv1.DeploymentConfig, error) {
	replicas := int32(1)
	if cr.Spec.Replicas != nil {
		replicas = *cr.Spec.Replicas
	}
	labels := map[string]string{
		"app": cr.Name,
	}
	ports := []corev1.ContainerPort{}
	bcNamespace := cr.Namespace
	if serviceBC.Spec.Output.To.Namespace != "" {
		bcNamespace = serviceBC.Spec.Output.To.Namespace
	}

	isTag, err := r.imageClient.ImageStreamTags(bcNamespace).Get(serviceBC.Spec.Output.To.Name, metav1.GetOptions{})
	if err != nil {
		log.Warn(cr.Spec.Name, " DeploymentConfig will start when ImageStream build completes.")
	}
	if len(isTag.Image.DockerImageMetadata.Raw) != 0 {
		obj := &dockerv10.DockerImage{}
		err = json.Unmarshal(isTag.Image.DockerImageMetadata.Raw, obj)
		if err != nil {
			log.Error(err)
		}
		for i, value := range obj.Config.Labels {
			if strings.Contains(i, "org.kie/") {
				result := strings.Split(i, "/")
				labels[result[len(result)-1]] = value
			}
			if i == "io.openshift.expose-services" {
				results := strings.Split(value, ",")
				for _, item := range results {
					portResults := strings.Split(item, ":")
					port, err := strconv.Atoi(portResults[0])
					if err != nil {
						log.Error(err)
					}
					portName := portResults[1]
					ports = append(ports, corev1.ContainerPort{Name: portName, ContainerPort: int32(port), Protocol: corev1.ProtocolTCP})
				}
			}
		}
	}

	depConfig := oappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: oappsv1.DeploymentConfigSpec{
			Replicas: replicas,
			Selector: labels,
			Strategy: oappsv1.DeploymentStrategy{
				Type: oappsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            cr.Spec.Name,
							Env:             cr.Spec.Env,
							Resources:       cr.Spec.Resources,
							Image:           serviceBC.Spec.Output.To.Name,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			Triggers: oappsv1.DeploymentTriggerPolicies{
				{Type: oappsv1.DeploymentTriggerOnConfigChange},
				{
					Type: oappsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &oappsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{cr.Spec.Name},
						From:           *serviceBC.Spec.Output.To,
					},
				},
			},
		},
	}
	if len(ports) != 0 {
		depConfig.Spec.Template.Spec.Containers[0].Ports = ports
	}
	depConfig.SetGroupVersionKind(oappsv1.SchemeGroupVersion.WithKind("DeploymentConfig"))
	err = controllerutil.SetControllerReference(cr, &depConfig, r.scheme)
	if err != nil {
		log.Error(err)
		return oappsv1.DeploymentConfig{}, err
	}

	return depConfig, nil
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

// createObj creates an object based on the error passed in from a `client.Get`
func (r *ReconcileSubApp) createObj(obj v1alpha1.OpenShiftObject, err error) (reconcile.Result, error) {
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

func (r *ReconcileSubApp) updateDeploymentConfigs(instance *v1alpha1.SubApp, depConfig oappsv1.DeploymentConfig) (bool, error) {
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

func (r *ReconcileSubApp) dcUpdateCheck(current, new oappsv1.DeploymentConfig, dcUpdates []oappsv1.DeploymentConfig, cr *v1alpha1.SubApp) []oappsv1.DeploymentConfig {
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

func getDeploymentsStatuses(dcs []oappsv1.DeploymentConfig, cr *v1alpha1.SubApp) v1alpha1.Deployments {
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
func (r *ReconcileSubApp) GetRouteHost(route oroutev1.Route, cr *v1alpha1.SubApp) string {
	route.SetGroupVersionKind(oroutev1.SchemeGroupVersion.WithKind("Route"))
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
