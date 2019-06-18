package app

import (
	"fmt"
	"github.com/paraterraform/para/utils"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

const (
	terraformExec     = "terraform"
	terraformReleases = "https://releases.hashicorp.com/terraform/"
)

var terraformVersionRe = *regexp.MustCompile(`href="/terraform/([\d\\.]+?)/"`)

func downloadTerraform(version, cacheDir string, refresh time.Duration) (string, error) {
	terraformCacheDir := filepath.Join(cacheDir, terraformExec)

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

	pathToVersionDir := filepath.Join(
		terraformCacheDir,
		versionToDownload,
		fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
	)
	pathToExecutable := filepath.Join(pathToVersionDir, terraformExec)
	if utils.PathExists(pathToExecutable) {
		// already downloaded & cached
		// given that checksums published for archives we will check them when fetching binaries and before unpacking
		return pathToVersionDir, nil
	}

	// windows binary has .exe suffix but there is no FUSE on windows so there is no para on windows ¯\_(ツ)_/¯
	expectedFileName := terragruntExec + "_" + runtime.GOOS + "_" + runtime.GOARCH
	urlVersionPrefix := utils.UrlJoin(terraformReleases, versionToDownload)
	urlVersionChecksums := utils.UrlJoin(urlVersionPrefix, terraformExec+versionToDownload+"_SHA256SUMS")
	urlVersionBinary := utils.UrlJoin(urlVersionPrefix, expectedFileName)

	sha256 := findChecksumForFile(
		urlVersionChecksums, expectedFileName,
		filepath.Join(terraformCacheDir, "checksums"), refresh,
	)

	err := utils.DownloadableFile{
		Url:            urlVersionBinary,
		Digest:         "sha256:" + sha256,
		ExtractPattern: "terraform*",
	}.SaveTo(pathToExecutable)
	if err != nil {
		return "", err
	}

	return pathToVersionDir, nil
}
