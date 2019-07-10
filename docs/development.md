# Development

## Go

In order to work on the provider, [Go](http://www.golang.org) should be installed first (version 1.11+ is *required*).
[goenv](https://github.com/syndbg/goenv) and [gvm](https://github.com/moovweb/gvm) are great utilities that can help a
lot with that and simplify setup tremendously. 
[GOPATH](http://golang.org/doc/code.html#GOPATH) should be setup correctly and `$GOPATH/bin` should be
added `$PATH`.

This plugin uses Go modules available starting from Go `1.11` and therefore it **should not** be checked out within `$GOPATH` tree.

## Source Code

Source code can be retrieved with `git`
```bash
$ git clone git@github.com:paraterraform/para.git .
```

## Dependencies

This project uses `go mod` to manage its dependencies and it's expected that all dependencies are vendored so that
it's buildable without internet access. When adding/removing a dependency run following commands:
```bash
$ go mod vendor
$ go mod tidy
```

## Build
In order to build plugin for the current platform use [GNU]make:
```bash
$ make build
GOPROXY="off" GOFLAGS="-mod=vendor" go build -o para

```

it will build provider from sources and put it into current working directory.

If Terraform was installed (as a binary) or via `go get -u github.com/hashicorp/terraform` it'll pick up the plugin if 
executed against a configuration in the same directory.

## Release

In order to prepare provider binaries for all platforms:
```bash
$ make release compress sums
make clean release compress sums
rm -rf ./release ./para
GOPROXY="off" GOFLAGS="-mod=vendor" GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o './release/para_v0.3.2_darwin-amd64'
GOPROXY="off" GOFLAGS="-mod=vendor" GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o './release/para_v0.3.2_linux-amd64'
upx ./release/*
                       Ultimate Packer for eXecutables
                          Copyright (C) 1996 - 2018
UPX 3.95        Markus Oberhumer, Laszlo Molnar & John Reiser   Aug 26th 2018

        File size         Ratio      Format      Name
   --------------------   ------   -----------   -----------
  11470148 ->   4272144   37.25%   macho/amd64   para_v0.3.2_darwin-amd64
  10602880 ->   3888952   36.68%   linux/amd64   para_v0.3.2_linux-amd64
   --------------------   ------   -----------   -----------
  22073028 ->   8161096   36.97%                 [ 2 files ]

Packed 2 files.
cd ./release; shasum -a 256 para* > SHA256SUMS
```

If you have [upx](https://upx.github.io) available you can compress release binaries with `make compress`.

## Versioning

This project follow [Semantic Versioning](https://semver.org/)

## Changelog

This project follows [keep a changelog](https://keepachangelog.com/en/1.0.0/) guidelines for changelog.
