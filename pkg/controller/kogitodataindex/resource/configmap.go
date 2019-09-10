package resource

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newProtobufConfigMap(instance *v1alpha1.KogitoDataIndex) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-protobufs", instance.Spec.Name),
			Namespace: instance.Namespace,
		},
		// those values will be populated later by the user with proto files
		Data: map[string]string{},
	}

	meta.SetGroupVersionKind(&cm.TypeMeta, meta.KindConfigMap)
	addDefaultMetadata(&cm.ObjectMeta, instance)

	return &cm
}
