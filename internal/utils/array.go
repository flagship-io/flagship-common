package utils

// IsInStringArray look for a string in a string array
func IsInStringArray(s string, a []string) bool {
	for _, i := range a {
		if i == s {
			return true
		}
	}
	return false
}
