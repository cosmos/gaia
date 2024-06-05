# Gaia v17.2.0  Release Notes 

⚠️ ***This is a special point release in the v17 release series. This release fixes a bug that resulted in the halt of the Cosmos Hub at height 20740970.***

## 🕐  Timeline
**This is a mandatory upgrade for all validators and full node operators.**
The upgrade height is [20740970](https://www.mintscan.io/cosmos/block/20740970).

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v17.2.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v17.1.0...v17.2.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v17.2.x/UPGRADING.md) when migrating from `v17.1.x` to `v17.2.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->
This release fixes a bug that resulted in the halt of the Cosmos Hub at height 20740970. The fix removes a panic from the source code. 

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v17.2.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.