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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// FromEnvToEnvVar Function to convert an array of Env parameters to Kube Core EnvVar
func FromEnvToEnvVar(envs []v1alpha1.Env) (envVars []corev1.EnvVar) {
	if &envs == nil {
		return nil
	}

	for _, env := range envs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	return envVars
}

// FromResourcesToResourcesRequirements Convert the exposed data structure Resources to Kube Core ResourceRequirements
func FromResourcesToResourcesRequirements(resources v1alpha1.Resources) (resReq corev1.ResourceRequirements) {
	if &resources == nil {
		return corev1.ResourceRequirements{}
	}
	if len(resources.Limits) == 0 && len(resources.Requests) == 0 {
		return corev1.ResourceRequirements{}
	}
	resReq = corev1.ResourceRequirements{}
	// only build what is need to not conflict with DeepCopy later
	if len(resources.Limits) > 0 {
		resReq.Limits = corev1.ResourceList{}
	}
	if len(resources.Requests) > 0 {
		resReq.Requests = corev1.ResourceList{}
	}

	// we have to enforce the string conversion first, so we normalize the value: https://issues.jboss.org/browse/KOGITO-415
	for _, limit := range resources.Limits {
		rawValue := resource.MustParse(limit.Value)
		resReq.Limits[corev1.ResourceName(limit.Resource)] = resource.MustParse(rawValue.String())
	}

	for _, request := range resources.Requests {
		rawValue := resource.MustParse(request.Value)
		resReq.Requests[corev1.ResourceName(request.Resource)] = resource.MustParse(rawValue.String())
	}

	return resReq
}

// ContainsResource checks whether or not the resource is presented in resources
func ContainsResource(resource v1alpha1.ResourceKind, resources []v1alpha1.ResourceMap) bool {
	for _, res := range resources {
		if res.Resource == resource {
			return true
		}
	}
	return false
}
