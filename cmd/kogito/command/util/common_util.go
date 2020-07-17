// Copyright 2020 Red Hat, Inc. and/or its affiliates
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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"strings"
)

// CheckKeyPair will parse the given string array for a valid key=pair format on each item.
// Returns an error if any item is not in the valid format.
func CheckKeyPair(array []string) error {
	if len(array) == 0 {
		return nil
	}
	for _, item := range array {
		if !strings.Contains(item, shared.KeyPairSeparator) {
			return fmt.Errorf("Item %s does not contain the key/pair separator '%s'", item, shared.KeyPairSeparator)
		}
		if strings.HasPrefix(item, shared.KeyPairSeparator) {
			return fmt.Errorf("Item %s starts with key/pair separator '%s'", item, shared.KeyPairSeparator)
		}
	}
	return nil
}

// CheckSecretKeyPair will parse the given string array for a valid key=pair#value format on each item.
// Returns an error if any item is not in the valid format.
func CheckSecretKeyPair(secretKeyValuePair []string) error {
	if err := CheckKeyPair(secretKeyValuePair); err != nil {
		return nil
	}

	keyPairMap := shared.FromStringsKeyPairToMap(secretKeyValuePair)
	for _, value := range keyPairMap {
		if !strings.Contains(value, shared.SecretNameKeySeparator) {
			return fmt.Errorf("item %s does not contain the secret key/pair separator '%s'", value, shared.SecretNameKeySeparator)
		}
	}
	return nil
}

// CheckImageTag checks the given image tag
func CheckImageTag(image string) error {
	if len(image) > 0 && !framework.DockerTagRegxCompiled.MatchString(image) {
		return fmt.Errorf("invalid name for image tag. Valid format is domain/namespace/image-name:tag. Received %s", image)
	}
	return nil
}
