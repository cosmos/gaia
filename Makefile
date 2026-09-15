#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in "github.com/cometbft/cometbft v0.34.7"
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
TEST_DOCKER_REPO=cosmos/contrib-gaiatest

GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION = 1.26

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(GAIA_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=gaia \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=gaiad \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TM_VERSION)

UNAME_S = $(shell uname -s)
ifeq ($(UNAME_S),Linux)
  extldflags += -z noexecstack
endif
ifeq (cleveldb,$(findstring cleveldb,$(GAIA_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
  extldflags += -Wl,-z,muldefs -static -z noexecstack
  ldflags += -linkmode=external
endif
ifeq (,$(findstring nostrip,$(GAIA_BUILD_OPTIONS)))
  ldflags += -w -s
endif
extldflags += $(EXTLDFLAGS)
extldflags := $(strip $(extldflags))
ldflags += -extldflags "$(extldflags)"
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(GAIA_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])

# The below include contains the tools target.
include contrib/devtools/Makefile

###############################################################################
###                                 Help                                    ###
###############################################################################

#? help: List available make targets
help: Makefile
	@echo " Choose a command to run in gaia:"
	@sed -n 's/^#?//p' $< | column -t -s ':' | sort | sed -e 's/^/ /'
.PHONY: help

###############################################################################
###                              Build                                      ###
###############################################################################

check_version:
ifneq ($(shell [ "$(GO_SYSTEM_VERSION)" \< "$(REQUIRE_GO_VERSION)" ] && echo true),)
	@echo "ERROR: Minimum Go version $(REQUIRE_GO_VERSION) is required for $(VERSION) of Gaia."
	exit 1
endif

#? all: Run install, lint, unit tests, e2e tests and vulncheck
all: install lint run-tests test-e2e vulncheck

BUILD_TARGETS := build install

#? build: Build the gaiad binary into ./build
#? install: Build and install gaiad into GOBIN
build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check_version go.sum $(BUILDDIR)/
	CGO_CFLAGS="-D__BLST_PORTABLE__" go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

#? vulncheck: Run govulncheck against all packages
vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@echo "--> Ensure dependencies have not been modified unless suppressed by SKIP_MOD_VERIFY"
ifndef SKIP_MOD_VERIFY
	go mod verify
endif
	go mod tidy
	@echo "--> Download go modules to local cache"
	go mod download

#? draw-deps: Render a dependency graph to dependency-graph.png (requires graphviz)
draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/gaiad -d 2 | dot -Tpng -o dependency-graph.png

#? clean: Remove build artifacts
clean:
	rm -rf $(BUILDDIR)/ artifacts/

#? distclean: Remove build artifacts and vendor/
distclean: clean
	rm -rf vendor/

###############################################################################
###                              Version Bump                               ###
###############################################################################

# Bump the major version number in go.mod and all import paths
# Usage: make bump-version OLD_VERSION=v25.2.0 NEW_VERSION=v26.0.0
#? bump-version: Bump the major version in go.mod and import paths (OLD_VERSION=... NEW_VERSION=...)
bump-version:
ifndef OLD_VERSION
	$(error OLD_VERSION is required. Usage: make bump-version OLD_VERSION=v25.2.0 NEW_VERSION=v26.0.0)
endif
ifndef NEW_VERSION
	$(error NEW_VERSION is required. Usage: make bump-version OLD_VERSION=v25.2.0 NEW_VERSION=v26.0.0)
endif
	$(eval OLD_MAJOR := $(shell echo $(OLD_VERSION) | sed 's/^v//' | cut -d. -f1))
	$(eval NEW_MAJOR := $(shell echo $(NEW_VERSION) | sed 's/^v//' | cut -d. -f1))
	$(eval OLD_FULL := $(shell echo $(OLD_VERSION) | sed 's/^v//'))
	$(eval NEW_FULL := $(shell echo $(NEW_VERSION) | sed 's/^v//'))
	@echo "--> Bumping version from v$(OLD_MAJOR) to v$(NEW_MAJOR)"
	@echo "--> Updating go.mod module path"
	@sed -i.bak 's|github.com/cosmos/gaia/v$(OLD_MAJOR)|github.com/cosmos/gaia/v$(NEW_MAJOR)|g' go.mod && rm go.mod.bak
	@echo "--> Updating import paths in Go files"
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -exec sed -i.bak 's|github.com/cosmos/gaia/v$(OLD_MAJOR)|github.com/cosmos/gaia/v$(NEW_MAJOR)|g' {} \; -exec rm {}.bak \;
	@echo "--> Updating tests/interchain/go.mod if it exists"
	@if [ -f tests/interchain/go.mod ]; then \
		sed -i.bak 's|github.com/cosmos/gaia/v$(OLD_MAJOR)|github.com/cosmos/gaia/v$(NEW_MAJOR)|g' tests/interchain/go.mod && rm tests/interchain/go.mod.bak; \
	fi
	@echo "--> Running go mod tidy"
	@go mod tidy
	@echo "--> Running go mod tidy in tests/interchain"
	@if [ -f tests/interchain/go.mod ]; then \
		cd tests/interchain && go mod tidy; \
	fi
	@echo "--> Updating old gaiad version in .github/workflows/test.yml to v$(OLD_FULL)"
	@sed -i.bak 's|download/v[0-9]*\.[0-9]*\.[0-9]*/gaiad-v[0-9]*\.[0-9]*\.[0-9]*|download/v$(OLD_FULL)/gaiad-v$(OLD_FULL)|g' .github/workflows/test.yml && rm .github/workflows/test.yml.bak
	@echo "--> Version bump complete!"
	@echo "    Remember to:"
	@echo "    1. Create new upgrade handler directory (app/upgrades/v$(NEW_MAJOR)_0_0/)"
	@echo "    2. Update CHANGELOG.md"
	@echo "    3. Register the new upgrade in app/app.go"

###############################################################################
###                                Release                                  ###
###############################################################################

GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
GORELEASER_IMAGE := ghcr.io/goreleaser/goreleaser-cross:v$(REQUIRE_GO_VERSION)
COSMWASM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm/v2 | sed 's/.* //')

# create tag and run goreleaser without publishing
# errors are possible while running goreleaser - the process can run for >30 min
# if the build is failing due to timeouts use goreleaser-build-local instead
#? create-release-dry-run: Run goreleaser for TAG without publishing
create-release-dry-run:
ifneq ($(strip $(TAG)),)
	@echo "--> Dry running release for tag: $(TAG)"
	@echo "--> Create tag: $(TAG) dry run"
	git tag -s $(TAG) -m $(TAG)
	git push origin $(TAG) --dry-run
	@echo "--> Delete local tag: $(TAG)"
	@git tag -d $(TAG)
	@echo "--> Running goreleaser"
	@go install github.com/goreleaser/goreleaser@latest
	@docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v `pwd`:/go/src/gaiad \
		-w /go/src/gaiad \
		$(GORELEASER_IMAGE) \
		release \
		--snapshot \
		--skip=publish \
		--verbose \
		--clean
	@rm -rf dist/
	@echo "--> Done create-release-dry-run for tag: $(TAG)"
else
	@echo "--> No tag specified, skipping tag release"
endif

# Build static binaries for linux/amd64 using docker buildx
# Pulled from neutron-org/neutron: https://github.com/neutron-org/neutron/blob/v4.2.2/Makefile#L107
#? build-static-linux-amd64: Build a static linux/amd64 binary with docker buildx
build-static-linux-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name gaiabuilder || true
	$(DOCKER) buildx use gaiabuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),muslc \
		--build-arg CGO_CFLAGS="-D__BLST_PORTABLE__" \
		--platform linux/amd64 \
		-t gaiad-static-amd64 \
		-f Dockerfile . \
		--load
	$(DOCKER) rm -f gaiabinary || true
	$(DOCKER) create -ti --name gaiabinary gaiad-static-amd64
	$(DOCKER) cp gaiabinary:/usr/local/bin/ $(BUILDDIR)/gaiad-linux-amd64
	$(DOCKER) rm -f gaiabinary

