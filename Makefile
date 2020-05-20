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

check:
	export GOPRIVATE="github.com/openshift/moactl"
	CGO_ENABLED=0 go test -v $(PKG)/cmd/... $(PKG)/pkg/...
	
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.23.8
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)
	find "$(DIR)scripts" -name "*.sh" -exec $(DIR)scripts/shellcheck.sh {} +

build-image:
	$(CONTAINER_ENGINE) build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	$(CONTAINER_ENGINE) tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

build:
	mkdir -p "$(OUT_DIR)"
	go build -o "$(OUT_DIR)" "$(DIR)cmd/..."

test: build
	"$(OSDE2E)" test -configs=e2e-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-informing: build
	"$(OSDE2E)" test -configs=informing-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-scale: build
	"$(OSDE2E)" test -configs=scale-mastervertical-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-addons: build
	"$(OSDE2E)" test -configs=addon-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-conformance: build
	"$(OSDE2E)" test -configs=conformance-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-middle-imageset: build
	"$(OSDE2E)" test -configs=e2e-suite,use-middle-version,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-oldest-imageset: build
	"$(OSDE2E)" test -configs=e2e-suite,use-oldest-version,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-docker:
	$(CONTAINER_ENGINE) run \
		-t \
		--rm \
		-e NO_DESTROY=$(NO_DESTROY) \
		-e CLUSTER_ID=$(CLUSTER_ID) \
		-e CLUSTER_NAME=$(CLUSTER_NAME) \
		-e CLEAN_RUNS=$(CLEAN_RUNS) \
		-e DRY_RUN=$(DRY_RUN) \
		-e MAJOR_TARGET=$(MAJOR_TARGET) \
		-e MINOR_TARGET=$(MINOR_TARGET) \
		-e CLUSTER_VERSION=$(CLUSTER_VERSION) \
		-e NO_DESTROY_DELAY=$(NO_DESTROY_DELAY) \
		-e GINKGO_SKIP="$(GINKGO_SKIP)" \
		-e GINKGO_FOCUS="$(GINKGO_FOCUS)" \
		-e UPGRADE_RELEASE_STREAM=$(UPGRADE_RELEASE_STREAM) \
		-e DEBUG_OSD=1 \
		-e OSD_ENV=$(OSD_ENV) \
		-e OCM_TOKEN=$(OCM_REFRESH_TOKEN) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		$(IMAGE_NAME):$(IMAGE_TAG)
