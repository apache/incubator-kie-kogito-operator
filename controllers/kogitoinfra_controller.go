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
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
)

const reconciliationStandardInterval = time.Second * 30

// KogitoInfraReconciler reconciles a KogitoInfra object
type KogitoInfraReconciler struct {
	*client.Client
	Log    logger.Logger
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoInfra object and makes changes based on the state read
// and what is in the KogitoInfra.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.kiegroup.org,resources=kogitoinfras/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create;list;delete
// +kubebuilder:rbac:groups=infinispan.org,resources=infinispans,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkatopics,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=keycloak.org,resources=keycloaks,verbs=get;create;list;delete;watch
// +kubebuilder:rbac:groups=apps,resourceNames=kogito-operator,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=brokers,verbs=get;list;watch
// +kubebuilder:rbac:groups=eventing.knative.dev,resources=triggers,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=sources.knative.dev,resources=sinkbindings,verbs=get;list;watch;create;delete;update
// +kubebuilder:rbac:groups=integreatly.org,resources=grafanadashboards,verbs=get;create;list;watch;create;delete;update
func (r *KogitoInfraReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Reconciling for", "KogitoInfra", req.Name, "Namespace", req.Namespace)

	// Fetch the KogitoInfra instance
	instance, resultErr := infrastructure.FetchKogitoInfraInstance(r.Client, req.Name, req.Namespace)
	if resultErr != nil {
		return reconcile.Result{}, resultErr
	}
	if instance == nil {
		r.Log.Debug("KogitoInfra instance not found", "Name", req.Name, "Namespace", req.Namespace)
		return reconcile.Result{}, nil
	}

	defer r.updateBaseStatus(r.Client, instance, &resultErr)

	reconciler, resultErr := r.getKogitoInfraReconciler(r.Client, instance, r.Scheme)
	if resultErr != nil {
		return r.getReconcileResultFor(resultErr, false)
	}

	requeue, resultErr := reconciler.Reconcile()
	return r.getReconcileResultFor(resultErr, requeue)
}

// SetupWithManager registers the controller with manager
func (r *KogitoInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Debug("Adding watched objects for KogitoInfra controller")
	b := ctrl.NewControllerManagedBy(mgr).For(&v1beta1.KogitoInfra{})
	b = appendInfinispanWatchedObjects(b)
	b = appendKafkaWatchedObjects(b)
	b = appendKeycloakWatchedObjects(b)
	b = appendMongoDBWatchedObjects(b)
	return b.Complete(r)
}

func (r *KogitoInfraReconciler) getReconcileResultFor(err error, requeue bool) (reconcile.Result, error) {

	switch reasonForError(err) {
	case v1beta1.ReconciliationFailure:
		r.Log.Warn("Error while reconciling KogitoInfra", "error", err.Error())
		return reconcile.Result{RequeueAfter: 0, Requeue: false}, err
	case v1beta1.ResourceMissingResourceConfig, v1beta1.ResourceConfigError:
		r.Log.Error(err, "KogitoInfra configuration error")
		return reconcile.Result{RequeueAfter: 0, Requeue: false}, nil
	}

	// no requeue, no errors, stop reconciliation
	if !requeue && err == nil {
		r.Log.Debug("No need reconciliation for KogitoInfra")
		return reconcile.Result{RequeueAfter: 0, Requeue: false}, nil
	}
	// caller is asking for a reconciliation
	if err == nil {
		r.Log.Info("Waiting for all resources to be created, scheduling reconciliation. Scheduling reconciliation for", "Instance", reconciliationStandardInterval.String())
	} else { // reconciliation duo to a problem in the env (CRDs missing), infra deployments not ready, operators not installed.. etc. See errors.go
		r.Log.Info("Waiting for all resources to be created", "scheduling reconciliation:", err.Error(), "Scheduling reconciliation for", reconciliationStandardInterval.String())
	}
	return reconcile.Result{RequeueAfter: reconciliationStandardInterval}, nil
}
