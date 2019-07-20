package index

import (
	"fmt"
	"github.com/paraterraform/para/utils"
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
	CacheDir            string
	KindToNameToPlugins map[string]map[string][]*Plugin
	Timestamp           time.Time
	Refresh             time.Duration
	Location            string
}

func DiscoverIndex(candidates []string, cacheDir string, refresh time.Duration) (*LoadingIndex, error) {
	var content []byte
	var timestamp time.Time
	var location string
	var err error

	indexCacheDir := filepath.Join(cacheDir, "index")

	for _, location = range candidates {
		if len(location) == 0 {
			continue
		}

		content, timestamp, err = utils.DownloadableFile{Url: location}.ReadAllWithCache(indexCacheDir, refresh)
		if err != nil {
			continue
		}
		break
	}

	if content == nil {
		return nil, fmt.Errorf(
			"failed to discover an index file at any of given locations: %s",
			strings.Join(candidates, ", "),
		)
	}

	index := &LoadingIndex{
		CacheDir:            cacheDir,
		KindToNameToPlugins: make(map[string]map[string][]*Plugin),
		Timestamp:           timestamp,
		Refresh:             refresh,
		Location:            location,
	}

	return index, index.loadPrimaryIndex(content)
}

func (i *LoadingIndex) loadPrimaryIndex(raw []byte) error {
	var parsed map[string]interface{}

	err := yml.Unmarshal(raw, &parsed)
	if err != nil {
		return err
	}

	for kind, kindSpec := range parsed {
		i.KindToNameToPlugins[kind] = make(map[string][]*Plugin)

		kindMap, kindSpecIsOk := kindSpec.(map[string]interface{})
		if !kindSpecIsOk {
			continue
		}

		for name, versionsSpec := range kindMap {
			plugins := i.parseVersions(kind, name, versionsSpec)
			i.KindToNameToPlugins[kind][name] = append(i.KindToNameToPlugins[kind][name], plugins...)
		}
	}

	return nil
}

func (i *LoadingIndex) LoadExtension(path string) error {
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

	plugins := i.parseVersions(kind, name, versionsSpec)

	if _, ok := i.KindToNameToPlugins[kind]; !ok {
		i.KindToNameToPlugins[kind] = make(map[string][]*Plugin)
	}

	i.KindToNameToPlugins[kind][name] = plugins

	return nil
}

func (i *LoadingIndex) parseVersions(kind, name string, versions interface{}) (result []*Plugin) {
	var versionMap map[string]interface{}

	extensionsCacheDir := filepath.Join(i.CacheDir, "index")

	versionMap, versionSpecIsOk := versions.(map[string]interface{})
	if !versionSpecIsOk {
		versionIndexUrl, versionIndexUrlOk := versions.(string)
		if !versionIndexUrlOk {
			return
		}
		versionSpecBytes, _, err := utils.DownloadableFile{Url: versionIndexUrl}.ReadAllWithCache(extensionsCacheDir, i.Refresh)
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

func (i *LoadingIndex) BuildRuntimeIndex() *RuntimeIndex {
	platformToPlugins := make(map[string]map[string]*Plugin)

	for _, nameToPlugins := range i.KindToNameToPlugins {
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
		cacheDir:                   i.CacheDir,
		openFiles:                  make(map[string]*os.File),
		alreadyOpened:              make(map[string]int),
	}
}
