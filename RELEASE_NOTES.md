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

- IBC rate limit

  IBC rate limit prevents massive inflows or outflows of IBC tokens in a short time frame to add an extra layer of protection on IBC transfers

- ICA controller

  With ICA controller the Cosmos Hub expands its functionality to become a controller chain allowing controlling the accounts on another host chain

- IBC fee middleware

  Allows transfer packet relaying incentives

- ICS epochs

  Reduces the amount of Interchain Security protocol IBC packets to reduce relaying costs

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