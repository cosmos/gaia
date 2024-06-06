---
title: What is Gaia?
sidebar_position: 1
---

The Cosmos Hub is a public Proof-of-Stake chain that uses ATOM as its native staking token. It is the first blockchain launched in the Cosmos Network and developed using the [cosmos-sdk](https://docs.cosmos.network/) development framework and [ibc-go](https://ibc.cosmos.network/).

Cosmos hub is also the first security aggregation platform that leverages the [interchain-security](https://cosmos.github.io/interchain-security/) protocol ([ICS-28](https://github.com/cosmos/ibc/tree/main/spec/app/ics-028-cross-chain-validation)) to facilitate the launch of cosmos-sdk blockchain projects.


:::tip
Interchain security features deployed on the Cosmos Hub blockchain allow anyone to launch a blockchain using a subset, or even the entire validator set of the Cosmos Hub blockchain. 
:::


:::info
* `gaia` is the name of the Cosmos SDK application for the Cosmos Hub.

* `gaiad` is the daemon and command-line interface (CLI) that operates the `gaia` blockchain application.
:::


The `gaia` blockchain uses the following cosmos-sdk, ibc-go and interchain-security modules, alongside some others:

## cosmos-sdk
* [x/auth](https://docs.cosmos.network/v0.47/build/modules/auth)
* [x/authz](https://docs.cosmos.network/v0.47/build/modules/authz)
* [x/bank](https://docs.cosmos.network/v0.47/build/modules/bank)
* [x/capability](https://docs.cosmos.network/v0.47/build/modules/capability)
* [x/consensus](https://docs.cosmos.network/v0.47/build/modules/consensus)
* [x/crisis](https://docs.cosmos.network/v0.47/build/modules/crisis)
* [x/distribution](https://docs.cosmos.network/v0.47/build/modules/distribution)
* [x/evidence](https://docs.cosmos.network/v0.47/build/modules/evidence)
* [x/feegrant](https://docs.cosmos.network/v0.47/build/modules/feegrant)
* [x/gov](https://docs.cosmos.network/v0.47/build/modules/gov)
* [x/mint](https://docs.cosmos.network/v0.47/build/modules/mint)
* [x/params](https://docs.cosmos.network/v0.47/build/modules/params)
* [x/slashing](https://docs.cosmos.network/v0.47/build/modules/slashing)
* [x/staking (with LSM changes)](https://docs.cosmos.network/v0.47/build/modules/staking)
* [x/upgrade](https://docs.cosmos.network/v0.47/build/modules/upgrade)

## ibc-go
* [transfer](https://ibc.cosmos.network/main/apps/transfer/overview)
* [interchain accounts - host](https://ibc.cosmos.network/v8/apps/interchain-accounts/client#host)
* [interchain accounts - controller](https://ibc.cosmos.network/v8/apps/interchain-accounts/client#controller)
* [interchain-security/provider](https://github.com/cosmos/interchain-security/tree/main/x/ccv/provider)
* [packetforward](https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware)
* [ibcfee](https://ibc.cosmos.network/v7/middleware/ics29-fee/overview)
* [ibc-rate-limiting](https://github.com/Stride-Labs/ibc-rate-limiting)

## gaia specific modules
* [x/metaprotocols](https://github.com/cosmos/gaia/tree/main/x/metaprotocols)

## other modules
* [fee market](https://github.com/skip-mev/feemarket)

Next, learn how to [install Gaia](./installation.md).
