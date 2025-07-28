package function

import "strings"

func IsExcludedDir(path string, excludeDirs []string) bool {
	for _, dir := range excludeDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}
	return false
}
