.PHONY: check generate build-image push-image push-latest test

PKG := github.com/openshift/osde2e/plugins/ocmprovider/pkg/main

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OUT_DIR := $(DIR)out
OCMPROVIDER := $(DIR)out/ocmprovider.so

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

build: check
	mkdir -p "$(OUT_DIR)"
	go build -buildmode=plugin -o "$(OCMPROVIDER)" $(PKG)

check:
	go test -v $(PKG)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.23.8
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)

