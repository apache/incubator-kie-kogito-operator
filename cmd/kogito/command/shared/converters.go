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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"strings"
)

const (
	// KeyPairSeparator separator for key value pair
	KeyPairSeparator = "="
	// SecretNameKeySeparator separator for secret key value pair
	SecretNameKeySeparator = "#"
)

// FromStringArrayToEnvs converts a string array in the format of key=value pairs to the required type for the Kubernetes EnvVar type
func FromStringArrayToEnvs(keyPairStrings []string, secretKeyPairStrings []string) (envVars []v1.EnvVar) {
	if len(keyPairStrings) > 0 {
		keyPairMap := FromStringsKeyPairToMap(keyPairStrings)
		for key, value := range keyPairMap {
			envVars = append(envVars, framework.CreateEnvVar(key, value))
		}
	}

	if len(secretKeyPairStrings) > 0 {
		keyPairMap := FromStringsKeyPairToMap(secretKeyPairStrings)
		for key, value := range keyPairMap {
			if strings.Contains(value, SecretNameKeySeparator) {
				secretNameKeyPair := strings.SplitN(value, SecretNameKeySeparator, 2)
				if len(secretNameKeyPair) == 2 {
					secretName := secretNameKeyPair[0]
					secretKey := secretNameKeyPair[1]
					secretEnvVar := framework.CreateSecretEnvVar(key, secretName, secretKey)
					envVars = append(envVars, secretEnvVar)
				}
			}
		}
	}
	return envVars
}

// FromStringArrayToResources ...
func FromStringArrayToResources(strings []string) v1.ResourceList {
	if strings == nil {
		return nil
	}
	res := v1.ResourceList{}
	mapStr := FromStringsKeyPairToMap(strings)
	for k, v := range mapStr {
		res[v1.ResourceName(k)] = resource.MustParse(v)
	}
	return res
}

// FromStringsKeyPairToMap converts a string array in the key/pair format (key=value) to a map. Unconvertable strings will be skipped.
func FromStringsKeyPairToMap(array []string) map[string]string {
	if len(array) == 0 {
		return nil
	}
	keyPairMap := map[string]string{}
	for _, item := range array {
		keyPair := strings.SplitN(item, KeyPairSeparator, 2)

		if len(keyPair[0]) == 0 {
			break
		}

		if len(keyPair) == 2 {
			keyPairMap[keyPair[0]] = keyPair[1]
		} else if len(keyPair) == 1 {
			keyPairMap[keyPair[0]] = ""
		}
	}
	return keyPairMap
}
