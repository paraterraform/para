package xio

import (
	"fmt"
	"github.com/mholt/archiver"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	schemaFile  = "file://"
	schemaHttp  = "http://"
	schemaHttps = "https://"
)

func IsRemote(url string) bool {
	return strings.HasPrefix(url, schemaHttp) || strings.HasPrefix(url, schemaHttps)
}

func UrlOpen(url string) (io.ReadCloser, error) {
	var err error
	var reader io.ReadCloser

	if IsRemote(url) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf(
				"non-200 response while fetching '%s': %s",
				url, http.StatusText(resp.StatusCode),
			)
		}
		reader = resp.Body
	} else {
		path := url

		if strings.HasPrefix(url, schemaFile) {
			path = url[len(schemaFile):]
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

	rawData, err := ioutil.TempFile("", fmt.Sprintf("para.raw.*.%s", filepath.Base(url)))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(rawData, reader)
	_ = reader.Close() // we have to close it regardless of the error status

	if err != nil {
		_ = rawData.Close()
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	_, err = rawData.Seek(0, 0)
	if err != nil {
		_ = rawData.Close()
		_ = os.Remove(rawData.Name())

		return nil, err
	}

	_, err = archiver.ByExtension(rawData.Name())
	if err != nil {
		// not compressed or not supported - treat it as raw data
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
		_, err = io.Copy(uncompressedData, f.ReadCloser)
		if err != nil {
			return err
		}
		return archiver.ErrStopWalk
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

func UrlReadAll(url string) ([]byte, error) {
	reader, err := UrlOpen(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	return ioutil.ReadAll(reader)
}
