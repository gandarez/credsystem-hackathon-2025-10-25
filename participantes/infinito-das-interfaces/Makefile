.DEFAULT_GOAL := build

BINARY_NAME?=cs_ura_intent
BUILD_DIR?="./build"
CGO_ENABLED?=0

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
		-o ${BUILD_DIR}/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./main.go