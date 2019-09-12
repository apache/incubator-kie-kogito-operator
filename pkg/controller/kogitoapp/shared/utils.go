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
func FromResourcesToResourcesRequirements(resources v1alpha1.Resources) (resReq *corev1.ResourceRequirements) {
	if &resources == nil {
		return nil
	}
	if len(resources.Limits) == 0 && len(resources.Requests) == 0 {
		return &corev1.ResourceRequirements{}
	}
	resReq = &corev1.ResourceRequirements{}
	// only build what is need to not conflict with DeepCopy later
	if len(resources.Limits) > 0 {
		resReq.Limits = corev1.ResourceList{}
	}
	if len(resources.Requests) > 0 {
		resReq.Requests = corev1.ResourceList{}
	}

	for _, limit := range resources.Limits {
		resReq.Limits[corev1.ResourceName(limit.Resource)] = resource.MustParse(string(limit.Value))
	}

	for _, request := range resources.Requests {
		resReq.Requests[corev1.ResourceName(request.Resource)] = resource.MustParse(string(request.Value))
	}

	return resReq
}
