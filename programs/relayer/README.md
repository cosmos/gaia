# POC Relayer Implementation

This is a proof of concept implementation of a relayer server for `solidity-ibc-eureka` and Cosmos SDK based chains.

This relayer works differently from other relayers in that it neither listens to events nor submits transactions to any chain. Instead, it runs a gRPC server that can be queried by a client to get the transactions that need to be submitted to the chain to relay packets.

The client submits the hashes of the transactions that need to be relayed to the relayer, and the relayer:
1. Queries the chain for the transactions with the given hashes.
2. Parses the transaction events to get the packet data.
3. Generates the corresponding IBC transactions and proof into a single transaction.
4. Does not sign nor submit the transaction to the chain, but returns it to the client.

In essence, this relayer is meant to be used in a setup where the client is a front-end application, or a service that can sign and submit transactions to the chain.

## Overview

The relayer is composed of multiple one-sided relayer servers, each of which is responsible for relaying packets from one chain to another. A relayer module is a rust struct that implements the [`RelayerModule`](https://github.com/cosmos/solidity-ibc-eureka/blob/debc0ad73acab0cd0a827a1a35a7ae4c1c65feb1/relayer/src/core/modules.rs#L10) trait.

You can see the protocol buffer definition for the gRPC service [here](https://github.com/cosmos/solidity-ibc-eureka/blob/debc0ad73acab0cd0a827a1a35a7ae4c1c65feb1/relayer/proto/relayer/relayer.proto).

This is a work-in-progress implementation, and the relayer is not yet usable. The relayer will only be able to relay IBC Eureka packets. There is a tracking issue for the relayer [here](https://github.com/cosmos/solidity-ibc-eureka/issues/121).

| **Source Chain** | **Target Chain** | **Light Client** | **Development Status** |
|:---:|:---:|:---:|:---:|
| Cosmos SDK | EVM | `sp1-ics07-tendermint` | ✅ |
| EVM | Cosmos SDK | `cw-ics08-wasm-eth` | ✅ |
| Cosmos SDK | Cosmos SDK | `07-tendermint` | ✅ |

## Usage

To run the relayer binary, you need to write a configuration file. At the moment, there is a working example configuration file at [`config.example.json`](./config.example.json). You can copy this file and modify it to suit your needs.

After building/installing the relayer binary, you can run the relayer with the following command:

```sh
relayer -c config.json
```
