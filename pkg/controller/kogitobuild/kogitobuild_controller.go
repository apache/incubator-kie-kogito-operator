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

package kogitobuild

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitobuild/build"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logger.GetLogger("kogitoruntime_controller")

const (
	imageStreamCreationReconcileTimeout = 10 * time.Second
)

// Add creates a new KogitoBuild Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKogitoBuild{client: kogitocli.NewForController(mgr.GetConfig(), mgr.GetClient()), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Debug("Adding watched objects for KogitoBuild controller")
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
	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileKogitoBuild).client, mgr, c, &appv1alpha1.KogitoBuild{})
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoBuild implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoBuild{}

// ReconcileKogitoBuild reconciles a KogitoBuild object
type ReconcileKogitoBuild struct {
	client *kogitocli.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoBuild object and makes changes based on the state read
// and what is in the KogitoBuild.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoBuild) Reconcile(request reconcile.Request) (result reconcile.Result, resultErr error) {
	log.Infof("Reconciling KogitoRuntime for %s in %s", request.Name, request.Namespace)

	// fetch the requested instance
	instance := &appv1alpha1.KogitoBuild{}
	exists, err := kubernetes.ResourceC(r.client).FetchWithKey(request.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{Requeue: false}, err
	}
	if !exists {
		return reconcile.Result{Requeue: false}, nil
	}

	result = reconcile.Result{Requeue: false}

	defer r.handleStatusChange(instance, &resultErr)

	if len(instance.Spec.Runtime) == 0 {
		instance.Spec.Runtime = appv1alpha1.QuarkusRuntimeType
	}
	if len(instance.Spec.TargetKogitoRuntime) == 0 {
		instance.Spec.TargetKogitoRuntime = instance.Name
	}

	// create the Kogito Image Streams to build the service if needed
	created, resultErr := build.CreateRequiredKogitoImageStreams(instance, r.client)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	buildMgr, resultErr := build.New(instance, r.client, r.scheme)
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
	writer := write.New(r.client.ControlCli)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
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
func (r *ReconcileKogitoBuild) onResourceChange(instance *appv1alpha1.KogitoBuild, resourceType reflect.Type, resources []resource.KubernetesResource) error {
	// add other resources if need
	switch resourceType {
	case reflect.TypeOf(buildv1.BuildConfig{}):
		return r.onBuildConfigChange(instance, resources)
	}
	return nil
}

// onBuildConfigChange triggers when a build config changes
func (r *ReconcileKogitoBuild) onBuildConfigChange(instance *appv1alpha1.KogitoBuild, buildConfigs []resource.KubernetesResource) error {
	// triggers only on source builds
	if instance.Spec.Type == appv1alpha1.RemoteSourceBuildType ||
		instance.Spec.Type == appv1alpha1.LocalSourceBuildType {
		for _, bc := range buildConfigs {
			// building from source
			if bc.GetName() == build.GetBuildBuilderName(instance) {
				log.Infof("Changes detected in BuildConfig %s, starting new build", bc.GetName())
				if err := build.StartNewBuild(bc.(*buildv1.BuildConfig), r.client); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ReconcileKogitoBuild) update(instance *appv1alpha1.KogitoBuild) error {
	if err := kubernetes.ResourceC(r.client).Update(instance); err != nil {
		return err
	}
	return nil
}
