package function

import "strings"

func IsIncludedExt(name string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(name), "."+strings.ToLower(ext)) {
			return true
		}
	}
	return false
}
