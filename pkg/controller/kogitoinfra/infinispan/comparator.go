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
	v1 "k8s.io/api/core/v1"
	"reflect"
)

// GetComparator prepare the functions to perform comparisons between Kubernetes resources
func GetComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()
	resourceComparator.SetComparator(createSecretComparator(resourceComparator))
	resourceComparator.SetComparator(createInfinispanComparator())

	return compare.MapComparator{Comparator: resourceComparator}
}

func createSecretComparator(resourceComparator compare.ResourceComparator) (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	secretType := reflect.TypeOf(v1.Secret{})
	defaultSecretComparator := resourceComparator.GetComparator(secretType)

	return secretType, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		if !reflect.DeepEqual(deployed.GetAnnotations()[annotationKeyMD5], requested.GetAnnotations()[annotationKeyMD5]) {
			return false
		}
		return defaultSecretComparator(deployed, requested)
	}
}

func createInfinispanComparator() (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	return reflect.TypeOf(infinispanv1.Infinispan{}), func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		infinispanDep := deployed.(*infinispanv1.Infinispan)
		infinispanReq := requested.(*infinispanv1.Infinispan).DeepCopy()

		return reflect.DeepEqual(infinispanDep.Spec.Security.EndpointSecretName, infinispanReq.Spec.Security.EndpointSecretName)
	}
}
