<!--
  A release notes template that should be adapted for every release
    - release: <v*.*.*>
    - release branch: <v*.x>
    - the last release: <v-last>
    - the last release branch: <v-last.x>
-->

# Gaia v16.0.0  Release Notes

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v16.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v15.2.0...v16.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v16.x/UPGRADING.md) when migrating from `v15.x` to `v16.x`.

## 🚀 Highlights

This releases adds several features made possible by the upgrade to Cosmos SDK v0.47:

- The [IBC rate limit module](https://github.com/Stride-Labs/ibc-rate-limiting) prevents massive inflows or outflows of IBC tokens in a short time frame to add an extra layer of protection on IBC transfers.
- The [ICA controller sub-module](https://ibc.cosmos.network/v7/apps/interchain-accounts/overview) enables Hub users to perform actions on other chains using their Hub accounts. 
- The [IBC fee middleware](https://ibc.cosmos.network/v7/middleware/ics29-fee/overview) enables creating IBC channels with in-protocol incentivization for relayers. 

This release also bumps ICS to [v4.1.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.0-lsm), which introduces [ICS epochs](https://cosmos.github.io/interchain-security/adrs/adr-014-epochs) to reduce the relaying cost for ICS. 

<!-- Add any highlights of this release -->

## ❤️ Contributors
* Binary Builders ([@binary_builders](https://twitter.com/binary_builders))
* Informal Systems ([@informalinc](https://twitter.com/informalinc))
* Hypha Worker Co-operative ([@HyphaCoop](https://twitter.com/HyphaCoop))
* Stride ([@stride_zone](https://twitter.com/stride_zone))

This list is non-exhaustive and ordered alphabetically.
Thank you to everyone who contributed to this release!

## 🔨 Build from source

You must use Golang v1.21 if building from source.

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v16.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.