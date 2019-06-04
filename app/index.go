package app

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	yml "gopkg.in/ashald/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const FieldUrl = "url"
const FieldSize = "size"
const FieldDigest = "digest"

type Index struct {
	plugins   map[string]map[string]*Plugin
	cacheDir  string
	openFiles map[string]*os.File

	alreadyOpened map[string]bool
}

func DiscoverIndex(candidates []string, cacheDir string) (*Index, string, error) {
	var content []byte
	var location string
	var err error

	for _, location = range candidates {
		if len(location) == 0 {
			continue
		}

		expandedPath, err := homedir.Expand(location)

		reader, err := openUrl(expandedPath)
		if err != nil {
			continue
		}

		content, err = ioutil.ReadAll(reader)
		_ = reader.Close() // What can possibly go wrong?
		if err == nil {
			break
		}
	}

	if content == nil {
		return nil, location, fmt.Errorf(
			"failed to discover an index file at any of given locations: %s",
			strings.Join(candidates, ", "),
		)
	}

	index, err := newIndex(content, cacheDir)
	return index, location, err
}

func newIndex(raw []byte, cacheDir string) (*Index, error) {
	var parsed map[string]interface{}

	err := yml.Unmarshal(raw, &parsed)
	if err != nil {
		return nil, err
	}

	plugins := make(map[string]map[string]*Plugin)

	for kind, kindsRaw := range parsed {
		kindMap := kindsRaw.(map[string]interface{})
		for name, versionsRaw := range kindMap {
			versionMap := versionsRaw.(map[string]interface{})
			for version, platformsRaw := range versionMap {
				platformsMap := platformsRaw.(map[string]interface{})
				for platform, specRaw := range platformsMap {
					specMap := specRaw.(map[string]interface{})

					urlRaw, okUrl := specMap[FieldUrl]
					urlStr, okUrlValue := urlRaw.(string)
					if !okUrl || !okUrlValue {
						continue // TODO trace
					}

					sizeRaw, okSize := specMap[FieldSize]
					size, err := strconv.ParseUint(fmt.Sprintf("%v", sizeRaw), 10, 64)
					if !okSize || err != nil {
						continue // TODO trace
					}

					var digestStr string
					digestRaw, okDigest := specMap[FieldDigest]
					if okDigest {
						digestStr, _ = digestRaw.(string)
					}

					p := Plugin{
						Kind:     kind,
						Name:     name,
						Platform: platform,
						Version:  version,
						Size:     size,
						Digest:   digestStr,
						Url:      urlStr,
					}

					if _, ok := plugins[platform]; !ok {
						plugins[platform] = make(map[string]*Plugin)
					}

					plugins[platform][p.Filename()] = &p
				}
			}
		}
	}

	// TODO: verify collisions in a case-insensitive way
	return &Index{
		plugins:       plugins,
		cacheDir:      cacheDir,
		openFiles:     make(map[string]*os.File),
		alreadyOpened: make(map[string]bool),
	}, nil
}

func (i *Index) ListPluginsForPlatform(platform string) []string {
	var keys []string
	for k := range i.plugins[platform] {
		keys = append(keys, k)
	}
	return keys
}

func (i *Index) ListPlatforms() []string {
	var keys []string
	for k := range i.plugins {
		keys = append(keys, k)
	}
	return keys
}

func (i *Index) LookupPlugin(platform, filename string) *Plugin {
	platformPlugins, knownPlatform := i.plugins[platform]
	if !knownPlatform {
		return nil
	}
	return platformPlugins[filename]
}

func (i Index) getPluginFilePath(plugin *Plugin) string {
	return filepath.Join(i.cacheDir, plugin.Kind, plugin.Name, plugin.Version, plugin.Platform)
}

func (i *Index) OpenPlugin(plugin *Plugin) error {
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

func (i *Index) GetReaderAt(plugin *Plugin) (io.ReaderAt, error) {
	path := i.getPluginFilePath(plugin)
	reader, ok := i.openFiles[path]
	if !ok {
		return nil, fmt.Errorf("plugin file '%s' platform has not been opened yet", path)
	}
	return reader, nil
}

func (i *Index) ClosePlugin(plugin *Plugin) error {
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
	pluginData, err := openUrl(url)
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
		return checkDigest(path, digest)
	}

	return nil
}
