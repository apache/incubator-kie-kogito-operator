package util

import (
	corev1 "k8s.io/api/core/v1"
)

// GetEnvVar returns the position of the EnvVar found by name
func GetEnvVar(envName string, env []corev1.EnvVar) int {
	for pos, v := range env {
		if v.Name == envName {
			return pos
		}
	}
	return -1
}

// EnvOverride replaces or appends the provided EnvVar to the collection
func EnvOverride(dst []corev1.EnvVar, src ...corev1.EnvVar) []corev1.EnvVar {
	for _, cre := range src {
		pos := GetEnvVar(cre.Name, dst)
		if pos != -1 {
			dst[pos] = cre
		} else {
			dst = append(dst, cre)
		}
	}
	return dst
}

//FromMapToEnvVar converts a map[string]string in the format KEY=VALUE into a EnvVar Kubernetes object
func FromMapToEnvVar(mapEnv map[string]string) []corev1.EnvVar {
	var envs []corev1.EnvVar

	for key, value := range mapEnv {
		envs = append(envs, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return envs
}

// FromEnvVarToMap will convert EnvVar resource to a Map definition
func FromEnvVarToMap(envs []corev1.EnvVar) map[string]string {
	envMap := map[string]string{}

	for _, env := range envs {
		envMap[env.Name] = env.Value
	}

	return envMap
}

// GetEnvVarFromContainer get the environment variable from the container
func GetEnvVarFromContainer(key string, container corev1.Container) string {
	if &container == nil {
		return ""
	}

	for _, env := range container.Env {
		if env.Name == key {
			return env.Value
		}
	}

	return ""
}

// SetEnvVar will update or add the environment variable into the given container
func SetEnvVar(key, value string, container *corev1.Container) {
	if container == nil {
		return
	}

	for i, env := range container.Env {
		if env.Name == key {
			container.Env[i].Value = value
			return
		}
	}

	container.Env = append(container.Env, corev1.EnvVar{Name: key, Value: value})
}

// EnvVarCheck checks whether the src and dst []EnvVar have the same values
func EnvVarCheck(dst, src []corev1.EnvVar) bool {
	for _, denv := range dst {
		if !envVarEqual(denv, src) {
			return false
		}
	}
	for _, senv := range src {
		if !envVarEqual(senv, dst) {
			return false
		}
	}
	return true
}

func envVarEqual(env corev1.EnvVar, envList []corev1.EnvVar) bool {
	match := false
	for _, e := range envList {
		if env.Name == e.Name {
			if env.Value == e.Value {
				match = true
				break
			}
		}
	}
	return match
}
