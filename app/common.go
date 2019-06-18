package app

import (
	"github.com/paraterraform/para/utils"
	"path/filepath"
	"strings"
	"time"
)

func findChecksumForFile(prefix, url, file, cache string, refresh time.Duration) string {
	checksums, _, err := utils.DownloadableFile{
		Url: url,
	}.ReadAllWithCache(filepath.Join(cache, "checksums"), refresh)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		lineHash := fields[0]
		lineName := fields[1]
		if lineName == file {
			return prefix + lineHash
		}

	}
	return ""
}
