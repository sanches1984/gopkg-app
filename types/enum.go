package types

import "strings"

// ExtendEnumWithLowerKeys ...
func ExtendEnumWithLowerKeys(data map[string]int32) map[string]int32 {
	result := make(map[string]int32)
	for k, v := range data {
		result[k] = v
		lc := strings.ToLower(k)
		if k != lc {
			result[lc] = v
		}
	}
	return result
}
