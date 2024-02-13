# Gaia v15.0.0  Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v15.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v14.1.0...v15.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v15.x/UPGRADING.md) when migrating from `v14.x` to `v15.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release upgrades Cosmos SDK to v0.47 -- it uses [v0.47.8-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.8-ics-lsm), a special Cosmos SDK branch with support for both ICS and LSM. Consequently, it also upgrades IBC to v7 ([v7.3.1](https://github.com/cosmos/ibc-go/releases/tag/v7.3.1)) and Comet BFT to v0.37 ([v0.37.4](https://github.com/cometbft/cometbft/releases/tag/v0.37.4)).

## 🔨 Build from source

You must use Golang `v1.21` if building from source.

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v15.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.