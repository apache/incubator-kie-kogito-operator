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
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/controllers/build"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1beta1 "github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
)

const (
	imageStreamCreationReconcileTimeout = 10 * time.Second
)

// KogitoBuildReconciler reconciles a KogitoBuild object
type KogitoBuildReconciler struct {
	*client.Client
	Log    logger.Logger
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoBuild object and makes changes based on the state read
// and what is in the KogitoBuild.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;list;delete
// +kubebuilder:rbac:groups=infinispan.org,resources=infinispans,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=apps,resourceNames=kogito-operator,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=sources.knative.dev,resources=sinkbindings,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
func (r *KogitoBuildReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, resultErr error) {
	r.Log.Info("Reconciling for", "KogitoBuild", req.Name, "Namespace", req.Namespace)

	// fetch the requested instance
	instance := &appv1beta1.KogitoBuild{}
	exists, err := kubernetes.ResourceC(r.Client).FetchWithKey(req.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{Requeue: false}, err
	}
	if !exists {
		return reconcile.Result{Requeue: false}, nil
	}

	result = reconcile.Result{Requeue: false}

	defer r.handleStatusChange(instance, resultErr)

	if len(instance.Spec.Runtime) == 0 {
		instance.Spec.Runtime = appv1beta1.QuarkusRuntimeType
	}
	envs := instance.Spec.Env
	instance.Spec.Env = framework.EnvOverride(envs, corev1.EnvVar{Name: infrastructure.RuntimeTypeKey, Value: string(instance.Spec.Runtime)})
	if len(instance.Spec.TargetKogitoRuntime) == 0 {
		instance.Spec.TargetKogitoRuntime = instance.Name
	}

	// create the Kogito Image Streams to build the service if needed
	created, resultErr := build.CreateRequiredKogitoImageStreams(instance, r.Client)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	buildMgr, resultErr := build.New(instance, r.Client, r.Scheme)
	if resultErr != nil {
		return
	}
	// get the resources as we want them to be
	requested, resultErr := buildMgr.GetRequestedResources()
	if resultErr != nil {
		return
	}
	// get the deployed resources
	deployed, resultErr := buildMgr.GetDeployedResources()
	if resultErr != nil {
		return
	}
	//let's compare
	comparator := buildMgr.GetComparator()
	deltas := comparator.Compare(deployed, requested)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		r.Log.Info("Updating kogito build", "Create", len(delta.Added), "Update", len(delta.Updated), "Delete", len(delta.Removed), "Instance", resourceType)
		_, resultErr = kubernetes.ResourceC(r.Client).CreateResources(delta.Added)
		if resultErr != nil {
			return
		}
		_, resultErr = kubernetes.ResourceC(r.Client).UpdateResources(deployed[resourceType], delta.Updated)
		if resultErr != nil {
			return
		}
		_, resultErr = kubernetes.ResourceC(r.Client).DeleteResources(delta.Removed)
		if resultErr != nil {
			return
		}

		if len(delta.Updated) > 0 {
			if resultErr = r.onResourceChange(instance, resourceType, delta.Updated); resultErr != nil {
				return
			}
		}
	}

	return
}

// SetupWithManager registers the controller with manager
func (r *KogitoBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoBuild controller")
	b := ctrl.NewControllerManagedBy(mgr).For(&appv1beta1.KogitoBuild{})
	if r.IsOpenshift() {
		b.Owns(&buildv1.BuildConfig{}).Owns(&imagev1.ImageStream{})
	}
	return b.Complete(r)
}

// onResourceChange triggers hooks when a resource is changed
func (r *KogitoBuildReconciler) onResourceChange(instance *appv1beta1.KogitoBuild, resourceType reflect.Type, resources []resource.KubernetesResource) error {
	// add other resources if need
	switch resourceType {
	case reflect.TypeOf(buildv1.BuildConfig{}):
		return r.onBuildConfigChange(instance, resources)
	}
	return nil
}

// onBuildConfigChange triggers when a build config changes
func (r *KogitoBuildReconciler) onBuildConfigChange(instance *appv1beta1.KogitoBuild, buildConfigs []resource.KubernetesResource) error {
	// triggers only on source builds
	if instance.Spec.Type == appv1beta1.RemoteSourceBuildType ||
		instance.Spec.Type == appv1beta1.LocalSourceBuildType {
		for _, bc := range buildConfigs {
			// building from source
			if bc.GetName() == build.GetBuildBuilderName(instance) {
				r.Log.Info("Changes detected for build config, starting again", "Build Config", bc.GetName())
				if err := build.StartNewBuild(bc.(*buildv1.BuildConfig), r.Client); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
