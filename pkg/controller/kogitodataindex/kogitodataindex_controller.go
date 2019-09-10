package kogitodataindex

import (
	"fmt"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appv1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	commonres "github.com/kiegroup/kogito-cloud-operator/pkg/resource"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("controller_kogitodataindex")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KogitoDataIndex Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	discover, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		panic(fmt.Sprintf("Error getting discovery client: %v", err))
	}

	return &ReconcileKogitoDataIndex{
		client: &client.Client{
			ControlCli: mgr.GetClient(),
			Discovery:  discover},
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kogitodataindex-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KogitoDataIndex
	err = c.Watch(&source.Kind{Type: &appv1alpha1.KogitoDataIndex{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchOwnedObjects := []runtime.Object{
		&corev1.ConfigMap{},
		&corev1.Service{},
		&appsv1.StatefulSet{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.KogitoDataIndex{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}
	return nil
}

// blank assignment to verify that ReconcileKogitoDataIndex implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKogitoDataIndex{}

// ReconcileKogitoDataIndex reconciles a KogitoDataIndex object
type ReconcileKogitoDataIndex struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client *client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KogitoDataIndex object and makes changes based on the state read
// and what is in the KogitoDataIndex.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKogitoDataIndex) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.With("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KogitoDataIndex")

	instances := &appv1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(r.client).ListWithNamespace(request.Namespace, instances); err != nil {
		return reconcile.Result{}, err
	}

	if len(instances.Items) > 1 {
		return reconcile.Result{RequeueAfter: time.Duration(5) * time.Minute},
			fmt.Errorf("There's more than one KogitoDataIndex resource in this namespace, please delete one of them")
	}

	// Fetch the KogitoDataIndex instance
	instance := &appv1alpha1.KogitoDataIndex{}
	if exists, err := kubernetes.ResourceC(r.client).FetchWithKey(request.NamespacedName, instance); err != nil {
		return reconcile.Result{}, err
	} else if !exists {
		return reconcile.Result{}, nil
	}

	// Create our inventory
	reqLogger.Infof("Ensure Kogito Data Index '%s' resources are created", instance.Spec.Name)
	resources, err := resource.CreateOrFetchResources(instance, commonres.FactoryContext{
		Client: r.client,
		PreCreate: func(object meta.ResourceObject) error {
			if object != nil {
				return controllerutil.SetControllerReference(instance, object, r.scheme)
			}
			return nil
		},
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	if !resources.StatefulSetStatus.New {
		reqLogger.Infof("Handling changes in Kogito Data Index '%s'", instance.Spec.Name)
		if err = resource.ManageResources(instance, &resources, r.client); err != nil {
			return reconcile.Result{}, err
		}
	}

	reqLogger.Infof("Handling Status updates on '%s'", instance.Spec.Name)
	if err = status.ManageStatus(instance, &resources, r.client); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
