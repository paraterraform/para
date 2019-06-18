package app

import (
	"encoding/json"
	"fmt"
	"github.com/paraterraform/para/utils"
	"path/filepath"
	"runtime"
	"time"
)

const (
	terragruntExec     = "terragrunt"
	terragruntReleases = "https://api.github.com/repos/gruntwork-io/terragrunt/releases"
	terragruntDownload = "https://github.com/gruntwork-io/terragrunt/releases/download"
)

func downloadTerragrunt(version, cacheDir string, refresh time.Duration) (string, error) {
	terragruntCacheDir := filepath.Join(cacheDir, terragruntExec)

	var versionToDownload string

	var urlRelease string
	if version != "" {
		versionToDownload = "v" + versionToDownload
	} else {
		urlRelease = utils.UrlJoin(terragruntReleases, "latest")
		releaseJsonBytes, _, err := utils.DownloadableFile{
			Url: urlRelease,
		}.ReadAllWithCache(filepath.Join(terragruntCacheDir, "versions"), refresh)
		if err != nil {
			return "", nil
		}

		var releaseJson map[string]interface{}
		if err := json.Unmarshal(releaseJsonBytes, &releaseJson); err != nil {
			panic(err)
		}
		versionRaw, okNameSet := releaseJson["name"]
		versionStr, okNameStr := versionRaw.(string)
		if !okNameSet || !okNameStr {
			return "", fmt.Errorf("erro cannot read release name at: %s", urlRelease)
		}
		versionToDownload = versionStr
	}

	pathToVersionDir := filepath.Join(
		terragruntCacheDir,
		versionToDownload,
		fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
	)
	pathToExecutable := filepath.Join(pathToVersionDir, terragruntExec)
	if utils.PathExists(pathToExecutable) {
		// already downloaded & cached
		// given that checksums published for archives we will check them when fetching binaries and before unpacking
		return pathToVersionDir, nil
	}
	// windows binary has .exe suffix but there is no FUSE on windows so there is no para on windows ¯\_(ツ)_/¯
	expectedFileName := terragruntExec + "_" + runtime.GOOS + "_" + runtime.GOARCH
	urlVersionPrefix := utils.UrlJoin(terragruntDownload, versionToDownload)
	urlVersionChecksums := utils.UrlJoin(urlVersionPrefix, "SHA256SUMS")
	urlVersionBinary := utils.UrlJoin(urlVersionPrefix, expectedFileName)

	sha256 := findChecksumForFile(
		urlVersionChecksums, expectedFileName,
		filepath.Join(terragruntCacheDir, "checksums"), refresh,
	)

	err := utils.DownloadableFile{
		Url:    urlVersionBinary,
		Digest: "sha256:" + sha256,
	}.SaveTo(pathToExecutable)
	if err != nil {
		return "", err
	}

	return pathToVersionDir, nil
}
