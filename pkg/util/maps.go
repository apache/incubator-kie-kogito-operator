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
	"bytes"
	"fmt"
	"strings"
)

const (
	keyPairSeparator = "="
	pairSeparator    = ","
)

// MapContainsMap returns true only if source contains expected map
func MapContainsMap(source, expected map[string]string) bool {
	if len(source) == 0 || len(expected) == 0 {
		return false
	}
	for k, v := range expected {
		if source[k] != v {
			return false
		}
	}
	return true
}

// FromMapToString converts a map into a string format such as key1=value1,key2=value2
func FromMapToString(labels map[string]string) string {
	b := new(bytes.Buffer)
	for k, v := range labels {
		fmt.Fprintf(b, "%s=%s%s", k, v, pairSeparator)
	}
	return strings.TrimSuffix(b.String(), pairSeparator)
}

// FromStringsKeyPairToMap converts a string array in the key/pair format (key=value) to a map. Unconvertable strings will be skipped.
func FromStringsKeyPairToMap(array []string) map[string]string {
	if array == nil || len(array) == 0 {
		return nil
	}
	keyPairMap := map[string]string{}
	for _, item := range array {
		keyPair := strings.SplitN(item, keyPairSeparator, 2)

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
