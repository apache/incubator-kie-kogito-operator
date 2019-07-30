package shared

// Contains check if a value is present into the string array
func Contains(strings []string, value string) bool {
	for _, str := range strings {
		if str == value {
			return true
		}
	}
	return false
}
