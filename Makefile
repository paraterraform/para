NAME := para
PLATFORMS := darwin/amd64 linux/amd64
VERSION = $(shell git describe 1>/dev/null 2>/dev/null && echo "_$$(git describe)")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

BASE := $(NAME)$(VERSION)
RELEASE_DIR := ./release

all: clean test release

.PHONY: clean
clean:
	rm -rf $(RELEASE_DIR)

.PHONY: format
format:
	GOPROXY="off" GOFLAGS="-mod=vendor" go fmt ./...

.PHONY: test
test:
	GOPROXY="off" GOFLAGS="-mod=vendor" go test -v ./...
	GOPROXY="off" GOFLAGS="-mod=vendor" go vet ./...

.PHONY: build
build:
	GOPROXY="off" GOFLAGS="-mod=vendor" go build -o $(NAME)

.PHONY: run
run:
	GOPROXY="off" GOFLAGS="-mod=vendor" go run . $(args)

.PHONY: release
release: $(PLATFORMS)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOPROXY="off" GOFLAGS="-mod=vendor" GOOS=$(os) GOARCH=$(arch) go build -ldflags="-s -w" -o '$(RELEASE_DIR)/$(BASE)_$(os)-$(arch)'

.PHONY: compress
compress:
	upx $(RELEASE_DIR)/*

.PHONY: sums
sums:
	cd $(RELEASE_DIR); shasum -a 256 para* > SHA256SUMS
