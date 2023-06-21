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
REQUIRE_GO_VERSION = 1.20

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

ifeq (cleveldb,$(findstring cleveldb,$(GAIA_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (,$(findstring nostrip,$(GAIA_BUILD_OPTIONS)))
  ldflags += -w -s
endif
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
###                              Build                                      ###
###############################################################################

check_version:
ifneq ($(GO_SYSTEM_VERSION), $(REQUIRE_GO_VERSION))
	@echo "ERROR: Go version 1.20 is required for $(VERSION) of Gaia."
	exit 1
endif

all: install lint run-tests test-e2e vulncheck

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check_version go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go install github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/gaiad -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf $(BUILDDIR)/ artifacts/

distclean: clean
	rm -rf vendor/

###############################################################################
###                                Release                                  ###
###############################################################################

# create tag and run goreleaser without publishing
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
	TM_VERSION=$(TM_VERSION) goreleaser release --snapshot --clean
	@rm -rf dist/
	@echo "--> Done create-release-dry-run for tag: $(TAG)"
else
	@echo "--> No tag specified, skipping tag release"
endif

# create tag and publish it
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
###                                 Devdoc                                  ###
###############################################################################

build-docs:
	@cd docs && \
	while read p; do \
		(git checkout $${p} && npm install && VUEPRESS_BASE="/$${p}/" npm run build) ; \
		mkdir -p ~/output/$${p} ; \
		cp -r .vuepress/dist/* ~/output/$${p}/ ; \
		cp ~/output/$${p}/index.html ~/output ; \
	done < versions ;
.PHONY: build-docs

sync-docs:
	cd ~/output && \
	echo "role_arn = ${DEPLOYMENT_ROLE_ARN}" >> /root/.aws/config ; \
	echo "CI job = ${CIRCLE_BUILD_URL}" >> version.html ; \
	aws s3 sync . s3://${WEBSITE_BUCKET} --profile terraform --delete ; \
	aws cloudfront create-invalidation --distribution-id ${CF_DISTRIBUTION_ID} --profile terraform --path "/*" ;
.PHONY: sync-docs


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

include sims.mk

PACKAGES_UNIT=$(shell go list ./... | grep -v -e '/tests/e2e')
PACKAGES_E2E=$(shell cd tests/e2e && go list ./... | grep '/e2e')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race test-e2e

test-unit: ARGS=-timeout=5m -tags='norace'
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)
test-unit-cover: ARGS=-timeout=5m -tags='norace' -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)
test-race: ARGS=-timeout=5m -race
test-race: TEST_PACKAGES=$(PACKAGES_UNIT)
test-e2e: ARGS=-timeout=25m -v
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

docker-build-debug:
	@docker build -t cosmos/gaiad-e2e -f e2e.Dockerfile .

# TODO: Push this to the Cosmos Dockerhub so we don't have to keep building it
# in CI.
docker-build-hermes:
	@cd tests/e2e/docker; docker build -t cosmos/hermes-e2e:latest -f hermes.Dockerfile .

docker-build-all: docker-build-debug docker-build-hermes

###############################################################################
###                                Linting                                  ###
###############################################################################
golangci_lint_cmd=golangci-lint
golangci_version=v1.52.2

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m

lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0

format:
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go" -not -path "./crypto/keys/secp256k1/*" | xargs gofumpt -w -l
	$(golangci_lint_cmd) run --fix
.PHONY: format

###############################################################################
###                                Localnet                                 ###
###############################################################################

start-localnet-ci: build
	rm -rf ~/.gaiad-liveness
	./build/gaiad init liveness --chain-id liveness --home ~/.gaiad-liveness
	./build/gaiad config chain-id liveness --home ~/.gaiad-liveness
	./build/gaiad config keyring-backend test --home ~/.gaiad-liveness
	./build/gaiad keys add val --home ~/.gaiad-liveness
	./build/gaiad add-genesis-account val 10000000000000000000000000stake --home ~/.gaiad-liveness --keyring-backend test
	./build/gaiad gentx val 1000000000stake --home ~/.gaiad-liveness --chain-id liveness
	./build/gaiad collect-gentxs --home ~/.gaiad-liveness
	sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' ~/.gaiad-liveness/config/app.toml
	./build/gaiad start --home ~/.gaiad-liveness --x-crisis-skip-assert-invariants

.PHONY: start-localnet-ci

###############################################################################
###                                Docker                                   ###
###############################################################################

test-docker:
	@docker build -f contrib/Dockerfile.test -t ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) .
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:latest

test-docker-push: test-docker
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD)
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker push ${TEST_DOCKER_REPO}:latest

.PHONY: all build-linux install format lint go-mod-cache draw-deps clean build \
	docker-build-debug docker-build-hermes docker-build-all


###############################################################################
###                                Protobuf                                 ###
###############################################################################
proto-gen:
	@echo "Generating Protobuf files"
	@sh ./proto/scripts/protocgen.sh

proto-doc:
	@echo "Generating Protoc docs"
	@sh ./proto/scripts/protoc-doc-gen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@sh ./proto/scripts/protoc-swagger-gen.sh

.PHONY: proto-gen proto-doc proto-swagger-gen
