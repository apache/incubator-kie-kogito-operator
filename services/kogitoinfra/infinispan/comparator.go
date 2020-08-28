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

package infinispan

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

// GetComparators prepare the functions to perform comparisons between Kubernetes resources
func GetComparators() []framework.Comparator {
	resourceComparator := compare.DefaultComparator()
	return []framework.Comparator{
		createSecretComparator(resourceComparator),
		createInfinispanComparator(),
	}
}

func createSecretComparator(resourceComparator compare.ResourceComparator) framework.Comparator {
	secretType := reflect.TypeOf(v1.Secret{})
	defaultSecretComparator := resourceComparator.GetComparator(secretType)

	return framework.Comparator{
		ResourceType: secretType,
		CompFunc: func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
			if !reflect.DeepEqual(deployed.GetAnnotations()[annotationKeyMD5], requested.GetAnnotations()[annotationKeyMD5]) {
				return false
			}
			return defaultSecretComparator(deployed, requested)
		},
	}
}

func createInfinispanComparator() framework.Comparator {
	return framework.Comparator{
		ResourceType: reflect.TypeOf(infinispanv1.Infinispan{}),
		CompFunc: func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
			infinispanDep := deployed.(*infinispanv1.Infinispan)
			infinispanReq := requested.(*infinispanv1.Infinispan).DeepCopy()

			return reflect.DeepEqual(infinispanDep.Spec.Security.EndpointSecretName, infinispanReq.Spec.Security.EndpointSecretName) ||
				reflect.DeepEqual(infinispanDep.Name, infinispanReq.Name)
		},
	}
}
