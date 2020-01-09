// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"fmt"
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newRoute creates a new Route resource based on the Service created for the KogitoApp container
func newRoute(kogitoApp *v1alpha1.KogitoApp, service *corev1.Service) (route *routev1.Route, err error) {
	if service == nil {
		return route, fmt.Errorf("Impossible to create a Route without a service on Kogito app %s", kogitoApp.Name)
	}
	route = &routev1.Route{
		ObjectMeta: *service.ObjectMeta.DeepCopy(),
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(framework.DefaultExportedPort),
			},
			To: routev1.RouteTargetReference{
				Kind: meta.KindService.Name,
				Name: service.Name,
			},
		},
	}
	addDefaultMeta(&route.ObjectMeta, kogitoApp)
	meta.SetGroupVersionKind(&route.TypeMeta, meta.KindRoute)
	route.ResourceVersion = ""
	return route, nil
}
