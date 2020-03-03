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
	"fmt"
	"strings"
)

const keyPairSeparator = "="

// Contains checks if the s string are within the array
func Contains(s string, array []string) bool {
	if len(s) == 0 {
		return false
	}
	for _, item := range array {
		if s == item {
			return true
		}
	}
	return false
}

// FromStringsKeyPairToMap converts a string array in the key/pair format (key=value) to a map. Unconvertable strings will be skipped.
func FromStringsKeyPairToMap(array []string) map[string]string {
	if array == nil || len(array) == 0 {
		return nil
	}
	keyPairMap := map[string]string{}
	for _, item := range array {
		keyPair := strings.SplitN(item, keyPairSeparator, 2)
		if len(keyPair) == 0 {
			break
		}

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

// ParseStringsForKeyPair will parse the given string array for a valid key=pair format on each item.
// Returns an error if any item is not in the valid format.
func ParseStringsForKeyPair(array []string) error {
	if array == nil || len(array) == 0 {
		return nil
	}
	for _, item := range array {
		if !strings.Contains(item, keyPairSeparator) {
			return fmt.Errorf("Item %s does not contain the key/pair separator '%s'", item, keyPairSeparator)
		}
		if strings.HasPrefix(item, keyPairSeparator) {
			return fmt.Errorf("Item %s starts with key/pair separator '%s'", item, keyPairSeparator)
		}
	}
	return nil
}

// ArrayToSet converts an array of string to a set
func ArrayToSet(array []string) map[string]bool {
	set := make(map[string]bool, len(array))

	for _, e := range array {
		set[e] = true
	}

	return set
}

// ContainsAll checks if all the elements of the second are in the first array
func ContainsAll(array1 []string, array2 []string) bool {
	set1 := ArrayToSet(array1)

	for _, e := range array2 {
		if !set1[e] {
			return false
		}
	}

	return true
}
