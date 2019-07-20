package utils

import (
	"fmt"
	"github.com/gobwas/glob"
	"github.com/mholt/archiver"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadableFile struct {
	Url            string
	Digest         string
	ExtractPattern string // must be set to extract archives
}

func (d DownloadableFile) Open() (io.ReadCloser, error) {
	var err error
	var reader io.ReadCloser

	if UrlIsRemote(d.Url) {
		resp, err := http.Get(d.Url)
		if err != nil {
			return nil, err
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf(
				"non-200 response while fetching '%s': %s",
				d.Url, http.StatusText(resp.StatusCode),
			)
		}
		reader = resp.Body
	} else {
		path := d.Url

		if strings.HasPrefix(d.Url, schemaFile) {
			path = d.Url[len(schemaFile):]
		}

		expandedPath, err := homedir.Expand(path)
		if err != nil {
			return nil, err
		}

		reader, err = os.Open(expandedPath)
		if err != nil {
			return nil, err
		}
	}

	// Create temp file to fetch data into
	rawData, err := ioutil.TempFile("", fmt.Sprintf("para.raw.*.%s", filepath.Base(d.Url)))
	if err != nil {
		return nil, err
	}

	// Download
	_, err = io.Copy(rawData, reader)
	_ = reader.Close() // we have to close it regardless of the error status

	if err != nil {
		_ = rawData.Close()
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	// Rewind to the beginning of the file
	_, err = rawData.Seek(0, 0)
	if err != nil {
		_ = rawData.Close()
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	// Verify digest if provided
	if d.Digest != "" {
		err = DigestVerify(rawData.Name(), d.Digest)
		if err != nil {
			_ = os.Remove(rawData.Name())

			return nil, err
		}
	}

	// Extract if supported and requested
	_, err = archiver.ByExtension(rawData.Name())
	if err != nil || d.ExtractPattern == "" {
		// not compressed or not supported or extraction not requested - treat it as raw data
		return VolatileTempFile{file: rawData}, nil
	}
	err = rawData.Close()
	if err != nil {
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	// file is archived and compression algorithm is supported
	uncompressedData, err := ioutil.TempFile("", fmt.Sprintf("%s.extracted.*", filepath.Base(rawData.Name())))
	if err != nil {
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	err = archiver.Walk(rawData.Name(), func(f archiver.File) error {
		if f.IsDir() {
			return nil
		}
		if glob.MustCompile(d.ExtractPattern).Match(f.Name()) {
			_, err = io.Copy(uncompressedData, f.ReadCloser)
			if err != nil {
				return err
			}
			return archiver.ErrStopWalk
		}
		return nil
	})

	_ = os.Remove(rawData.Name()) // we don't need it anymore regardless of the error status
	if err != nil {
		_ = uncompressedData.Close()
		_ = os.Remove(uncompressedData.Name())

		return nil, err
	}

	_, err = uncompressedData.Seek(0, 0)
	if err != nil {
		_ = uncompressedData.Close()
		_ = os.Remove(uncompressedData.Name())

		return nil, err
	}

	return VolatileTempFile{file: uncompressedData}, nil
}

func (d DownloadableFile) SaveTo(path string) error {
	// Get the data
	pluginData, err := d.Open()
	if err != nil {
		return err
	}
	defer func() { _ = pluginData.Close() }()

	// Create the file
	err = os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	defer func() { _ = os.Chmod(path, 0755) }()

	// Write the body to file
	_, err = io.Copy(out, pluginData)
	return err
}

func (d DownloadableFile) ReadAll() ([]byte, error) {
	reader, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	return ioutil.ReadAll(reader)
}

func (d DownloadableFile) ReadAllWithCache(cacheDir string, refresh time.Duration) ([]byte, time.Time, error) {
	if UrlIsRemote(d.Url) {
		indexCachePath := filepath.Join(cacheDir, HashString(d.Url))
		cacheData, errCacheData := ioutil.ReadFile(indexCachePath)
		cacheMeta, errCacheMeta := os.Stat(indexCachePath)
		var cacheTimestamp time.Time
		if errCacheMeta != nil {
			cacheTimestamp = time.Unix(0, 0)
		} else {
			cacheTimestamp = cacheMeta.ModTime()
		}

		if cacheTimestamp.Before(time.Now().Add(-refresh)) || errCacheData != nil {
			freshData, errFreshData := d.ReadAll()
			if errFreshData != nil {
				if errCacheData != nil {
					return nil, time.Now(), errFreshData // we're not sure our cache is valid and failed to fetch so just fail
				}
				// failed to fetch but there is _some_ kind of a cache - just return it
				return cacheData, cacheTimestamp, nil
			}
			// we fetched fresh data - let's try to cache it but don't sweat if fail
			_ = os.MkdirAll(cacheDir, 0755)
			_ = ioutil.WriteFile(indexCachePath, freshData, 0644)
			return freshData, time.Now(), nil
		} else {
			return cacheData, cacheTimestamp, nil
		}
	} else {
		data, err := d.ReadAll()
		return data, time.Now(), err
	}
}
