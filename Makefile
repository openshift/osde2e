.PHONY: check generate build-image push-image push-latest test

PKG := github.com/openshift/osde2e
DOC_PKG := $(PKG)/cmd/osde2e-docs

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OUT_DIR := $(DIR)out
OSDE2E := $(DIR)out/osde2e

OSDE2E_IMAGE_NAME := quay.io/app-sre/osde2e
OSDE2ECTL_IMAGE_NAME := quay.io/app-sre/osde2ectl
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

CONTAINER_ENGINE ?= docker

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

check: shellcheck diffproviders.txt diffreporters.txt
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.23.8
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)
	cmp -s diffproviders.txt "$(DIR)pkg/common/providers/providers_generated.go"
	cmp -s diffreporters.txt "$(DIR)pkg/reporting/reporters/reporters_generated.go"
	
	CGO_ENABLED=0 go test -v $(PKG)/cmd/... $(PKG)/pkg/...

shellcheck:
	find "$(DIR)scripts" -name "*.sh" -exec $(DIR)scripts/shellcheck.sh {} +

build-image:
	$(CONTAINER_ENGINE) build -f "$(DIR)Dockerfile.osde2e" -t "$(OSDE2E_IMAGE_NAME):$(IMAGE_TAG)" .
	$(CONTAINER_ENGINE) build -f "$(DIR)Dockerfile.osde2ectl" -t "$(OSDE2ECTL_IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(OSDE2E_IMAGE_NAME):$(IMAGE_TAG)"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(OSDE2ECTL_IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	$(CONTAINER_ENGINE) tag "$(OSDE2E_IMAGE_NAME):$(IMAGE_TAG)" "$(OSDE2E_IMAGE_NAME):latest"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(OSDE2E_IMAGE_NAME):latest"
	$(CONTAINER_ENGINE) tag "$(OSDE2ECTL_IMAGE_NAME):$(IMAGE_TAG)" "$(OSDE2ECTL_IMAGE_NAME):latest"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(OSDE2ECTL_IMAGE_NAME):latest"

generate-providers:
	"$(DIR)scripts/generate-providers-import.sh" > "$(DIR)pkg/common/providers/providers_generated.go"

generate-reporters:
	"$(DIR)scripts/generate-reporters-import.sh" > "$(DIR)pkg/reporting/reporters/reporters_generated.go"

build: pkger
	mkdir -p "$(OUT_DIR)"
	go build -o "$(OUT_DIR)" "$(DIR)cmd/..."

pkger:
	GOFLAGS='' go get github.com/markbates/pkger/cmd/pkger
	pkger --include $(DIR)assets --include $(DIR)configs

diffproviders.txt:
	"$(DIR)scripts/generate-providers-import.sh" > diffproviders.txt

diffreporters.txt:
	"$(DIR)scripts/generate-reporters-import.sh" > diffreporters.txt

.INTERMEDIATE: diffproviders.txt diffreporters.txt
