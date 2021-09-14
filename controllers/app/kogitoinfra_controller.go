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
	"context"
	"github.com/kiegroup/kogito-operator/core/kogitoinfra"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/internal"
	"github.com/kiegroup/kogito-operator/version"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kiegroup/kogito-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kiegroup/kogito-operator/apis/v1beta1"
)

// KogitoInfraReconciler reconciles a KogitoInfra object
type KogitoInfraReconciler struct {
	*client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras/finalizers,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;create;delete;update
//+kubebuilder:rbac:groups=infinispan.org,resources=infinispans,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;create;list;delete;watch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch
//+kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=sources.knative.dev,resources=sinkbindings,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
//+kubebuilder:rbac:groups=mongodb.com,resources=mongodb,verbs=get;create;list;watch;delete

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
func (r *KogitoInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx)
	log.Info("Reconciling KogitoInfra")

	// create kogitoContext
	kogitoContext := operator.Context{
		Client:  r.Client,
		Log:     log,
		Scheme:  r.Scheme,
		Version: version.Version,
	}

	// Fetch the KogitoInfra instance
	infraHandler := internal.NewKogitoInfraHandler(kogitoContext)
	instance, err := infraHandler.FetchKogitoInfraInstance(req.NamespacedName)
	if err != nil {
		return reconcile.Result{}, err
	}
	if instance == nil {
		log.Debug("KogitoInfra instance not found")
		return reconcile.Result{}, nil
	}
	var resultErr error
	statusHandler := kogitoinfra.NewStatusHandler(kogitoContext, infraHandler)
	defer statusHandler.UpdateBaseStatus(instance, &resultErr)

	instance.GetStatus().SetEnvs(nil)
	instance.GetStatus().SetConfigMapEnvFromReferences(nil)
	instance.GetStatus().SetConfigMapVolumeReferences(nil)
	instance.GetStatus().SetSecretEnvFromReferences(nil)
	instance.GetStatus().SetSecretVolumeReferences(nil)

	reconcilerHandler := kogitoinfra.NewReconcilerHandler(kogitoContext)
	if !instance.GetSpec().IsResourceEmpty() {
		var reconciler kogitoinfra.Reconciler
		reconciler, resultErr = reconcilerHandler.GetInfraReconciler(instance)
		if resultErr != nil {
			return reconcilerHandler.GetReconcileResultFor(resultErr, false)
		}

		resultErr = reconciler.Reconcile()
		if resultErr != nil {
			return reconcilerHandler.GetReconcileResultFor(resultErr, false)
		}
	}

	appConfigMapReconciler := reconcilerHandler.GetInfraPropertiesReconciler(instance)
	if resultErr = appConfigMapReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	// Set envs in status
	if len(instance.GetSpec().GetEnvs()) > 0 {
		instance.GetStatus().AddEnvs(instance.GetSpec().GetEnvs())
	}

	configMapReferenceReconciler := reconcilerHandler.GetConfigMapReferenceReconciler(instance)
	if resultErr = configMapReferenceReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	secretReferenceReconciler := reconcilerHandler.GetSecretReferenceReconciler(instance)
	if resultErr = secretReferenceReconciler.Reconcile(); resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	return reconcile.Result{}, nil
}

// SetupWithManager registers the controller with manager
func (r *KogitoInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// do not call reconciler if only owner reference is added or deleted
			return reflect.DeepEqual(e.ObjectNew.GetOwnerReferences(), e.ObjectOld.GetOwnerReferences())
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).For(&v1beta1.KogitoInfra{}, builder.WithPredicates(pred))
	b = kogitoinfra.AppendInfinispanWatchedObjects(b)
	b = kogitoinfra.AppendKafkaWatchedObjects(b)
	b = kogitoinfra.AppendKeycloakWatchedObjects(b)
	b = kogitoinfra.AppendMongoDBWatchedObjects(b)
	b = kogitoinfra.AppendConfigMapWatchedObjects(b)
	b = kogitoinfra.AppendSecretWatchedObjects(b)
	return b.Complete(r)
}
