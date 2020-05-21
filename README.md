# Gaia
Gaia is the first implementation of the Cosmos Hub, built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).  Gaia and other Cosmos Hubs allow fully sovereign blockchains to interact with one another using a protocol called [IBC](https://github.com/cosmos/ics/tree/master/ibc) that enables Inter-Blockchain Communication.  

[![CircleCI](https://circleci.com/gh/cosmos/gaia/tree/master.svg?style=shield)](https://circleci.com/gh/cosmos/gaia/tree/master)
[![codecov](https://codecov.io/gh/cosmos/gaia/branch/master/graph/badge.svg)](https://codecov.io/gh/cosmos/gaia)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmos/gaia)](https://goreportcard.com/report/github.com/cosmos/gaia)
[![license](https://img.shields.io/github/license/cosmos/gaia.svg)](https://github.com/cosmos/gaia/blob/master/LICENSE)
[![LoC](https://tokei.rs/b1/github/cosmos/gaia)](https://github.com/cosmos/gaia)
[![GolangCI](https://golangci.com/badges/github.com/cosmos/gaia.svg)](https://golangci.com/r/github.com/cosmos/gaia)


## Mainnet Full Node Quick Start

This assumes that you're running Linux or MacOS and have installed [Go 1.14+](https://golang.org/dl/).  This guide helps you:

* build and install Gaia
* allow you to name your node
* add seeds to your config file
* download genesis state
* start your node 
* use gaiacli to check the status of your node.  

Build, Install, and Name your Node
```bash
# Clone Gaia
git clone -b v2.0.9 https://github.com/cosmos/gaia
# Enter the folder Gaia was cloned into
cd gaia
# Comile and install Gaia
make install
# Initalize Gaiad in ~/.gaiad and name your node
gaiad init yournodenamehere
```

Add Seeds
```bash
# Edit config.toml
nano ~/.gaiad/config/config.toml
```

Scroll down to seeds in `config.toml`, and add some of these seeds as a comma-separated list:
```
ba3bacc714817218562f743178228f23678b2873@5.83.160.108:26656
1e63e84945837b026f596ed8ae68708783d04ad4@51.75.145.123:26656
d2d452e7c9c43fa5ef017552688de60a5c0053ee@34.245.217.163:26656
dd36969b56c740bb40bb8badd4d4c6facc35dc24@206.189.115.41
a0aca8fb801c69653a290bd44872e8457f8b0982@47.99.180.54
27f8dd3bdbecbef7192291083706c156e523d8e0@3.122.248.21
aee0df1a660f301d456a0c2f805b372f7341e8ec@63.35.230.143:26656
7d1f660b361d6286715c098a3a171e554e9642bb@34.254.205.37
fa105c2291ac4aa452552fa4835266300a8209e1@88.198.41.62
bd410d4564f7e0dd9a0eb16a64c337a059e11b80@47.103.35.130
```

Download Genesis, Start your Node, Check your Node Status
```bash
# Download genesis.json
wget -O $HOME/.gaiad/config/genesis.json https://raw.githubusercontent.com/cosmos/launch/master/genesis.json 
# Start Gaiad
gaiad start
# Check your node's status with gaiacli
gaiacli status
```

Welcome to the Cosmos!

## Talk to us!

We have active, helpful communities on Twitter, Discord, and Telegram.

* [Discord](https://discord.gg/huHEBUX)
* [Twitter](https://twitter.com/cosmos)
* [Telegram](https://t.me/cosmosproject)

## Archives

With each version of the Cosmos Hub, the chain is restarted from a new Genesis state.  We are currently on cosmoshub-3.  Archives of the state of cosmoshub-1 and cosmoshub-2 are available [here](./docs/resources/archives.md).

## Disambiguation

Gaia is not related to the [React-Cosmos](https://github.com/react-cosmos/react-cosmos) project (yet). Many thanks to Evan Coury and Ovidiu (@skidding) for this Github organization name. Per our agreement, this disambiguation notice will stay here.
