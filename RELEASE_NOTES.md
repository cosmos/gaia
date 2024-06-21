# Gaia v18.0.0  Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v18.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v17.2.0...v18.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v18.x/UPGRADING.md) when migrating from `v17.2.x` to `v18.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release adds the following features:

- Permissioned CosmWasm (as per [prop 895](https://www.mintscan.io/cosmos/proposals/895)) enables the governance-gated deployment of CosmWasm contracts. See [this forum discussion](https://forum.cosmos.network/t/discussion-hub-cosmwasm-guidelines/13788) for more details on what contracts should be deployed on the Hub.
- Skip's [feemarket module](https://github.com/skip-mev/feemarket) (as per [prop 842](https://www.mintscan.io/cosmos/proposals/842)) enables the dynamic adjustment of the base transaction fee based on the block utilization (the more transactions in a block, the higher the base fee). This module replaces the [x/globalfee module](https://hub.cosmos.network/v17.1.0/architecture/adr/adr-002-globalfee).  
- [Expedited proposals](https://docs.cosmos.network/v0.50/build/modules/gov#expedited-proposals) (as per [prop 926](https://www.mintscan.io/cosmos/proposals/926)) enable governance proposals with a shorter voting period (i.e., one week instead of two), but with a higher tally threshold (i.e., 66.7% of Yes votes for the proposal to pass) and a higher minimum deposit (i.e., 500 ATOMs instead of the 250 for regular proposals). Initially, only `MsgSoftwareUpgrade` and `MsgCancelUpgrade` can be expedited.

The release also bumps the following dependencies:
- Golang to v1.22
- Cosmos SDK to [v0.47.16-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.16-ics-lsm)
- IBC to [v7.6.0](https://github.com/cosmos/ibc-go/releases/tag/v7.6.0)
- ICS to [v4.3.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.0-lsm)

## ❤️ Contributors

* Binary Builders ([@binary_builders](https://x.com/binary_builders))
* Informal Systems ([@informalinc](https://x.com/informalinc))
* Hypha Worker Co-operative ([@HyphaCoop](https://x.com/HyphaCoop))
* Skip Protocol ([@skipprotocol](https://x.com/skipprotocol))
* Strangelove ([@strangelovelabs](https://x.com/strangelovelabs))

This list is non-exhaustive and ordered alphabetically.
Thank you to everyone who contributed to this release!

## 🔨 Build from source

❗***You must use Golang v1.22 if building from source.***

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v18.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.