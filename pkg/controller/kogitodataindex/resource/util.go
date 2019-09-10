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

// removeManagyedEnvVars will remove any managed environment variable from KogitoDataIndex.Spec.Env. Users don't supposed to add managed env vars to the dataindex directly.
func removeManagedEnvVars(instance *v1alpha1.KogitoDataIndex) {
	for _, key := range managedEnvKeys {
		delete(instance.Spec.Env, key)
	}
}

// extractManagedEnvVars removes managed env vars from a given container and return them
func extractManagedEnvVars(container *corev1.Container) []corev1.EnvVar {
	managedEnvs := []corev1.EnvVar{}
	nonManagedEnvs := []corev1.EnvVar{}
	isManaged := false

	for _, env := range container.Env {
		for _, managed := range managedEnvKeys {
			if managed == env.Name {
				managedEnvs = append(managedEnvs, env)
				isManaged = true
				break
			}
		}
		if !isManaged {
			nonManagedEnvs = append(nonManagedEnvs, env)
		} else {
			isManaged = false
		}
	}

	container.Env = nonManagedEnvs

	return managedEnvs
}
