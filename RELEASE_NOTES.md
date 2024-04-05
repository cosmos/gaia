# Gaia v15.2.0  Release Notes 

***This is a special point release in the v15 release series.***

## 🕐  Timeline

**This is a mandatory upgrade for all validators and full node operators.**
The upgrade height is [19939000](https://www.mintscan.io/cosmos/block/19939000), which is approx. April 10th 2024, 15:00 CET.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v15.2.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v15.1.0...v15.2.0) from last release.

<!-- Add the following line for releases that require a coordinated upgrade -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v15.2.x/UPGRADING.md) when migrating from `v15.1.x` to `v15.2.x`.

## 🚀 Highlights

<!-- Add any highlights of this release --> 

This release fixes two issues identified after the v15 upgrade:

- Increases x/gov metadata fields length to 10200.
- Fixes parsing of historic Txs with TxExtensionOptions.

As both fixes are state breaking, a coordinated upgrade is necessary. 

## 🔨 Build from source

You must use Golang `v1.21` if building from source.

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v15.2.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.