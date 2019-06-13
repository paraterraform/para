package app

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"strings"
)

func pathExists(path string) bool {
	if len(path) == 0 {
		return false
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true // TODO verify it's writable?
	}
	return false
}

func simplifyPath(path string) string {
	if strings.HasPrefix(path, os.TempDir()) {
		return fmt.Sprintf("$TMPDIR/%s", path[len(os.TempDir()):])
	}
	if home, err := homedir.Dir(); err != nil {
		if strings.HasPrefix(path, home) {
			return fmt.Sprintf("~/%s", path[len(home):])
		}
	}
	return path
}
