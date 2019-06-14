package utils

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
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
	if strings.HasPrefix(path, os.TempDir()) {
		return fmt.Sprintf("$TMPDIR/%s", path[len(os.TempDir()):])
	}

	home, err := homedir.Dir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return fmt.Sprintf("~%s", path[len(home):])
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
