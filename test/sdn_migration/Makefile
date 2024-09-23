.DEFAULT_GOAL := build

CONTAINER_ENGINE ?= podman
OUT_DIR := $(DIR)out


build:
	mkdir -p "$(OUT_DIR)"
	go test -v -c -o "$(OUT_DIR)"/sdn_migration

build-image:
	${CONTAINER_ENGINE} build -t validate-sdn-ovn-migration:latest  -f ./Dockerfile ../../
