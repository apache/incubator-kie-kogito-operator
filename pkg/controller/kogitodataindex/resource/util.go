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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func extractResources(instance *v1alpha1.KogitoDataIndex) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if len(instance.Spec.CPULimit) > 0 {
		resources.Limits[corev1.ResourceCPU] = resource.MustParse(instance.Spec.CPULimit)
	}

	if len(instance.Spec.MemoryLimit) > 0 {
		resources.Limits[corev1.ResourceMemory] = resource.MustParse(instance.Spec.MemoryLimit)
	}

	if len(instance.Spec.CPURequest) > 0 {
		resources.Requests[corev1.ResourceCPU] = resource.MustParse(instance.Spec.CPURequest)
	}

	if len(instance.Spec.MemoryRequest) > 0 {
		resources.Requests[corev1.ResourceMemory] = resource.MustParse(instance.Spec.MemoryRequest)
	}

	// ensuring equality with the API
	if len(resources.Limits) == 0 {
		resources.Limits = nil
	}
	if len(resources.Requests) == 0 {
		resources.Requests = nil
	}

	return resources
}
