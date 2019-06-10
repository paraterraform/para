package transport

import (
	"fmt"
	"github.com/mholt/archiver"
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

func OpenUrl(url string) (io.ReadCloser, error) {
	var err error
	var reader io.ReadCloser

	if strings.HasPrefix(url, schemaHttp) || strings.HasPrefix(url, schemaHttps) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		reader = resp.Body
	} else {
		path := url

		if strings.HasPrefix(url, schemaFile) {
			path = url[len(schemaFile):]
		}

		reader, err = os.Open(path)
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

type VolatileTempFile struct {
	file *os.File
}

func (f VolatileTempFile) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f VolatileTempFile) Close() error {
	defer func() { _ = os.Remove(f.file.Name()) }()
	return f.file.Close()
}