# Build static binaries for linux/arm64 using docker buildx
# Pulled from neutron-org/neutron: https://github.com/neutron-org/neutron/blob/v4.2.2/Makefile#L107
#? build-static-linux-arm64: Build a static linux/arm64 binary with docker buildx
build-static-linux-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name gaiabuilder || true
	$(DOCKER) buildx use gaiabuilder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),muslc \
		--build-arg CGO_CFLAGS="-D__BLST_PORTABLE__" \
		--platform linux/arm64 \
		-t gaiad-static-arm64 \
		-f Dockerfile . \
		--load
	$(DOCKER) rm -f gaiabinary || true
	$(DOCKER) create -ti --name gaiabinary gaiad-static-arm64
	$(DOCKER) cp gaiabinary:/usr/local/bin/ $(BUILDDIR)/gaiad-linux-arm64
	$(DOCKER) rm -f gaiabinary


# uses goreleaser to create static binaries for darwin on local machine
#? goreleaser-build-local: Build static darwin/linux binaries locally with goreleaser
goreleaser-build-local:
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v `pwd`:/go/src/gaiad \
		-w /go/src/gaiad \
		$(GORELEASER_IMAGE) \
		release \
		--snapshot \
		--skip=publish \
		--release-notes ./RELEASE_NOTES.md \
		--timeout 90m \
		--verbose

