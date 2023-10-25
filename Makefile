SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

SHELL=/bin/bash -o pipefail

SRC 			:= $(abspath ./src)
BUILD			:= $(abspath ./bin)
tools_bin_path 	:= $(abspath ./_tools/bin)

org_package_root           := github.com/mmihic
repo_package               := $(org_package_root)/golib

BUILD_MOD_VENDOR          ?= true
VENDOR                    := $(repo_package)/$(vendor_prefix)
GO_BUILD_TAGS_LIST        :=
GO_BUILD_COMMON_ENV       ?= CGO_ENABLED=0
GO_PATH                   := ${GOPATH}
OS                        := $(shell uname | tr '[[:upper:]]' '[[:lower:]]')
ARCH                      := $(shell uname -m | sed "s/x86_64/amd64/")
LINUX_AMD64_ENV           := GOOS=linux GOARCH=amd64 $(GO_BUILD_COMMON_ENV)
DARWIN_AMD64_ENV          := GOOS=darwin GOARCH=amd64 $(GO_BUILD_COMMON_ENV)
DARWIN_ARM64_ENV          := GOOS=darwin GOARCH=arm64 $(GO_BUILD_COMMON_ENV)

GIT_REVISION              := $(shell git rev-parse --short=8 HEAD)
GIT_BRANCH                := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE                := $(shell date -u  +"%Y-%m-%dT%H:%M:%SZ") # Use RFC-3339 date format
BUILD_TS_UNIX             := $(shell date '+%s') # second since epoch

# NB: To add a new Go tool, first add the name of the tool to the GO_TOOLS variable
# below and then define the package that we can install the tool from as a variable
# of the form GO_TOOL_<NAME>_PKG.

GO_TOOLS :=     	\
	fossa         	\
	gocov         	\
	golangci-lint

GO_TOOL_FOSSA_PKG         := github.com/fossas/fossa-cli/cmd/fossa
GO_TOOL_GOCOV_PKG         := github.com/axw/gocov/gocov
GO_TOOL_GOLANGCI-LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

# install tools
define GO_TOOLS_RULES
.PHONY: install-$(TOOL)
install-$(TOOL):
	@echo "--- :toolbox: Installing Go tool $(TOOL)"
	which $(tools_bin_path)/$(TOOL) &> /dev/null || (cd tools && GOBIN=$(tools_bin_path) go install -tags tools $(GO_TOOL_$(shell echo fossa | tr '[:lower:]' '[:upper:]')_PKG))
endef

$(foreach TOOL,$(GO_TOOLS),$(eval $(GO_TOOLS_RULES)))

.PHONY: install-go-tools
install-go-tools: $(foreach v,$(GO_TOOLS), install-$(v))

# top-level rules
.PHONY: setup
setup:
	mkdir -p $(BUILD)

# install all tools
install-tools: install-go-tools

# Runs a fossa license report
.PHONY: fossa
fossa: install-fossa
	$(tools_bin_path)/fossa

# Runs linting
lint: install-golangci-lint lint-no-deps

.PHONY: lint-no-deps
lint-no-deps: export GO_BUILD_TAGS = $(GO_BUILD_TAGS_LIST)
lint-no-deps: export CGO_ENABLED=0
lint-no-deps:
	@echo "--- :golang: Running linters"
	@PATH=./bin:$(realpath $(tools_bin_path)):$(PATH) ./scripts/run_linter.sh

# Runs tests
test:
	go test $(test_args) -parallel 4 ./src/...

# Runs tests with coverage
test-cover:
	test_args="-cover -coverprofile=cover.out" make test

# Cleans up binaries
clean:
	go clean -testcache
	rm -rf $(BUILD)

# Build all services
all: $(SERVICES)

.DEFAULT_GOAL := all



