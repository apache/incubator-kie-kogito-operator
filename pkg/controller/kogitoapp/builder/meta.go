package builder

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultAnnotations = map[string]string{
	"org.kie.kogito/managed-by":   "Kogito Operator",
	"org.kie.kogito/operator-crd": "KogitoApp",
}

const (
	// LabelKeyAppName is the default label added to all resources
	LabelKeyAppName = "app"
	// LabelKeyServiceName is the default label added to the service
	LabelKeyServiceName = "service"
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
		addDefaultLabels(&objectMeta.Labels, kogitoApp)
	}
}

// addDefaultLabels adds the default labels
func addDefaultLabels(m *map[string]string, kogitoApp *v1alpha1.KogitoApp) {
	if *m == nil {
		(*m) = map[string]string{}
	}
	(*m)[LabelKeyAppName] = kogitoApp.Spec.Name
}

// addServiceLabels adds the service labels
func addServiceLabels(objectMeta *metav1.ObjectMeta, kogitoApp *v1alpha1.KogitoApp) {
	if objectMeta != nil {
		if objectMeta.Labels == nil {
			objectMeta.Labels = map[string]string{}
		}

		if kogitoApp.Spec.Service.Labels == nil {
			objectMeta.Labels[LabelKeyServiceName] = kogitoApp.Spec.Name
		} else {
			for key, value := range kogitoApp.Spec.Service.Labels {
				objectMeta.Labels[key] = value
			}
		}

	}
}
