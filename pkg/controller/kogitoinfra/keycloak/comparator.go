// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package keycloak

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"reflect"
)

// GetComparators gets the comparator for Keycloak resources
func GetComparators() []framework.Comparator {
	return []framework.Comparator{createKeycloakComparator(), createKeycloakRealmComparator()}
}

func createKeycloakComparator() framework.Comparator {
	return framework.Comparator{
		ResourceType: reflect.TypeOf(keycloakv1alpha1.Keycloak{}),
		CompFunc: func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
			keycloakDep := deployed.(*keycloakv1alpha1.Keycloak)
			keycloakReq := requested.(*keycloakv1alpha1.Keycloak).DeepCopy()
			// we just care for the instance name, other attributes can be changed at will by the user
			return reflect.DeepEqual(keycloakDep.Name, keycloakReq.Name) && reflect.DeepEqual(keycloakDep.Labels, keycloakReq.Labels)
		},
	}
}

func createKeycloakRealmComparator() framework.Comparator {
	return framework.Comparator{
		ResourceType: reflect.TypeOf(keycloakv1alpha1.KeycloakRealm{}),
		CompFunc: func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
			defaultComparatorFunc := compare.DefaultComparator().GetDefaultComparator()

			keycloakRealmDep := deployed.(*keycloakv1alpha1.KeycloakRealm)
			keycloakRealmReq := requested.(*keycloakv1alpha1.KeycloakRealm).DeepCopy()

			if !reflect.DeepEqual(keycloakRealmDep.Name, keycloakRealmReq.Name) ||
				!reflect.DeepEqual(keycloakRealmDep.Labels, keycloakRealmReq.Labels) {
				return false
			}

			return defaultComparatorFunc(deployed, requested)
		},
	}
}
