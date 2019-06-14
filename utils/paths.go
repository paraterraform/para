package utils

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"strings"
)

func PathExists(path string) bool {
	if len(path) == 0 {
		return false
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true // TODO verify it's writable?
	}
	return false
}

func PathSimplify(path string) string {
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		if strings.HasPrefix(path, tmpDir) {
			return filepath.Join("$TMPDIR", path[len(os.TempDir()):])
		}
	}

	home, err := homedir.Dir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return filepath.Join("~", path[len(home):])
	}
	return path
}

func PathExpand(path string) string {
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return path
	}
	return expandedPath
}
