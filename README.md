# Agoric Gaia

This repository contains sources for the [Agoric blockchain's](https://agoric.com/) `agaiad` program, a fork of the [Cosmos Gaia implementation](https://github.com/cosmos/gaia).

TL;DR: Compile `agaiad` with `make build`, and run with `build/agaiad`.

The first Agoric mainnet (phase 0) will not have the [Agoric SDK](https://github.com/Agoric/agoric-sdk) enabled until governance votes to turn it on.  Until then,
validators run `agaiad` to bootstrap the chain with support for Cosmos-layer
validation, staking, and governance.

Please refer to https://agoric.com to learn about Agoric and get involved.

*The rest of the original Gaia README follows:*

----

# Original Gaia README
Gaia is the first implementation of the Cosmos Hub, built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).  Gaia and other Cosmos Hubs allow fully sovereign blockchains to interact with one another using a protocol called [IBC](https://github.com/cosmos/ics/tree/master/ibc) that enables Inter-Blockchain Communication.

[![codecov](https://codecov.io/gh/cosmos/gaia/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/gaia)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/gaia)](https://goreportcard.com/report/github.com/cosmos/gaia)
[![license](https://img.shields.io/github/license/cosmos/gaia.svg)](https://github.com/cosmos/gaia/blob/main/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/gaia)](https://github.com/cosmos/gaia)
[![GolangCI](https://golangci.com/badges/github.com/cosmos/gaia.svg)](https://golangci.com/r/github.com/cosmos/gaia)

## Documentation

Documentation for the Cosmos Hub lives at [hub.cosmos.network](https://hub.cosmos.network/main/hub-overview/overview.html).

## Talk to us!

We have active, helpful communities on Twitter, Discord, and Telegram.

* [Discord](https://discord.gg/vcExX9T)
* [Twitter](https://twitter.com/cosmos)
* [Telegram](https://t.me/cosmosproject)

## Archives & Genesis

With each version of the Cosmos Hub, the chain is restarted from a new Genesis state. 
Mainnet is currently running as `cosmoshub-4`. Archives of the state of `cosmoshub-1`, `cosmoshub-2`, and `cosmoshub-3` are available [here](./docs/resources/archives.md).

If you are looking for historical genesis files and other data [`cosmos/mainnet`](http://github.com/cosmos/mainnet) is an excellent resource.
