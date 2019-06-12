package index

import (
	"fmt"
	"github.com/paraterraform/para/app/crypto"
	"github.com/paraterraform/para/app/xio"
	yml "gopkg.in/ashald/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const fieldUrl = "url"
const fieldSize = "size"
const fieldDigest = "digest"

type LoadingIndex struct {
	cacheDir            string
	kindToNameToPlugins map[string]map[string][]*Plugin
}

func DiscoverIndex(candidates []string, cacheDir string, refresh time.Duration) (*LoadingIndex, string, error) {
	var content []byte
	var location string
	var err error

	for _, location = range candidates {
		if len(location) == 0 {
			continue
		}

		content, err = openIndex(location, cacheDir, refresh)
		if err != nil {
			continue
		}
		break
	}

	if content == nil {
		return nil, location, fmt.Errorf(
			"failed to discover an index file at any of given locations: %s",
			strings.Join(candidates, ", "),
		)
	}

	index, err := newLoadingIndex(content, cacheDir, refresh)
	return index, location, err
}

func openIndex(location string, cacheDir string, refresh time.Duration) ([]byte, error) {
	if xio.IsRemote(location) {
		indexCacheDir := filepath.Join(cacheDir, "index")
		indexCachePath := filepath.Join(indexCacheDir, crypto.DefaultStringHash(location))

		cacheData, errCacheData := ioutil.ReadFile(indexCachePath)
		cacheMeta, errCacheMeta := os.Stat(indexCachePath)

		maxAge := time.Duration(refresh) * time.Minute
		if errCacheMeta != nil || cacheMeta.ModTime().Before(time.Now().Add(-maxAge)) || errCacheData != nil {
			freshData, errFreshData := xio.UrlReadAll(location)
			if errFreshData != nil {
				if errCacheData != nil {
					return nil, errFreshData // we're not sure our cache is valid and failed to fetch so just fail
				}
				// failed to fetch but there is _some_ kind of a cache - just return it
				return cacheData, nil
			}
			// we fetched fresh data - let's try to cache it but don't sweat if fail
			_ = os.MkdirAll(indexCacheDir, 0755)
			_ = ioutil.WriteFile(indexCachePath, freshData, 0644)
			return freshData, nil
		} else {
			return cacheData, nil
		}
	} else {
		return xio.UrlReadAll(location)
	}
}

func newLoadingIndex(raw []byte, cacheDir string, refresh time.Duration) (*LoadingIndex, error) {
	var parsed map[string]interface{}

	err := yml.Unmarshal(raw, &parsed)
	if err != nil {
		return nil, err
	}

	kindToNameToPlugins := make(map[string]map[string][]*Plugin)

	for kind, kindSpec := range parsed {
		kindToNameToPlugins[kind] = make(map[string][]*Plugin)

		kindMap, kindSpecIsOk := kindSpec.(map[string]interface{})
		if !kindSpecIsOk {
			continue
		}

		for name, versionsSpec := range kindMap {
			plugins := parseVersions(kind, name, versionsSpec, cacheDir, refresh)
			kindToNameToPlugins[kind][name] = append(kindToNameToPlugins[kind][name], plugins...)
		}
	}

	// TODO: verify collisions in a case-insensitive way
	return &LoadingIndex{
		cacheDir:            cacheDir,
		kindToNameToPlugins: kindToNameToPlugins,
	}, nil
}

func parseVersions(kind, name string, versions interface{}, cacheDir string, refresh time.Duration) (result []*Plugin) {
	var versionMap map[string]interface{}

	versionMap, versionSpecIsOk := versions.(map[string]interface{})
	if !versionSpecIsOk {
		versionIndexUrl, versionIndexUrlOk := versions.(string)
		if !versionIndexUrlOk {
			return
		}
		versionSpecBytes, err := openIndex(versionIndexUrl, cacheDir, refresh)
		if err != nil {
			return
		}
		err = yml.Unmarshal(versionSpecBytes, &versionMap)
		if err != nil {
			return
		}
	}

	for version, platformsSpec := range versionMap {
		platformsMap, platformsSpecIsOk := platformsSpec.(map[string]interface{})
		if !platformsSpecIsOk {
			continue
		}

		for platform, platformSpec := range platformsMap {
			specMap, platformSpecIsOk := platformSpec.(map[string]interface{})
			if !platformSpecIsOk {
				continue
			}

			urlRaw, okUrl := specMap[fieldUrl]
			urlStr, okUrlValue := urlRaw.(string)
			if !okUrl || !okUrlValue {
				continue // TODO trace
			}

			sizeRaw, okSize := specMap[fieldSize]
			size, err := strconv.ParseUint(fmt.Sprintf("%v", sizeRaw), 10, 64)
			if !okSize || err != nil {
				continue // TODO trace
			}

			digestRaw, okDigest := specMap[fieldDigest]
			digestStr, okDigestValue := digestRaw.(string)
			if !okDigest || !okDigestValue {
				continue // TODO trace
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

			result = append(result, &p)
		}
	}
	return
}

func (i *LoadingIndex) LoadExtension(path string, refresh time.Duration) error {
	filename := filepath.Base(path)
	if strings.ToLower(filename) != filename {
		return fmt.Errorf("extension file must be in lowercase: '%s'", filename)
	}
	tokens := strings.SplitN(filename, ".", 3)
	if len(tokens) != 3 || tokens[2] != "yaml" {
		return fmt.Errorf(
			"extension file name '%s' does not match expected pattern of <kind>.<name>.yaml",
			filename,
		)
	}
	kind := tokens[0]
	name := tokens[1]

	var versionsSpec interface{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yml.Unmarshal(content, &versionsSpec)
	if err != nil {
		return err
	}

	plugins := parseVersions(kind, name, versionsSpec, i.cacheDir, refresh)

	if _, ok := i.kindToNameToPlugins[kind]; !ok {
		i.kindToNameToPlugins[kind] = make(map[string][]*Plugin)
	}

	i.kindToNameToPlugins[kind][name] = plugins

	return nil
}

func (i *LoadingIndex) BuildPlatformIndex() *RuntimeIndex {
	platformToPlugins := make(map[string]map[string]*Plugin)

	for _, nameToPlugins := range i.kindToNameToPlugins {
		for _, plugins := range nameToPlugins {
			for _, p := range plugins {
				if _, ok := platformToPlugins[p.Platform]; !ok {
					platformToPlugins[p.Platform] = make(map[string]*Plugin)
				}
				platformToPlugins[p.Platform][p.Filename()] = p
			}
		}
	}

	return &RuntimeIndex{
		platformToFilenameToPlugin: platformToPlugins,
		cacheDir:                   i.cacheDir,
		openFiles:                  make(map[string]*os.File),
		alreadyOpened:              make(map[string]bool),
	}
}
