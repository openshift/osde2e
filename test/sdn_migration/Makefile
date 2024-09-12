.DEFAULT_GOAL := build

CONTAINER_ENGINE ?= podman

clean:
	rm -rf *.test

GOFLAGS=-mod=mod
build: clean
	CGO_ENABLED=0 go test -v -c

build-image:
	${CONTAINER_ENGINE} build -t validate-sdn-ovn-migration:latest .
