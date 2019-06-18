package utils

import (
	"net/url"
	"path"
	"strings"
)

const (
	schemaFile  = "file://"
	schemaHttp  = "http://"
	schemaHttps = "https://"
)

func UrlIsRemote(url string) bool {
	return strings.HasPrefix(url, schemaHttp) || strings.HasPrefix(url, schemaHttps)
}

func UrlJoin(base string, elems ...string) string {
	u, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(append([]string{u.Path}, elems...)...)
	return u.String()
}
