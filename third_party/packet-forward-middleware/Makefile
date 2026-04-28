#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --always) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
GOMOD := $(shell go list -m)
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
SIMAPP = ./testing/simapp
MOCKS_DIR = $(CURDIR)/test/mock
HTTPS_GIT := https://github.com/cosmos/ibc-go.git
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

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

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=sim \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=simd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

all: tools build lint test

# The below include contains the tools and runsim targets.
#include contrib/devtools/Makefile

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/
build-linux:
	GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false $(MAKE) build -o bin/simd ./cmd/simd

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build-simd-all: go.sum
	$(DOCKER) rm latest-build || true
	$(DOCKER) run --volume=$(CURDIR):/sources:ro \
        --env TARGET_PLATFORMS='linux/amd64 darwin/amd64 linux/arm64 windows/amd64' \
        --env APP=simd \
        --env VERSION=$(VERSION) \
        --env COMMIT=$(COMMIT) \
        --env LEDGER_ENABLED=$(LEDGER_ENABLED) \
        --name latest-build cosmossdk/rbuilder:latest
	$(DOCKER) cp -a latest-build:/home/builder/artifacts/ $(CURDIR)/

build-simd-linux: go.sum $(BUILDDIR)/
	$(DOCKER) rm latest-build || true
	$(DOCKER) run --volume=$(CURDIR):/sources:ro \
        --env TARGET_PLATFORMS='linux/amd64' \
        --env APP=simd \
        --env VERSION=$(VERSION) \
        --env COMMIT=$(COMMIT) \
        --env LEDGER_ENABLED=false \
        --name latest-build cosmossdk/rbuilder:latest
	$(DOCKER) cp -a latest-build:/home/builder/artifacts/ $(CURDIR)/
	cp artifacts/simd-*-linux-amd64 $(BUILDDIR)/simd

cosmovisor:
	$(MAKE) -C cosmovisor cosmovisor

.PHONY: build build-linux build-simd-all build-simd-linux cosmovisor

mocks: $(MOCKS_DIR)
	go install go.uber.org/mock/mockgen@v0.4.0
	mockgen -package=mock -destination=./test/mock/transfer_keeper.go $(GOMOD)/packetforward/types TransferKeeper
	mockgen -package=mock -destination=./test/mock/bank_keeper.go $(GOMOD)/packetforward/types BankKeeper
	mockgen -package=mock -destination=./test/mock/channel_keeper.go $(GOMOD)/packetforward/types ChannelKeeper
	mockgen -package=mock -destination=./test/mock/ics4_wrapper.go github.com/cosmos/ibc-go/v10/modules/core/05-port/types ICS4Wrapper
	mockgen -package=mock -destination=./test/mock/ibc_module.go github.com/cosmos/ibc-go/v10/modules/core/05-port/types IBCModule

.PHONY: mocks

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

distclean: clean tools-clean
clean:
	rm -rf \
    $(BUILDDIR)/ \
    artifacts/ \
    tmp-swagger-gen/

.PHONY: distclean clean

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit
test-all: test-unit test-ledger-mock test-race

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-tags='cgo ledger test_ledger_mock'
test-unit-amino: ARGS=-tags='ledger test_ledger_mock test_amino'
test-ledger: ARGS=-tags='cgo ledger'
test-ledger-mock: ARGS=-tags='ledger test_ledger_mock'
test-race: ARGS=-race -tags='cgo ledger test_ledger_mock'
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

# check-* compiles and collects tests without running them
# note: go test -c doesn't support multiple packages yet (https://github.com/golang/go/issues/15513)
CHECK_TEST_TARGETS := check-test-unit check-test-unit-amino
check-test-unit: ARGS=-tags='cgo ledger test_ledger_mock'
check-test-unit-amino: ARGS=-tags='ledger test_ledger_mock test_amino'
$(CHECK_TEST_TARGETS): EXTRA_ARGS=-run=none
$(CHECK_TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)

###############################################################################
###                             e2e interchain test                         ###
###############################################################################

local-image:
	docker build . -t pfm:local

ictest-forward:
	cd e2e && go test -race -v -timeout 30m -run TestPacketForwardMiddleware .

ictest-timeout:
	cd e2e && go test -race -v -timeout 30m -run TestTimeoutOnForward .

ictest-storage-leak:
	cd e2e && go test -race -v -timeout 30m -run TestStorageLeak .

ictest-upgrade:
	cd e2e && go test -race -v -timeout 30m -run TestPFMUpgrade .

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0
.PHONY: lint lint-fix

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -path "./tests/mocks/*" -not -name '*.pb.go' | xargs goimports -w -local github.com/cosmos/cosmos-sdk
.PHONY: format

###############################################################################
###                                Protobuf                                 ###
###############################################################################

containerProtoVer=v0.2
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(containerProtoVer)
containerProtoGenSwagger=cosmos-sdk-proto-gen-swagger-$(containerProtoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

protoVer=0.16.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protoc-swagger-gen.sh; fi

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

TM_URL              = https://raw.githubusercontent.com/tendermint/tendermint/v0.34.14/proto/tendermint
GOGO_PROTO_URL      = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
CONFIO_URL          = https://raw.githubusercontent.com/confio/ics23/v0.6.3
SDK_PROTO_URL 		= https://raw.githubusercontent.com/cosmos/cosmos-sdk/v0.44.3/proto/cosmos

TM_CRYPTO_TYPES     = third_party/proto/tendermint/crypto
TM_ABCI_TYPES       = third_party/proto/tendermint/abci
TM_TYPES            = third_party/proto/tendermint/types
TM_VERSION          = third_party/proto/tendermint/version
TM_LIBS             = third_party/proto/tendermint/libs/bits
TM_P2P              = third_party/proto/tendermint/p2p

SDK_QUERY 			= third_party/proto/cosmos/base/query/v1beta1
SDK_BASE 			= third_party/proto/cosmos/base/v1beta1
SDK_UPGRADE			= third_party/proto/cosmos/upgrade

GOGO_PROTO_TYPES    = third_party/proto/gogoproto
CONFIO_TYPES        = third_party/proto/confio

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

## Importing of tendermint protobuf definitions currently requires the
## use of `sed` in order to build properly with cosmos-sdk's proto file layout
## (which is the standard Buf.build FILE_LAYOUT)
## Issue link: https://github.com/tendermint/tendermint/issues/5021
	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto
	@curl -sSL $(TM_URL)/types/validator.proto > $(TM_TYPES)/validator.proto

	@mkdir -p $(TM_VERSION)
	@curl -sSL $(TM_URL)/version/types.proto > $(TM_VERSION)/types.proto

	@mkdir -p $(TM_LIBS)
	@curl -sSL $(TM_URL)/libs/bits/types.proto > $(TM_LIBS)/types.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

	@mkdir -p $(CONFIO_TYPES)
	@curl -sSL $(CONFIO_URL)/proofs.proto > $(CONFIO_TYPES)/proofs.proto

## insert go package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32
	@sed -i '4ioption go_package = "github.com/confio/ics23/go";' $(CONFIO_TYPES)/proofs.proto

.PHONY: proto-all proto-gen proto-gen-any proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps
