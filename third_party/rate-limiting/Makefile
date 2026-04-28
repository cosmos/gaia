#!/usr/bin/make -f
VERSION := $(shell echo $(shell git describe --tags))
BUILDDIR ?= $(CURDIR)/build
build=s
cache=false
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.7.0
STRIDE_HOME=./
DOCKERNET_HOME=./dockernet
DOCKERNET_COMPOSE_FILE=$(DOCKERNET_HOME)/docker-compose.yml
LOCALSTRIDE_HOME=./testutil/localstride
LOCALNET_COMPOSE_FILE=$(LOCALSTRIDE_HOME)/localnet/docker-compose.yml
STATE_EXPORT_COMPOSE_FILE=$(LOCALSTRIDE_HOME)/state-export/docker-compose.yml
LOCAL_TO_MAIN_COMPOSE_FILE=./scripts/local-to-mainnet/docker-compose.yml

# process build tags
LEDGER_ENABLED ?= true
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

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=stride \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=strided \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

.PHONY: build

###############################################################################
###                            Build & Clean                                ###
###############################################################################

build:
	which go
	mkdir -p $(BUILDDIR)/
	go build -mod=readonly $(BUILD_FLAGS) -trimpath -o $(BUILDDIR) ./...;

build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum
	go install $(BUILD_FLAGS) ./cmd/strided

clean:
	rm -rf $(BUILDDIR)/*

###############################################################################
###                                CI                                       ###
###############################################################################

gosec:
	gosec -exclude-dir=deps -severity=high ./...

golangci_lint_cmd=golangci-lint
golangci_version=v1.60.1

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m

###############################################################################
###                                Tests                                    ###
###############################################################################

test-unit:
	@go test -mod=readonly ./keeper/...


###############################################################################
###                                Protobuf                                 ###
###############################################################################

BUF_VERSION=1.31.0
BUILDER_VERSION=0.14.0

proto-all: proto-format proto-lint proto-gen

proto-format:
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write

proto-gen:
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./scripts/generate.sh

proto-lint:
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint

###############################################################################
###                             e2e interchain test                         ###
###############################################################################

local-image:
	docker build . -f Dockerfile -t rate-limiting:local

ictest-ratelimit:
	cd e2e && go test -race -v -timeout 15m -run TestRateLimitBasic .