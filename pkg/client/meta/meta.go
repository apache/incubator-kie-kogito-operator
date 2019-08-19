package meta

import (
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DefinitionKind is a resource kind representation from a Kubernetes/Openshift cluster
type DefinitionKind struct {
	// Name of the resource
	Name string
	// IsFromOpenShift identifies if this Resource only exists on OpenShift cluster
	IsFromOpenShift bool
	// Identifies the group version for the OpenShift APIs
	GroupVersion schema.GroupVersion
}

var (
	// KindService for service
	KindService = DefinitionKind{"Service", false, corev1.SchemeGroupVersion}
	// KindBuildConfig for a buildConfig
	KindBuildConfig = DefinitionKind{"BuildConfig", true, buildv1.GroupVersion}
	// KindDeploymentConfig for a DeploymentConfig
	KindDeploymentConfig = DefinitionKind{"DeploymentConfig", true, appsv1.GroupVersion}
	// KindRoleBinding for a RoleBinding
	KindRoleBinding = DefinitionKind{"RoleBinding", false, corev1.SchemeGroupVersion}
	// KindServiceAccount for a ServiceAccount
	KindServiceAccount = DefinitionKind{"ServiceAccount", false, corev1.SchemeGroupVersion}
	// KindRoute for a Route
	KindRoute = DefinitionKind{"Route", true, routev1.SchemeGroupVersion}
	// KindImageStreamTag for a ImageStreamTag
	KindImageStreamTag = DefinitionKind{"ImageStreamTag", true, imgv1.SchemeGroupVersion}
	// KindBuildRequest for a BuildRequest
	KindBuildRequest = DefinitionKind{"BuildRequest", true, buildv1.SchemeGroupVersion}
	// KindNamespace for a Namespace
	KindNamespace = DefinitionKind{"Namespace", false, corev1.SchemeGroupVersion}
)

// SetGroupVersionKind sets the group, version and kind for the resource
func SetGroupVersionKind(typeMeta *metav1.TypeMeta, kind DefinitionKind) {
	typeMeta.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kind.GroupVersion.Group,
		Version: kind.GroupVersion.Version,
		Kind:    kind.Name,
	})
}
