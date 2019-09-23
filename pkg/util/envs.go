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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"os"
	"strconv"
)

// GetBoolEnv gets a env variable as a boolean
func GetBoolEnv(key string) bool {
	val := GetEnv(key, "false")
	ret, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return ret
}

// GetEnv gets a env variable
func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// GetHomeDir gets the user home directory
func GetHomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// EnvToMap converts an array of Env to a map
func EnvToMap(env []v1alpha1.Env) map[string]string {
	envMap := map[string]string{}

	for _, e := range env {
		envMap[e.Name] = e.Value
	}

	return envMap
}
