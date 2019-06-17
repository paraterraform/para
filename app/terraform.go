package app

import (
	"fmt"
	"github.com/paraterraform/para/utils"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const terraformReleases = "https://releases.hashicorp.com/terraform/"

var terraformVersionRe = *regexp.MustCompile(`href="/terraform/([\d\\.]+?)/"`)

func downloadTerraform(version, cacheDir string, refresh time.Duration) (string, error) {
	terraformCacheDir := filepath.Join(cacheDir, execTerraform)

	var versionToDownload string

	if version != "" {
		versionToDownload = version
	} else {
		versionsHtmlBytes, _, err := utils.DownloadableFile{Url: terraformReleases}.ReadAllWithCache(
			filepath.Join(terraformCacheDir, "versions"), refresh,
		)
		if err != nil {
			return "", err
		}
		var knownVersions []string
		for _, match := range terraformVersionRe.FindAllStringSubmatch(string(versionsHtmlBytes), -1) {
			knownVersions = append(knownVersions, match[1])
		}
		versionToDownload = knownVersions[0] // keep it sample for now, assume list is sorted
	}

	pathToTerraform := filepath.Join(
		terraformCacheDir,
		versionToDownload,
		fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
		execTerraform,
	)
	if utils.PathExists(pathToTerraform) {
		// already downloaded & cached
		// given that checksums published for archives we will check them when fetching binaries and before unpacking
		return pathToTerraform, nil
	}

	urlVersion := utils.UrlJoin(terraformReleases, versionToDownload)
	urlVersionChecksums := utils.UrlJoin(urlVersion, fmt.Sprintf("terraform_%s_SHA256SUMS", versionToDownload))

	checksumsBytes, _, err := utils.DownloadableFile{Url:urlVersionChecksums}.ReadAllWithCache(
		filepath.Join(terraformCacheDir, "checksums"), time.Hour*24*365*10,
	)
	if err != nil {
		return "", err
	}

	var targetChecksum string
	var targetFilename string

	for _, line := range strings.Split(string(checksumsBytes), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if strings.HasPrefix(
			fields[1],
			fmt.Sprintf("terraform_%s_%s_%s", versionToDownload, runtime.GOOS, runtime.GOARCH),
		) {
			targetChecksum = fields[0]
			targetFilename = fields[1]
		}
	}
	urlToDownload := utils.UrlJoin(urlVersion, targetFilename)
	err = utils.DownloadableFile{
		Url:urlToDownload,
		Digest:"sha256:"+targetChecksum,
		ExtractPattern:"terraform*",
	}.SaveTo(pathToTerraform)
	if err != nil {
		return "", err
	}

	return pathToTerraform, nil
}
