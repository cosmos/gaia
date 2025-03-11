---
title: Installing Gaia
sidebar_position: 2
---

This guide will explain how to install the `gaiad` binary and run the cli. With this binary installed on a server, you can participate on the mainnet as either a [Full Node](../hub-tutorials/join-mainnet.md) or a [Validator](../validators/validator-setup.md).

## Build Requirements

At present, the SDK fully supports installation on linux distributions. For the purpose of this instruction set, we'll be using `Ubuntu 22.04 LTS`. It is also possible to install `gaiad` on Unix, while Windows may require additional unsupported third party installation. All steps are listed below for a clean install.

1. [Update & install build tools](#build-tools)
2. [Install Go](#install-go)
3. [Install `Gaiad` binaries](#install-the-binaries)

## Build Tools

Install `make` and `gcc`.

**Ubuntu:**

```bash
sudo apt update

sudo apt install -y make gcc build-essential
```

## Install Go

:::tip
**Go 1.23+** is required.
:::

We suggest the following two ways to install Go. Check out the [official docs](https://golang.org/doc/install) and Go installer for the correct download for your operating system. Alternatively, you can install Go yourself from the command line. Detailed below are standard default installation locations, but feel free to customize.

Since the introduction of CosmWasm in Gaia v18 it is recommended to build the binaries with `CGO` enabled - simply set `CGO_ENABLED=1` in your terminal before building the binary.

Building the `gaiad` binary on Windows is not supported due to [dependency issues](https://github.com/CosmWasm/wasmvm).

**[Go Binary Downloads](https://go.dev/dl/)**

**Ubuntu:**

At the time of this writing, the latest release is `1.23.6`. We're going to download the tarball, extract it to `/usr/local`, and export `GOROOT` to our `$PATH`

```bash
curl -OL https://go.dev/dl/go1.23.6.darwin-amd64.tar.gz

sudo tar -C /usr/local -xvf go1.23.6.darwin-amd64.tar.gz


export PATH=$PATH:/usr/local/go/bin

```

Remember to add `GOPATH` to your `$PATH` environment variable. If you're not sure where that is, run `go env GOPATH`. This will allow us to run the `gaiad` binary in the next step. If you're not sure how to set your `$PATH` take a look at [these instructions](https://superuser.com/questions/284342/what-are-path-and-other-environment-variables-and-how-can-i-set-or-use-them).

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Install the binaries

Next, let's install the latest version of Gaia. Make sure you `git checkout` the
correct [released version](https://github.com/cosmos/gaia/releases).

```bash
git clone -b <latest-release-tag> https://github.com/cosmos/gaia.git
cd gaia && make install
```

If this command fails due to the following error message, you might have already set `LDFLAGS` prior to running this step.

```sh
# github.com/cosmos/gaia/cmd/gaiad
flag provided but not defined: -L
usage: link [options] main.o
...
make: *** [install] Error 2
```

Unset this environment variable and try again.

```sh
LDFLAGS="" make install
```

> _NOTE_: If you still have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `gaiad` binary. Verify that everything installed successfully by running:

```bash
gaiad version --long
```

You should see something similar to the following:

```bash
name: gaia
server_name: gaiad
version: v18.0.0
commit: 682770f2410ab0d33ac7f0c7203519d7a99fa2b6
build_tags: netgo,ledger,muslc
go: go version go1.23.6 linux/amd64
```

## Docker

`Dockerfile` is available in the gaia repo.

Building:

```shell
git clone -b <latest-release-tag> https://github.com/cosmos/gaia.git
cd gaia
docker build -t cosmos-ics:local -f Dockerfile ./
```

## Static linking

In case you need to build a binary with all dependencies statically linked please check our [Dockerfile](https://github.com/cosmos/gaia/blob/main/Dockerfile).

You must have `libwasmvm` available on your machine.
Choose the build that matches your platform and OS:
* https://github.com/CosmWasm/wasmvm?tab=readme-ov-file#builds-of-libwasmvm

For more information, please check `wasmvm` [documentation](https://github.com/CosmWasm/wasmvm).


### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag         | Description                                     |
|-------------------|-------------------------------------------------|
| netgo             | Name resolution will use pure Go code           |
| ledger            | Ledger devices are supported (hardware wallets) |
| static_wasm/muslc | Used for statically linked wasmd dependencies   |


## Work with a Cosmos SDK Clone

To work with your own modifications of the Cosmos SDK, make a fork of this repo, and add a `replace` clause to the `go.mod` file.
The `replace` clause you add to `go.mod` must provide the correct import path.

```shell
go mod edit -replace github.com/cosmos/cosmos-sdk=../cosmos-sdk
go mod tidy
make install # or make build
```
## Next

Now you can [join the mainnet](../hub-tutorials/join-mainnet), [the public testnet](../hub-tutorials/join-testnet).
