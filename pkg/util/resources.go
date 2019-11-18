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
)

// EnvToMap converts an array of Env to a map
func EnvToMap(env []v1alpha1.Env) map[string]string {
	envMap := map[string]string{}

	for _, e := range env {
		envMap[e.Name] = e.Value
	}

	return envMap
}

// AppendOrReplaceEnv add a new env or replace the exist one
func AppendOrReplaceEnv(env v1alpha1.Env, envs []v1alpha1.Env) []v1alpha1.Env {
	if envs == nil {
		envs = []v1alpha1.Env{}
	}

	for pos, e := range envs {
		if e.Name == env.Name {
			envs[pos] = env
			return envs
		}
	}

	envs = append(envs, env)
	return envs
}

// GetEnvValue gets the value from the env by it's key
func GetEnvValue(key string, envs []v1alpha1.Env) string {
	for _, e := range envs {
		if e.Name == key {
			return e.Value
		}
	}
	return ""
}
