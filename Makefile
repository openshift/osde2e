.PHONY: build-image

PKG := github.com/openshift/osde2e

IMAGE_NAME := quay.io/app-sre/osde2e
IMAGE_TAG := $(shell git rev-parse --short=7 HEAD)

build-image:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

push-image:
	docker push "$(IMAGE_NAME):$(IMAGE_TAG)"

out/osde2e: out
	go build -v -o $@ $(PKG)/cmd/osde2e

out:
	mkdir -p $@
