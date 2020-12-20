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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createRequiredRoute creates a new Route resource based on the given Service
func createRequiredRoute(instance v1beta1.KogitoService, service *corev1.Service) (route *routev1.Route) {
	if service == nil || len(service.Spec.Ports) == 0 {
		log.Warn("Impossible to create a Route without a target service on Kogito Service", "KogitoService", instance.GetName())
		return route
	}

	route = &routev1.Route{
		ObjectMeta: service.ObjectMeta,
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(service.Spec.Ports[0].Name),
			},
			To: routev1.RouteTargetReference{
				Kind: meta.KindService.Name,
				Name: service.Name,
			},
		},
	}

	route.ResourceVersion = ""
	return route
}
