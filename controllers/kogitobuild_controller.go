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

package controllers

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitobuild"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/internal"
	"github.com/kiegroup/kogito-operator/version"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/kiegroup/kogito-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	appv1beta1 "github.com/kiegroup/kogito-operator/api/v1beta1"
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

// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitobuilds/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=build.openshift.io,resources=builds;buildconfigs,verbs=get;create;list;watch;delete;update
// +kubebuilder:rbac:groups=image.openshift.io,resources=imagestreams;imagestreamtags,verbs=get;create;list;watch;delete;update

// Reconcile reads that state of the cluster for a KogitoBuild object and makes changes based on the state read
// and what is in the KogitoBuild.Spec
func (r *KogitoBuildReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, resultErr error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling for KogitoBuild")

	// create context
	context := &operator.Context{
		Client:  r.Client,
		Log:     log,
		Scheme:  r.Scheme,
		Version: version.Version,
	}

	// fetch the requested instance
	buildInstanceHandler := internal.NewKogitoBuildHandler(context)
	instance, resultErr := buildInstanceHandler.FetchKogitoBuildInstance(req.NamespacedName)
	if resultErr != nil {
		return
	} else if instance == nil {
		log.Warn("Kogito Build not found")
		return
	}

	buildStatusHandler := kogitobuild.NewStatusHandler(context)
	defer buildStatusHandler.HandleStatusChange(instance, resultErr)

	if len(instance.GetSpec().GetRuntime()) == 0 {
		instance.GetSpec().SetRuntime(api.QuarkusRuntimeType)
	}
	envs := instance.GetSpec().GetEnv()
	instance.GetSpec().SetEnv(framework.EnvOverride(envs, corev1.EnvVar{Name: infrastructure.RuntimeTypeKey, Value: string(instance.GetSpec().GetRuntime())}))
	if len(instance.GetSpec().GetTargetKogitoRuntime()) == 0 {
		instance.GetSpec().SetTargetKogitoRuntime(instance.GetName())
	}

	// create the Kogito Image Streams to build the service if needed
	buildImageHandler := kogitobuild.NewImageSteamHandler(context)
	created, resultErr := buildImageHandler.CreateRequiredKogitoImageStreams(instance)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	deltaProcessor, resultErr := kogitobuild.NewDeltaProcessor(context, instance)
	if resultErr != nil {
		return
	}
	resultErr = deltaProcessor.ProcessDelta()
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
