package index

import (
	"fmt"
	"github.com/paraterraform/para/utils"
	"io"
	"os"
	"path/filepath"
)

type RuntimeIndex struct {
	platformToFilenameToPlugin map[string]map[string]*Plugin
	cacheDir                   string
	openFiles                  map[string]*os.File

	alreadyOpened map[string]bool
}

func (i *RuntimeIndex) ListPluginsForPlatform(platform string) []string {
	var keys []string
	for k := range i.platformToFilenameToPlugin[platform] {
		keys = append(keys, k)
	}
	return keys
}

func (i *RuntimeIndex) ListPlatforms() []string {
	var keys []string
	for k := range i.platformToFilenameToPlugin {
		keys = append(keys, k)
	}
	return keys
}

func (i *RuntimeIndex) LookupPlugin(platform, filename string) *Plugin {
	platformPlugins, knownPlatform := i.platformToFilenameToPlugin[platform]
	if !knownPlatform {
		return nil
	}
	return platformPlugins[filename]
}

func (i RuntimeIndex) getPluginFilePath(plugin *Plugin) string {
	return filepath.Join(i.cacheDir, "plugins", plugin.Kind, plugin.Name, plugin.Version, plugin.Platform)
}

func (i *RuntimeIndex) OpenPlugin(plugin *Plugin) error {
	path := i.getPluginFilePath(plugin)

	cached := true
	cachedStateStr := "cached"

	if verifyPlugin(path, plugin.Size, plugin.Digest) != nil {
		cached = false
		cachedStateStr = "downloading"
	}

	// Trying to blend in with Terraform output nicely
	// Would always print 1 extra newline at the end and then rewrite if we believe we're rewriting our content
	// Just so that there is a nice indentation with other sections
	lineControl := ""
	if len(i.alreadyOpened) > 0 {
		lineControl = "\x1b[1A"
	}

	if v, ok := i.alreadyOpened[path]; !ok || !v {
		fmt.Printf(
			"%s- Para provides 3rd-party Terraform %s plugin '%s' version '%s' for '%s' (%s)\n\n",
			lineControl, plugin.Kind, plugin.Name, plugin.Version, plugin.Platform, cachedStateStr,
		)
		i.alreadyOpened[path] = true
	}

	if !cached {
		err := downloadPlugin(plugin.Url, path)
		if err != nil {
			fmt.Printf("   * Error reading '%s': %s\n", plugin.Url, err)
			return err
		}
		err = verifyPlugin(path, plugin.Size, plugin.Digest)
		if err != nil {
			_ = os.Remove(path)
			fmt.Printf("   * Error reading '%s': %s\n", plugin.Url, err)
			return err
		}
	}
	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	i.openFiles[path] = reader

	return nil
}

func (i *RuntimeIndex) GetReaderAt(plugin *Plugin) (io.ReaderAt, error) {
	path := i.getPluginFilePath(plugin)
	reader, ok := i.openFiles[path]
	if !ok {
		return nil, fmt.Errorf("plugin file '%s' platform has not been opened yet", path)
	}
	return reader, nil
}

func (i *RuntimeIndex) ClosePlugin(plugin *Plugin) error {
	path := i.getPluginFilePath(plugin)
	reader, ok := i.openFiles[path]
	if ok {
		err := reader.Close()
		if err != nil {
			return err
		}
	}
	delete(i.openFiles, path)
	return nil
}

func downloadPlugin(url string, saveTo string) error {
	// Get the data
	pluginData, err := utils.UrlOpen(url)
	if err != nil {
		return err
	}
	defer func() { _ = pluginData.Close() }()

	// Create the file
	err = os.MkdirAll(filepath.Dir(saveTo), 0755)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(saveTo, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	defer func() { _ = os.Chmod(saveTo, 0666) }()

	// Write the body to file
	_, err = io.Copy(out, pluginData)
	return err
}

func verifyPlugin(path string, size uint64, digest string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if uint64(info.Size()) != size {
		return fmt.Errorf("actual size of %d does not match expected value of %d", info.Size(), size)
	}

	if len(digest) > 0 {
		return utils.DigestVerify(path, digest)
	}

	return nil
}
