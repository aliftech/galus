package function

import (
	"os"
	"path/filepath"
	"strings"
)

func CheckGoFiles(rootDir string) (bool, error) {
	hasGoFiles := false
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".go") {
			hasGoFiles = true
		}
		return nil
	})
	return hasGoFiles, err
}
