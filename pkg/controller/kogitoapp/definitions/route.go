package definitions

import (
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewRoute creates a new Route resource based on the Service created for the KogitoApp container
func NewRoute(kogitoApp *v1alpha1.KogitoApp, service *corev1.Service) (route *routev1.Route, err error) {
	if service == nil {
		return route, fmt.Errorf("Impossible to create a Route without a service on Kogito app %s", kogitoApp.Name)
	}
	route = &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(defaultExportedProtocol),
			},
			To: routev1.RouteTargetReference{
				Kind: KindService.Name,
				Name: service.Name,
			},
		},
	}
	addDefaultMeta(&route.ObjectMeta, kogitoApp)
	SetGroupVersionKind(&route.TypeMeta, KindRoute)
	route.ResourceVersion = ""
	return route, nil
}
