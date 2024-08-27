.PHONY: check generate test

PKG := github.com/openshift/osde2e
DOC_PKG := $(PKG)/cmd/osde2e-docs

DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OUT_DIR := $(DIR)out

ifndef $(GOPATH)
    GOPATH=$(shell go env GOPATH)
    export GOPATH
endif

fmt:
	gofmt -s -w .

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2
	(cd "$(DIR)"; golangci-lint run -c .golang-ci.yml ./...)

check: lint shellcheck vipercheck diffproviders.txt diffreporters.txt
	cmp -s diffproviders.txt "$(DIR)pkg/common/providers/providers_generated.go"
	cmp -s diffreporters.txt "$(DIR)pkg/reporting/reporters/reporters_generated.go"

	CGO_ENABLED=0 go test -v $(PKG)/cmd/... $(PKG)/pkg/...

shellcheck:
	find "$(DIR)scripts" -name "*.sh" -exec $(DIR)scripts/shellcheck.sh {} +

vipercheck:
	@if [ "$(shell go list -f '{{.Name}} {{.Imports}}' ./... | grep -v -E "^concurrentviper" | grep 'github.com/spf13/viper'| wc -l)" -gt 0 ]; then echo "Error: Code contains direct import of github.com/spf13/viper, use github.com/openshift/osde2e/pkg/common/concurrentviper instead." && exit 1; else echo "make vipercheck has passed, concurrentViper is being used."; fi

generate-providers:
	"$(DIR)scripts/generate-providers-import.sh" > "$(DIR)pkg/common/providers/providers_generated.go"

generate-reporters:
	"$(DIR)scripts/generate-reporters-import.sh" > "$(DIR)pkg/reporting/reporters/reporters_generated.go"

build:
	mkdir -p "$(OUT_DIR)"
	go build -o "$(OUT_DIR)" "$(DIR)cmd/..."

diffproviders.txt:
	"$(DIR)scripts/generate-providers-import.sh" > diffproviders.txt

diffreporters.txt:
	"$(DIR)scripts/generate-reporters-import.sh" > diffreporters.txt

.INTERMEDIATE: diffproviders.txt diffreporters.txt
