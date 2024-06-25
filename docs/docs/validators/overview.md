---
title: Validator Overview
order: 1
---

## Introduction

The Cosmos Hub is based on [CometBFT](https://docs.cometbft.com/v0.37/introduction/what-is-cometbft) that relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes that contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own ATOM and have ATOM ["delegated"](../delegators/delegator-guide-cli.md), or staked, to them by token holders. The Cosmos Hub has 180 active validators, but over time the number of validators can be changed through governance (`MaxValidators` parameter). Validator voting power is determined by the total number of ATOM tokens delegated to them. Validators that do not have enough voting power to be in the top 180 are considered inactive. Inactive validators can become active if their staked amount increases so that they fall into the top 180 validators.

Validators and their delegators earn ATOM as block provisions and tokens as transaction fees through execution of the Tendermint consensus protocol. Note that validators can set a commission percentage on the fees their delegators receive as additional incentive. You can find an overview of all current validators and their voting power on [Mintscan](https://www.mintscan.io/cosmos/validators).

If validators double sign or are offline for an [extended period](./validator-faq.md#what-are-the-slashing-conditions), their staked ATOM (including ATOM of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

For validator key management, validators must set up a physical operation that is secured with restricted access. A good starting place, for example, would be co-locating in secure data centers.

Validators are expected to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall, and switching and then small servers with redundant hard drive and failover.

You can find the minimum hardware requirements on the instructions for [joining the Cosmos Hub mainnet](../hub-tutorials/join-mainnet.md). As the network grows, bandwidth, CPU, and memory requirements rise. Large hard drives are recommended for storing years of blockchain history, as well as significant RAM to process the increasing amount of transactions.

## Create a Validator Website

To get started as a validator, create your dedicated validator website and signal your intention to become a validator in the [Interchain Discord](https://discord.gg/interchain). Posting your validator website is essential because delegators want to have information about the entity they are delegating their ATOM to.

## Seek Legal Advice

As always, do your own research and seek legal advice if you intend to run a validator node.

## Community

Discuss the finer details of being a validator on our community Discord and sign up for the Cosmos newsletter to get regular updates:

* [Discord](https://discord.gg/interchain)
* [Newsletter](https://cosmos.network/updates/signup/)
