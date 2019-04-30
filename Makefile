.PHONY: build-image

PKG := github.com/openshift/osde2e

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

build-image:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):$(IMAGE_TAG)"

push-latest:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(IMAGE_NAME):latest"
	@docker --config=$(DOCKER_CONF) push "$(IMAGE_NAME):latest"

out/osde2e: out
	go build -v -o $@ $(PKG)/cmd/osde2e

out:
	mkdir -p $@
