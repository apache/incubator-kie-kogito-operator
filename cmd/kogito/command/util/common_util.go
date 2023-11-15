/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package util

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/core/framework"
	"os"
	"strings"
)

const (
	// KeyPairSeparator separator for key value pair
	KeyPairSeparator = "="
	// SecretNameKeySeparator separator for secret key value pair
	SecretNameKeySeparator = "#"
)

// CheckKeyPair will parse the given string array for a valid key=pair format on each item.
// Returns an error if any item is not in the valid format.
func CheckKeyPair(array []string) error {
	if len(array) == 0 {
		return nil
	}
	for _, item := range array {
		if !strings.Contains(item, KeyPairSeparator) {
			return fmt.Errorf("Item %s does not contain the key/pair separator '%s'", item, KeyPairSeparator)
		}
		if strings.HasPrefix(item, KeyPairSeparator) {
			return fmt.Errorf("Item %s starts with key/pair separator '%s'", item, KeyPairSeparator)
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

	keyPairMap := FromStringsKeyPairToMap(secretKeyValuePair)
	for _, value := range keyPairMap {
		if !strings.Contains(value, SecretNameKeySeparator) {
			return fmt.Errorf("item %s does not contain the secret key/pair separator '%s'", value, SecretNameKeySeparator)
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

// CheckFileExists checks if the given path is valid
func CheckFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}
	return true, nil
}