# uses goreleaser to create static binaries for linux an darwin
# requires access to GITHUB_TOKEN which has to be available in the CI environment
ifdef GITHUB_TOKEN
ci-release:
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v `pwd`:/go/src/gaiad \
		-w /go/src/gaiad \
		$(GORELEASER_IMAGE) \
		release \
		--release-notes ./RELEASE_NOTES.md \
		--timeout=90m \
		--clean
else
ci-release:
	@echo "Error: GITHUB_TOKEN is not defined. Please define it before running 'make release'."
endif

# create tag and publish it
#? create-release: Create and push a signed release tag (TAG=vX.Y.Z)
create-release:
ifneq ($(strip $(TAG)),)
	@echo "--> Running release for tag: $(TAG)"
	@echo "--> Create release tag: $(TAG)"
	git tag -s $(TAG) -m $(TAG)
	git push origin $(TAG)
	@echo "--> Done creating release tag: $(TAG)"
else
	@echo "--> No tag specified, skipping create-release"
endif

###############################################################################
###                              Documentation                              ###
###############################################################################

#? build-docs: Build the documentation site
build-docs:
	@cd docs && ./build.sh

.PHONY: build-docs


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

include sims.mk

PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e')
PACKAGES_E2E=$(shell cd tests/e2e && go list ./... | grep '/e2e')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race test-e2e

#? mocks: Regenerate test mocks and format the result
mocks: gen-mocks format

gen-mocks:
	@echo "--> generating mocks"
	@go install github.com/vektra/mockery/v2
	@go run github.com/vektra/mockery/v2

#? test-unit: Run unit tests
test-unit: ARGS=-timeout=5m -tags='norace'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
#? test-unit-cover: Run unit tests with coverage (coverage.txt)
test-unit-cover: ARGS=-timeout=5m -tags='norace' -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
#? test-race: Run unit tests with the race detector
test-race: ARGS=-timeout=5m -race
test-race: TEST_PACKAGES=$(PACKAGES_UNIT)
#? test-e2e: Run end-to-end tests (requires docker)
test-e2e: ARGS=-timeout=35m -v
test-e2e: TEST_PACKAGES=$(PACKAGES_E2E)
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@echo "--> Running tests"
	@go test -mod=readonly -json $(ARGS) $(TEST_PACKAGES) | tparse
