package definitions

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultAnnotations = map[string]string{
	"org.kie.kogito/managed-by":   "Kogito Operator",
	"org.kie.kogito/operator-crd": "KogitoApp",
}

// DefinitionKind is a resource kind representation from a Kubernetes/Openshift cluster
type DefinitionKind string

const (
	// ServiceKind for service
	ServiceKind DefinitionKind = "Service"
	// BuildConfigKind for a buildConfig
	BuildConfigKind DefinitionKind = "BuildConfig"
	// DeploymentConfigKind for a DeploymentConfig
	DeploymentConfigKind DefinitionKind = "DeploymentConfig"
	// RoleBindingKind for a RoleBinding
	RoleBindingKind DefinitionKind = "RoleBinding"
	// ServiceAccountKind for a ServiceAccount
	ServiceAccountKind DefinitionKind = "ServiceAccount"
	// RouteKind for a Route
	RouteKind DefinitionKind = "Route"
)

const (
	labelAppName = "app"
)

// addDefaultMeta adds the default annotations and labels
func addDefaultMeta(objectMeta *metav1.ObjectMeta, kogitoApp *v1alpha1.KogitoApp) {
	if objectMeta != nil {
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = map[string]string{}
		}
		if objectMeta.Labels == nil {
			objectMeta.Labels = map[string]string{}
		}
		for key, value := range defaultAnnotations {
			objectMeta.Annotations[key] = value
		}
		//objectMeta.Labels[labelAppName] = kogitoApp.Spec.Name
		addDefaultLabels(&objectMeta.Labels, kogitoApp)
	}
}

// addDefaultLabels adds the default labels
func addDefaultLabels(m *map[string]string, kogitoApp *v1alpha1.KogitoApp) {
	if *m == nil {
		(*m) = map[string]string{}
	}
	(*m)[labelAppName] = kogitoApp.Spec.Name
}

func setGroupVersionKind(typeMeta *metav1.TypeMeta, kind DefinitionKind) {
	typeMeta.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind(string(kind)))
}
