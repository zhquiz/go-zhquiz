package util

// MakeSet converts array to set
func MakeSet(arr []string) map[string]bool {
	var out map[string]bool

	for _, r := range arr {
		out[r] = true
	}

	return out
}
