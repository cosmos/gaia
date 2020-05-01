# Gaia
Gaia is the first implementation of the Cosmos Hub, built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).  Gaia and other Cosmos Hubs allow fully sovereign blockchains to interact with one another using a protocl called [IBC](https://github.com/cosmos/ics/tree/master/ibc) that enables Inter-Blockchain Communication.  

[![CircleCI](https://circleci.com/gh/cosmos/gaia/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/gaia/tree/master)
[![codecov](https://codecov.io/gh/cosmos/gaia/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/gaia)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/gaia)](https://goreportcard.com/report/github.com/cosmos/gaia)
[![license](https://img.shields.io/github/license/cosmos/gaia.svg)](https://github.com/cosmos/gaia/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/gaia)](https://github.com/cosmos/gaia)
[![GolangCI](https://golangci.com/badges/github.com/cosmos/gaia.svg)](https://golangci.com/r/github.com/cosmos/gaia)


## Mainnet Full Node Quick Start

This assumes that you're running Linux or MacOS and have installed [Go 1.14+](https://golang.org/dl/).  It will build and install Gaia, allow you to name your node, add seeds to your config file, start your node and use gaiacli to check the status of your node.  Welcome to the Cosmos!

```
git clone -b v2.0.9 https://github.com/cosmos/gaia
cd gaia
make install
gaiad init yournodenamehere
SEEDS=$(cat seeds); original_string="seeds = \"\""; replace_string="seeds = \"$SEEDS\""; sed -i -e "s/$original_string/$replace_string/g" "$HOME/.gaiad/config/config.toml"
curl https://raw.githubusercontent.com/cosmos/launch/master/genesis.json > $HOME/.gaiad/config/genesis.json
gaiad start
gaiacli status
```

## Talk to us!

* [Discord](https://discord.gg/huHEBUX)
* [Twitter](https://twitter.com/cosmos)
* [Telegram](https://t.me/cosmosproject)

## Archives

With each version of the Cosmos Hub, the chain is restarted from a new Genesis state.  We are currently on cosmoshub-3.  Archives of the state of cosmoshub-1 and cosmoshub-2 are available [here](./docs/resources/archives.md).

## Disambiguation

This Cosmos-SDK project is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. Per our agreement, this disambiguation notice will stay here.
