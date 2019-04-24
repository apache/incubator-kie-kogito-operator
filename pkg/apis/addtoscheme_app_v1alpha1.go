package apis

import (
	"github.com/kiegroup/submarine-cloud-operator/pkg/apis/app/v1alpha1"
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha1.SchemeBuilder.AddToScheme,
		oappsv1.SchemeBuilder.AddToScheme,
		routev1.SchemeBuilder.AddToScheme,
		oimagev1.SchemeBuilder.AddToScheme,
		buildv1.SchemeBuilder.AddToScheme,
	)
}
