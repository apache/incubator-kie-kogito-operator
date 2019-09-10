package meta

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	coreappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
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
	KindBuildConfig = DefinitionKind{"BuildConfig", true, buildv1.SchemeGroupVersion}
	// KindDeploymentConfig for a DeploymentConfig
	KindDeploymentConfig = DefinitionKind{"DeploymentConfig", true, appsv1.SchemeGroupVersion}
	// KindRoute for a Route
	KindRoute = DefinitionKind{"Route", true, routev1.SchemeGroupVersion}
	// KindImageStreamTag for a ImageStreamTag
	KindImageStreamTag = DefinitionKind{"ImageStreamTag", true, imgv1.SchemeGroupVersion}
	// KindImageStream for a ImageStream
	KindImageStream = DefinitionKind{"ImageStream", true, imgv1.SchemeGroupVersion}
	// KindBuildRequest for a BuildRequest
	KindBuildRequest = DefinitionKind{"BuildRequest", true, buildv1.SchemeGroupVersion}
	// KindNamespace for a Namespace
	KindNamespace = DefinitionKind{"Namespace", false, corev1.SchemeGroupVersion}
	// KindCRD for a CustomResourceDefinition
	KindCRD = DefinitionKind{"CustomResourceDefinition", false, apiextensionsv1beta1.SchemeGroupVersion}
	// KindKogitoApp for a KogitoApp controller
	KindKogitoApp = DefinitionKind{"KogitoApp", false, v1alpha1.SchemeGroupVersion}
	// KindConfigMap for a ConfigMap
	KindConfigMap = DefinitionKind{"ConfigMap", false, corev1.SchemeGroupVersion}
	// KindDeployment for a Deployment
	KindDeployment = DefinitionKind{"Deployment", false, coreappsv1.SchemeGroupVersion}
	// KindStatefulSet for a StatefulSet
	KindStatefulSet = DefinitionKind{"StatefulSet", false, coreappsv1.SchemeGroupVersion}
)

// SetGroupVersionKind sets the group, version and kind for the resource
func SetGroupVersionKind(typeMeta *metav1.TypeMeta, kind DefinitionKind) {
	typeMeta.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kind.GroupVersion.Group,
		Version: kind.GroupVersion.Version,
		Kind:    kind.Name,
	})
}

// GetRegisteredSchema gets all schema and types registered for use with CLI, unit tests, custom clients and so on
func GetRegisteredSchema() *runtime.Scheme {
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Namespace{})
	s.AddKnownTypes(apiextensionsv1beta1.SchemeGroupVersion, &apiextensionsv1beta1.CustomResourceDefinition{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.KogitoApp{}, &v1alpha1.KogitoAppList{})

	return s
}
