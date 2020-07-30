.PHONY: check generate build-image push-image push-latest test

PKG := github.com/openshift/osde2e
DOC_PKG := $(PKG)/cmd/osde2e-docs

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OUT_DIR := $(DIR)out
OSDE2E := $(DIR)out/osde2e

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

CONTAINER_ENGINE ?= docker

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

check: diffproviders.txt
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.23.8
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)
	
	CGO_ENABLED=0 go test -v $(PKG)/cmd/... $(PKG)/pkg/...
	find "$(DIR)scripts" -name "*.sh" -exec $(DIR)scripts/shellcheck.sh {} +
	cmp -s diffproviders.txt "$(DIR)pkg/common/providers/providers_generated.go"

build-image:
	$(CONTAINER_ENGINE) build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	$(CONTAINER_ENGINE) tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

generate-providers:
	"$(DIR)scripts/generate-providers-import.sh" > "$(DIR)pkg/common/providers/providers_generated.go"

build:
	mkdir -p "$(OUT_DIR)"
	go build -o "$(OUT_DIR)" "$(DIR)cmd/..."

diffproviders.txt:
	"$(DIR)scripts/generate-providers-import.sh" > diffproviders.txt

.INTERMEDIATE: diffproviders.txt
