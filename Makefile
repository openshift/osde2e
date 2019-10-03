.PHONY: check generate build-image push-image push-latest test

PKG := github.com/openshift/osde2e
DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

check: cmd/osde2e-docs
	go run $(PKG)/$< --check
	CGO_ENABLED=0 go test -v ./cmd/... ./pkg/... 
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint && \
    golangci-lint run ./... 

generate: docs/Options.md

build-image:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

test: out/osde2e
	$< -test.v -ginkgo.skip="$(GINKGO_SKIP)" -test.timeout 8h

docker-test:
	docker run \
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
		-e TARGET_STREAM=$(TARGET_STREAM) \
		-e NO_DESTROY_DELAY=$(NO_DESTROY_DELAY) \
		-e GINKGO_SKIP="$(GINKGO_SKIP)" \
		-e UPGRADE_RELEASE_STREAM=$(UPGRADE_RELEASE_STREAM) \
		-e DEBUG_OSD=1 \
		-e OSD_ENV=$(OSD_ENV) \
		-e UHC_TOKEN=$(UHC_REFRESH_TOKEN) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-e TESTGRID_BUCKET=$(TESTGRID_BUCKET) \
		-e TESTGRID_PREFIX=$(TESTGRID_PREFIX) \
		-e TESTGRID_SERVICE_ACCOUNT=$(TESTGRID_SERVICE_ACCOUNT) \
		$(IMAGE_NAME):$(IMAGE_TAG)

out/osde2e: out
	CGO_ENABLED=0 go test -v -c -o $@ $(PKG)

out/osde2e-report: out
	CGO_ENABLED=0 go build -v -o $@ $(PKG)

out:
	mkdir -p $@

docs/Options.md: cmd/osde2e-docs pkg/config/config.go
	go run $(PKG)/$<
