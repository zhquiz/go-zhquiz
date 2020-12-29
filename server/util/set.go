package util

// MakeSet converts array to set
func MakeSet(arr []string) map[string]bool {
	out := map[string]bool{}

	for _, r := range arr {
		out[r] = true
	}

	return out
}
