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
	"github.com/kiegroup/kogito-cloud-operator/core/kogitoinfra"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
)

// KogitoInfraReconciler reconciles a KogitoInfra object
type KogitoInfraReconciler struct {
	*client.Client
	Log    logger.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=infinispan.org,resources=infinispans,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=sources.knative.dev,resources=sinkbindings,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
// +kubebuilder:rbac:groups=mongodb.com,resources=mongodb,verbs=get;create;list;watch;delete

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
func (r *KogitoInfraReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("name", req.Name, "namespace", req.Namespace)
	log.Info("Reconciling KogitoInfra")

	// create context
	context := &operator.Context{
		Client: r.Client,
		Log:    log,
		Scheme: r.Scheme,
	}

	// Fetch the KogitoInfra instance
	infraHandler := internal.NewKogitoInfraHandler(context)
	instance, resultErr := infraHandler.FetchKogitoInfraInstance(req.NamespacedName)
	if resultErr != nil {
		return reconcile.Result{}, resultErr
	}
	if instance == nil {
		log.Debug("KogitoInfra instance not found")
		return reconcile.Result{}, nil
	}

	statusHandler := kogitoinfra.NewStatusHandler(context)
	defer statusHandler.UpdateBaseStatus(instance, &resultErr)

	reconcilerHandler := kogitoinfra.NewReconcilerHandler(context)
	reconciler, resultErr := reconcilerHandler.GetInfraReconciler(instance, r.Scheme)
	if resultErr != nil {
		return reconcilerHandler.GetReconcileResultFor(resultErr, false)
	}

	requeue, resultErr := reconciler.Reconcile()
	return reconcilerHandler.GetReconcileResultFor(resultErr, requeue)
}

// SetupWithManager registers the controller with manager
func (r *KogitoInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoInfra controller")
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// do not call reconciler if only owner reference is added or deleted
			return reflect.DeepEqual(e.MetaNew.GetOwnerReferences(), e.MetaOld.GetOwnerReferences())
		},
	}
	b := ctrl.NewControllerManagedBy(mgr).For(&v1beta1.KogitoInfra{}, builder.WithPredicates(pred))
	b = kogitoinfra.AppendInfinispanWatchedObjects(b)
	b = kogitoinfra.AppendKafkaWatchedObjects(b)
	b = kogitoinfra.AppendKeycloakWatchedObjects(b)
	b = kogitoinfra.AppendMongoDBWatchedObjects(b)
	return b.Complete(r)
}
