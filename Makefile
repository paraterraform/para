NAME := para
PLATFORMS := darwin/amd64 linux/amd64
VERSION = $(shell git describe 1>/dev/null 2>/dev/null && echo "_$$(git describe)")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

BASE := $(NAME)$(VERSION)
RELEASE_DIR := ./release

all: clean test release

clean:
	rm -rf $(RELEASE_DIR) ./$(NAME)

format:
	GOPROXY="off" GOFLAGS="-mod=vendor" go fmt ./...

test:
	GOPROXY="off" GOFLAGS="-mod=vendor" go test -v ./...
	GOPROXY="off" GOFLAGS="-mod=vendor" go vet ./...

build:
	GOPROXY="off" GOFLAGS="-mod=vendor" go build -o $(NAME)

run:
	GOPROXY="off" GOFLAGS="-mod=vendor" go run . $(args)

release: $(PLATFORMS)

$(PLATFORMS):
	GOPROXY="off" GOFLAGS="-mod=vendor" GOOS=$(os) GOARCH=$(arch) go build -ldflags="-s -w" -o '$(RELEASE_DIR)/$(BASE)_$(os)-$(arch)'

compress:
	upx $(RELEASE_DIR)/*

.PHONY: $(PLATFORMS) release build test fmt clean all
