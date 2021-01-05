# Gaia

Gaia 3.0 is a test version of the Stargate binaries. It is intended to use in testing compatibility with the post Stargate upgrade.
The biggest change is that there is no longer two separate binaries. There is just `gaiad`.
The rest and new gRPC interfaces can be configured using the `app.toml`.
You can interact via the cli interface using a second instance of the binary while a full node is running.
Key an eye on our [audit](https://github.com/cosmosdevs/stargate/pull/8) of interface of changes to help with upgrading.

Gaia is the first implementation of the Cosmos Hub, built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).  Gaia and other Cosmos Hubs allow fully sovereign blockchains to interact with one another using a protocol called [IBC](https://github.com/cosmos/ics/tree/master/ibc) that enables Inter-Blockchain Communication.

[![CircleCI](https://circleci.com/gh/cosmos/gaia/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/gaia/tree/master)
[![codecov](https://codecov.io/gh/cosmos/gaia/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/gaia)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/gaia)](https://goreportcard.com/report/github.com/cosmos/gaia)
[![license](https://img.shields.io/github/license/cosmos/gaia.svg)](https://github.com/cosmos/gaia/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/gaia)](https://github.com/cosmos/gaia)
[![GolangCI](https://golangci.com/badges/github.com/cosmos/gaia.svg)](https://golangci.com/r/github.com/cosmos/gaia)

## Talk to us!

We have active, helpful communities on Twitter, Discord, and Telegram.

* [Discord](https://discord.gg/huHEBUX)
* [Twitter](https://twitter.com/cosmos)
* [Telegram](https://t.me/cosmosproject)

## Archives

With each version of the Cosmos Hub, the chain is restarted from a new Genesis state.  We are currently on cosmoshub-3.  Archives of the state of cosmoshub-1 and cosmoshub-2 are available [here](./docs/resources/archives.md).

Gaia is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. Per our agreement, this disambiguation notice will stay here.
