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
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/services/kogitobuild"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"

	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
)

const (
	imageStreamCreationReconcileTimeout = 10 * time.Second
)

// KogitoBuildReconciler reconciles a KogitoBuild object
type KogitoBuildReconciler struct {
	Client *kogitocli.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
}

// Reconcile reads that state of the cluster for a KogitoBuild object and makes changes based on the state read
// and what is in the KogitoBuild.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.

func (r *KogitoBuildReconciler) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	r.Log.Infof("Reconciling KogitoBuild for %s in %s", request.Name, request.Namespace)

	// fetch the requested instance
	instance := &appv1alpha1.KogitoBuild{}
	exists, err := kubernetes.ResourceC(r.Client).FetchWithKey(request.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{Requeue: false}, err
	}
	if !exists {
		return reconcile.Result{Requeue: false}, nil
	}

	result = reconcile.Result{Requeue: false}

	defer kogitobuild.HandleStatusChange(instance, r.Client, &resultErr)

	if len(instance.Spec.Runtime) == 0 {
		instance.Spec.Runtime = appv1alpha1.QuarkusRuntimeType
	}
	if len(instance.Spec.TargetKogitoRuntime) == 0 {
		instance.Spec.TargetKogitoRuntime = instance.Name
	}

	// create the Kogito Image Streams to build the service if needed
	created, resultErr := kogitobuild.CreateRequiredKogitoImageStreams(instance, r.Client)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	buildMgr, resultErr := kogitobuild.New(instance, r.Client, r.Scheme)
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
	writer := write.New(r.Client.ControlCli)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		r.Log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
		_, resultErr = writer.AddResources(delta.Added)
		if resultErr != nil {
			return
		}
		_, resultErr = writer.UpdateResources(deployed[resourceType], delta.Updated)
		if resultErr != nil {
			return
		}
		_, resultErr = writer.RemoveResources(delta.Removed)
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

// onResourceChange triggers hooks when a resource is changed
func (r *KogitoBuildReconciler) onResourceChange(instance *appv1alpha1.KogitoBuild, resourceType reflect.Type, resources []resource.KubernetesResource) error {
	// add other resources if need
	switch resourceType {
	case reflect.TypeOf(buildv1.BuildConfig{}):
		return r.onBuildConfigChange(instance, resources)
	}
	return nil
}

// onBuildConfigChange triggers when a build config changes
func (r *KogitoBuildReconciler) onBuildConfigChange(instance *appv1alpha1.KogitoBuild, buildConfigs []resource.KubernetesResource) error {
	// triggers only on source builds
	if instance.Spec.Type == appv1alpha1.RemoteSourceBuildType ||
		instance.Spec.Type == appv1alpha1.LocalSourceBuildType {
		for _, bc := range buildConfigs {
			// building from source
			if bc.GetName() == kogitobuild.GetBuildBuilderName(instance) {
				r.Log.Infof("Changes detected in BuildConfig %s, starting new build", bc.GetName())
				if err := kogitobuild.StartNewBuild(bc.(*buildv1.BuildConfig), r.Client); err != nil {
					return err
				}
			}
		}
	}
	return nil
}


func (r *KogitoBuildReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoBuild controller")
	// Create a new controller
	c, err := controller.New("kogitobuild-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoBuild
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoBuild{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchedObjects := []framework.WatchedObjects{
		{
			GroupVersion: buildv1.GroupVersion,
			AddToScheme:  buildv1.Install,
			Objects:      []runtime.Object{&buildv1.BuildConfig{}},
		},
		{
			GroupVersion: imagev1.GroupVersion,
			AddToScheme:  imagev1.Install,
			Objects:      []runtime.Object{&imagev1.ImageStream{}},
		},
	}
	controllerWatcher := framework.NewControllerWatcher(r.Client, mgr, c, &appv1alpha1.KogitoBuild{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}
