package index

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/paraterraform/para/app/transport"
	yml "gopkg.in/ashald/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const fieldUrl = "url"
const fieldSize = "size"
const fieldDigest = "digest"

type LoadingIndex struct {
	kindToNameToPlugins map[string]map[string][]*Plugin
}

func DiscoverIndex(candidates []string) (*LoadingIndex, string, error) {
	var content []byte
	var location string
	var err error

	for _, location = range candidates {
		if len(location) == 0 {
			continue
		}

		expandedPath, err := homedir.Expand(location)

		reader, err := transport.OpenUrl(expandedPath)
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

	index, err := newLoadingIndex(content)
	return index, location, err
}

func newLoadingIndex(raw []byte) (*LoadingIndex, error) {
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
			versionMap, versionSpecIsOk := versionsSpec.(map[string]interface{})
			if !versionSpecIsOk {
				continue
			}

			plugins := parseVersions(kind, name, versionMap)
			kindToNameToPlugins[kind][name] = append(kindToNameToPlugins[kind][name], plugins...)
		}
	}

	// TODO: verify collisions in a case-insensitive way
	return &LoadingIndex{
		kindToNameToPlugins: kindToNameToPlugins,
	}, nil
}

func parseVersions(kind, name string, versionMap map[string]interface{}) (result []*Plugin) {
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

	var versionsMap map[string]interface{}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yml.Unmarshal(content, &versionsMap)
	if err != nil {
		return err
	}

	plugins := parseVersions(kind, name, versionsMap)

	if _, ok := i.kindToNameToPlugins[kind]; !ok {
		i.kindToNameToPlugins[kind] = make(map[string][]*Plugin)
	}

	i.kindToNameToPlugins[kind][name] = plugins

	return nil
}

func (i *LoadingIndex) BuildPlatformIndex(cacheDir string) *RuntimeIndex {
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
		cacheDir:                   cacheDir,
		openFiles:                  make(map[string]*os.File),
		alreadyOpened:              make(map[string]bool),
	}
}
