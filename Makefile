.PHONY: build-image push-image push-latest test

PKG := github.com/openshift/osde2e
DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

build-image:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

test: out/osde2e
	$< -test.v -test.timeout 3h

docker-test:
	docker run \
		-t \
		--rm \
		-e CLUSTER_ID=$(CLUSTER_ID) \
		-e CLEAN_RUNS=$(CLEAN_RUNS) \
		-e UPGRADE_RELEASE_STREAM=$(UPGRADE_RELEASE_STREAM) \
		-e DEBUG_OSD=1 \
		-e USE_PROD=$(USE_PROD) \
		-e UHC_TOKEN=$(UHC_REFRESH_TOKEN) \
		-e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-e TESTGRID_BUCKET=$(TESTGRID_BUCKET) \
		-e TESTGRID_PREFIX=$(TESTGRID_PREFIX) \
		-e TESTGRID_SERVICE_ACCOUNT=$(TESTGRID_SERVICE_ACCOUNT) \
		$(IMAGE_NAME):$(IMAGE_TAG)

out/osde2e: out
	CGO_ENABLED=0 go test -v -c -o $@ $(PKG)

out:
	mkdir -p $@
