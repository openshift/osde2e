.PHONY: check generate build-image push-image push-latest test

PKG := github.com/openshift/osde2e
DOC_PKG := $(PKG)/cmd/osde2e-docs
RUNNER_PKG := $(PKG)/runner

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

CONTAINER_ENGINE ?= docker

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

check:
	CGO_ENABLED=0 go test -v $(PKG)/cmd/... $(PKG)/pkg/...
	
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.21.0
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)

generate: $(DIR)/docs/Options.md

build-image:
	$(CONTAINER_ENGINE) build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	$(CONTAINER_ENGINE) tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@$(CONTAINER_ENGINE) --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

build:
	CGO_ENABLED=0 go test ./suites/e2e -v -c -o ./out/osde2e
	CGO_ENABLED=0 go test ./suites/scale -v -c -o ./out/osde2e-scale

test:
	go test $(RUNNER_PKG) -test.v -ginkgo.skip="$(GINKGO_SKIP)" -ginkgo.focus="$(GINKGO_FOCUS)" -test.timeout 8h -configs=e2e-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-scale:
	go test $(RUNNER_PKG) -test.v -ginkgo.skip="$(GINKGO_SKIP)" -ginkgo.focus="$(GINKGO_FOCUS)" -test.timeout 8h -configs=scale-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-addons:
	go test $(RUNNER_PKG) -test.v -ginkgo.skip="$(GINKGO_SKIP)" -ginkgo.focus="$(GINKGO_FOCUS)" -test.timeout 8h -configs=addon-suite,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-middle-imageset:
	go test $(RUNNER_PKG) -test.v -ginkgo.skip="$(GINKGO_SKIP)" -ginkgo.focus="$(GINKGO_FOCUS)" -test.timeout 8h -configs=e2e-suite,use-middle-version,log-metrics -custom-config=$(CUSTOM_CONFIG)

test-oldest-imageset:
	go test $(RUNNER_PKG) -test.v -ginkgo.skip="$(GINKGO_SKIP)" -ginkgo.focus="$(GINKGO_FOCUS)" -test.timeout 8h -configs=e2e-suite,use-oldest-version,log-metrics -custom-config=$(CUSTOM_CONFIG)

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

$(DIR)/docs/Options.md: $(DIR)/cmd/osde2e-docs $(DIR)/pkg/common/config/config.go
	go run $(DOC_PKG)
