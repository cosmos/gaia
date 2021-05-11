<!--
order: 2
-->

# Install Gaia

This guide will explain how to install the `gaiad` entrypoint
onto your system. With these installed on a server, you can participate in the
mainnet as either a [Full Node](./join-mainnet.md) or a
[Validator](../validators/validator-setup.md).

## Install build requirements

Install `make` and `gcc`.

On Ubuntu this can be done with the following:
```bash
sudo apt-get update

sudo apt-get install -y make gcc
```

## Install Go

Install `go` by following the [official docs](https://golang.org/doc/install).
Remember to set your `$PATH` environment variable, for example:

```bash
mkdir -p $HOME/go/bin
echo "export PATH=$PATH:$(go env GOPATH)/bin" >> ~/.bash_profile
source ~/.bash_profile
```

::: tip
**Go 1.16+** or later is required for the Cosmos SDK.
:::

## Install the binaries

Next, let's install the latest version of Gaia. Make sure you `git checkout` the
correct [released version](https://github.com/cosmos/gaia/releases).

```bash
git clone -b <latest-release-tag> https://github.com/cosmos/gaia
cd gaia && make install
```

If this command fails due to the following error message, you might have already set `LDFLAGS` prior to running this step.

```
# github.com/cosmos/gaia/cmd/gaiad
flag provided but not defined: -L
usage: link [options] main.o
...
make: *** [install] Error 2
```

Unset this environment variable and try again.

```
LDFLAGS="" make install
```

> _NOTE_: If you still have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `gaiad` binary. Verify that everything is OK:

```bash
gaiad version --long
```

`gaiad` for instance should output something similar to:

```bash
name: gaia
server_name: gaiad
version: v4.2.1
commit: dbd8a6fb522c571debf958837f9113c56d418f6b
build_tags: netgo,ledger
go: go version go1.15 darwin/amd64
```

### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag | Description                                     |
| --------- | ----------------------------------------------- |
| netgo     | Name resolution will use pure Go code           |
| ledger    | Ledger devices are supported (hardware wallets) |

## Work with a Cosmos SDK Clone

To work with your own modifications of the Cosmos SDK, make a fork of this repo, and add a `replace` clause to the `go.mod` file.
The `replace` clause you add to `go.mod` must provide the correct import path:

- Make appropriate changes
- Add `replace github.com/cosmos/cosmos-sdk => /path/to/clone/cosmos-sdk` to `go.mod`
- Run `make clean install` or `make clean build`
- Test changes

## Next

Now you can [join the mainnet](./join-mainnet.md), [the public testnet](./join-testnet.md) or [create you own testnet](./deploy-testnet.md)
