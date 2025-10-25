.DEFAULT_GOAL := build

# globals
BINARY_NAME?=<nome_do_seu_binario>
BUILD_DIR?="./build"
CGO_ENABLED?=0

# basic Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# targets
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -v \
		-o ${BUILD_DIR}/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/api/main.go
