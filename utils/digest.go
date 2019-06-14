package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"sort"
	"strings"
)

type newHash = func() hash.Hash

var supportedHashes = map[string]newHash{
	"md5":    func() hash.Hash { return md5.New() },
	"sha1":   func() hash.Hash { return sha1.New() },
	"sha256": func() hash.Hash { return sha256.New() },
	"sha512": func() hash.Hash { return sha512.New() },
}

func DigestVerify(path, digest string) error {
	tokens := strings.SplitN(digest, ":", 2)
	if _, ok := supportedHashes[tokens[0]]; len(tokens) != 2 || !ok {
		var algsSlice []string
		for k := range supportedHashes {
			algsSlice = append(algsSlice, k)
		}
		sort.Strings(algsSlice)
		algs := strings.Join(algsSlice, "|")
		return fmt.Errorf("wrong digest format: '%s' does not match expected pattern '<%s>:<hash>'", digest, algs)
	}

	alg := tokens[0]
	expected := tokens[1]

	sink := supportedHashes[alg]()

	source, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	_, err = io.Copy(sink, source)
	if err != nil {
		return err
	}

	actual := fmt.Sprintf("%x", sink.Sum(nil))
	if actual != expected {
		return fmt.Errorf("actual %s hash value of %s does not match expected of %s", alg, actual, expected)
	}
	return nil
}
