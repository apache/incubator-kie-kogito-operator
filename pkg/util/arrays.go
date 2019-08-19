package util

import (
	"fmt"
	"strings"
)

const keyPairSeparator = "="

// FromStringsKeyPairToMap converts a string array in the key/pair format (key=value) to a map. Unconvertable strings will be skipped.
func FromStringsKeyPairToMap(array []string) map[string]string {
	if array == nil || len(array) == 0 {
		return nil
	}
	kp := map[string]string{}
	for _, item := range array {
		spplited := strings.SplitN(item, keyPairSeparator, 2)
		if len(spplited) == 0 {
			break
		}

		if len(spplited) == 2 {
			kp[spplited[0]] = spplited[1]
		} else if len(spplited) == 1 {
			kp[spplited[0]] = ""
		}
	}
	return kp
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
	}
	return nil
}
