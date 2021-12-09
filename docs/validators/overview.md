<!--
order: 1
-->

# Validators Overview

## Introduction

The [Cosmos Hub](../README.md) is based on [Tendermint](https://github.com/tendermint/tendermint/tree/master/docs/introduction), which relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes that contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own ATOM and have ATOM ["delegated"](../delegators/delegator-guide-cli.md), or staked, to them by token holders. The Cosmos Hub has 125 validators, but over time the number of validators will increase to 300 according to a predefined schedule. The validators are determined by who has the most stake delegated to them — the top 125 validator candidates with the most stake are the current Cosmos validators.

Validators and their delegators earn ATOM as block provisions and tokens as transaction fees through execution of the Tendermint consensus protocol. Initially, transaction fees are paid in ATOM but in the future, any token in the Cosmos ecosystem will be valid as fee tender if the token is whitelisted by governance. Note that validators can set commission on the fees their delegators receive as additional incentive.

If validators double sign, are frequently offline, or do not participate in governance, their staked ATOM (including ATOM of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

For validator key management, validators must set up a physical operation that is secured with restricted access. A good starting place, for example, would be co-locating in secure data centers. 

Validators are expected to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall, and switching and then small servers with redundant hard drive and failover. Hardware can be on the low end of datacenter gear to start out with.

Initial network requirements can be low. The current testnet requires minimal resources. Then bandwidth, CPU, and memory requirements rise as the network grows. Large hard drives are recommended for storing years of blockchain history.

## Set Up a Website

Set up a dedicated validator's website and signal your intention to become a validator on the Cosmos Forum [Validator Candidates](https://forum.cosmos.network/t/validator-candidates-websites/127/3) site. Posting your validator website is important because delegators want to have information about the entity they are delegating their ATOM to.

## Seek Legal Advice

Seek legal advice if you intend to run a Validator.

## Community

Discuss the finer details of being a validator on our community chat and forum:

* [Validator Chat](https://riot.im/app/#/room/#cosmos_validators:matrix.org)
* [Validator Forum](https://forum.cosmos.network/c/validating)
