package builder

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var log = logger.GetLogger("builder_kogitoapp")

// KogitoAppInventory has a reference for every resource needed to deploy the KogitoApp
type KogitoAppInventory struct {
	ResourceInventoryStatus
	ServiceAccount     *corev1.ServiceAccount
	RoleBinding        *rbacv1.RoleBinding
	BuildConfigS2I     *buildv1.BuildConfig
	BuildConfigService *buildv1.BuildConfig
	DeploymentConfig   *appsv1.DeploymentConfig
	Route              *routev1.Route
	Service            *corev1.Service
}

// ResourceInventoryStatusKind defines the kind of the resource status in the cluster
type ResourceInventoryStatusKind struct {
	IsNew bool
}

// ResourceInventoryStatus defines the resource status in the cluster
type ResourceInventoryStatus struct {
	ServiceAccountStatus     ResourceInventoryStatusKind
	RoleBindingStatus        ResourceInventoryStatusKind
	BuildConfigS2IStatus     ResourceInventoryStatusKind
	BuildConfigServiceStatus ResourceInventoryStatusKind
	DeploymentConfigStatus   ResourceInventoryStatusKind
	RouteStatus              ResourceInventoryStatusKind
	ServiceStatus            ResourceInventoryStatusKind
}
