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

package util

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
