package resource

import (
	"fmt"

	routev1 "github.com/openshift/api/route/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newRoute creates a new Route resource based on the Service created for the KogitoApp container
func newRoute(instance *v1alpha1.KogitoDataIndex, service *corev1.Service) (route *routev1.Route, err error) {
	if service == nil {
		return route, fmt.Errorf("Impossible to create a Route without a service on Kogito Data Index %s", instance.Name)
	}

	route = &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(defaultExposedPort),
			},
			To: routev1.RouteTargetReference{
				Kind: meta.KindService.Name,
				Name: service.Name,
			},
		},
	}

	addDefaultMetadata(&route.ObjectMeta, instance)
	meta.SetGroupVersionKind(&route.TypeMeta, meta.KindRoute)
	route.ResourceVersion = ""
	return route, nil
}
