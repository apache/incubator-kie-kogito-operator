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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/kogitobuild"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
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
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling for KogitoBuild")

	// fetch the requested instance
	buildInstanceHandler := internal.NewKogitoBuildHandler(r.Client, log)
	instance, resultErr := buildInstanceHandler.FetchKogitoBuildInstance(req.NamespacedName)
	if resultErr != nil {
		return
	} else if instance == nil {
		log.Info("Kogito Build not found")
		return
	}

	buildStatusHandler := kogitobuild.NewStatusHandler(r.Client, log)
	defer buildStatusHandler.HandleStatusChange(instance, resultErr)

	if len(instance.GetSpec().GetRuntime()) == 0 {
		instance.GetSpec().SetRuntime(api.QuarkusRuntimeType)
	}
	if len(instance.GetSpec().GetTargetKogitoRuntime()) == 0 {
		instance.GetSpec().SetTargetKogitoRuntime(instance.GetName())
	}

	// create the Kogito Image Streams to build the service if needed
	buildImageHandler := kogitobuild.NewImageSteamHandler(r.Client, log)
	created, resultErr := buildImageHandler.CreateRequiredKogitoImageStreams(instance)
	if resultErr != nil {
		return result, fmt.Errorf("Error while creating Kogito ImageStreams: %s ", resultErr)
	}
	if created {
		result = reconcile.Result{RequeueAfter: imageStreamCreationReconcileTimeout, Requeue: true}
		return
	}

	// get the build manager to start the reconciliation logic
	deltaProcessor, resultErr := kogitobuild.NewDeltaProcessor(instance, r.Client, r.Scheme, log)
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
