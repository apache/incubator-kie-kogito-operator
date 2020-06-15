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

package framework

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

// EnvVarToMap converts an array of Env to a map
func EnvVarToMap(env []corev1.EnvVar) map[string]string {
	envMap := make(map[string]string, len(env))

	for _, e := range env {
		envMap[e.Name] = e.Value
	}

	return envMap
}

// EnvVarArrayEquals checks if the elements of two arrays of EnvVar are identical
func EnvVarArrayEquals(array1 []corev1.EnvVar, array2 []corev1.EnvVar) bool {
	if len(array1) != len(array2) {
		return false
	}

	map1 := EnvVarToMap(array1)
	map2 := EnvVarToMap(array2)
	return reflect.DeepEqual(map1, map2)
}

// MapToEnvVar converts a map to an array of EnvVar
func MapToEnvVar(env map[string]string) (envVars []corev1.EnvVar) {
	for key, value := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return envVars
}

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

// GetEnvVarFromContainer gets the environment variable value from the container
func GetEnvVarFromContainer(key string, container *corev1.Container) string {
	if container == nil {
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

// SetEnvVarFromSecret will set the Environment Variable from a Secret
func SetEnvVarFromSecret(key, secretKey string, secret *corev1.Secret, container *corev1.Container) {
	if container == nil || secret == nil {
		return
	}
	valueFrom := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
			Key:                  secretKey,
		},
	}
	SetEnvVar(key, "", container)
	for i, env := range container.Env {
		if env.Name == key {
			container.Env[i].ValueFrom = valueFrom
			return
		}
	}
	container.Env = append(container.Env, corev1.EnvVar{Name: key, ValueFrom: valueFrom})
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
