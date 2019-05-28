package shared

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

// EnvOverride replaces or appends the provided EnvVar to the collection
func EnvOverride(dst, src []corev1.EnvVar) []corev1.EnvVar {
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
