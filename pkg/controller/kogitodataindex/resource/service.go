package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newService creates a Service resource based on deployment spec. It's expected that the container exposes at least one port to be able to create a valid service
func newService(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) *corev1.Service {
	ports := extractPortsFromDeployment(statefulset)
	if len(ports) == 0 {
		// a service without port to expose doesn't exist
		log.Warnf("The deployment spec '%s' doesn't have any ports exposed. Won't be possible to create a new service.", statefulset.Name)
		return nil
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.Name,
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: statefulset.Spec.Selector.MatchLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	addDefaultMetadata(&svc.ObjectMeta, instance)
	meta.SetGroupVersionKind(&svc.TypeMeta, meta.KindService)

	statefulset.Spec.ServiceName = svc.Name

	return &svc
}

//TODO: util function that can be extract from this file
func extractPortsFromDeployment(statefulset *appsv1.StatefulSet) []corev1.ServicePort {
	ports := []corev1.ServicePort{}

	// for now, we should have at least 1 container defined.
	if len(statefulset.Spec.Template.Spec.Containers) == 0 ||
		len(statefulset.Spec.Template.Spec.Containers[0].Ports) == 0 {
		return ports
	}

	for _, port := range statefulset.Spec.Template.Spec.Containers[0].Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.ContainerPort,
			TargetPort: intstr.FromInt(int(port.ContainerPort)),
		},
		)
	}
	return ports
}