else
	@echo "--> Running tests"
	@go test -mod=readonly $(ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests $(TEST_TARGETS)

#? docker-build-debug: Build the cosmos/gaiad-e2e docker image
docker-build-debug:
	@docker build -t cosmos/gaiad-e2e -f Dockerfile .

###############################################################################
###                                Linting                                  ###
###############################################################################
golangci_lint_cmd=golangci-lint
golangci_version=v2.11.1

#? lint: Run golangci-lint
lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m

#? lint-fix: Run golangci-lint and auto-fix issues
lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --fix --issues-exit-code=0

#? format: Format Go sources with gofumpt and golangci-lint
format:
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(golangci_version)
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	$(golangci_lint_cmd) run --fix
.PHONY: format

###############################################################################
###                                Localnet                                 ###
###############################################################################

#? start-localnet-ci: Build gaiad and start a single-node local chain
start-localnet-ci: build
	rm -rf ~/.gaiad-liveness
	./build/gaiad init liveness --chain-id liveness --home ~/.gaiad-liveness
	./build/gaiad config set client chain-id liveness --home ~/.gaiad-liveness
	./build/gaiad config set client keyring-backend test --home ~/.gaiad-liveness
	./build/gaiad keys add val --home ~/.gaiad-liveness --keyring-backend test
	./build/gaiad genesis add-genesis-account val 10000000000000000000000000stake --home ~/.gaiad-liveness --keyring-backend test
	./build/gaiad genesis gentx val 1000000000stake --home ~/.gaiad-liveness --chain-id liveness --keyring-backend test
	./build/gaiad genesis collect-gentxs --home ~/.gaiad-liveness
	sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' ~/.gaiad-liveness/config/app.toml
	./build/gaiad start --home ~/.gaiad-liveness

.PHONY: start-localnet-ci

###############################################################################
###                                Docker                                   ###
###############################################################################

#? test-docker: Build the test docker image
test-docker:
	@docker build -f contrib/Dockerfile.test -t ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) .
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:latest

#? test-docker-push: Build and push the test docker image
test-docker-push: test-docker
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD)
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker push ${TEST_DOCKER_REPO}:latest

.PHONY: all build-linux install format lint draw-deps clean build \
	docker-build-debug


###############################################################################
###                                Protobuf                                 ###
###############################################################################
protoVer=0.15.2
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

#? proto-all: Format, lint and generate protobuf files
proto-all: proto-format proto-lint proto-gen

#? proto-gen: Generate Go code from protobuf definitions
proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./proto/scripts/protocgen.sh

#? proto-swagger-gen: Generate swagger docs from protobuf definitions
proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./proto/scripts/protoc-swagger-gen.sh

#? proto-format: Format protobuf files with clang-format
proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

#? proto-lint: Lint protobuf files with buf
proto-lint:
	@$(protoImage) buf lint --error-format=json

#? proto-check-breaking: Check protobuf files for breaking changes against main
proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

#? proto-update-deps: Update protobuf dependencies with buf
proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps

###############################################################################
###                                Localnet                                 ###
###############################################################################

localnet-build-env:
	$(MAKE) -C contrib/images gaiad-env

localnet-build-nodes:
	$(DOCKER) run --rm -v $(CURDIR)/.testnets:/data cosmos/gaiad \
			  testnet init-files --v 4 -o /data --starting-ip-address 192.168.10.2 --keyring-backend=test --chain-id=localchain --use-docker=true
	docker compose up -d

#? localnet-stop: Stop the docker localnet
localnet-stop:
	docker compose down

# localnet-start will run a 4-node testnet locally. The nodes are
# based off the docker images in: ./contrib/images/simd-env
#? localnet-start: Start a 4-node docker localnet
localnet-start: localnet-stop localnet-build-env localnet-build-nodes


.PHONY: localnet-start localnet-stop localnet-build-env localnet-build-nodes