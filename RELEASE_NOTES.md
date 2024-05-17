# Gaia v17.0.0  Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v17.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v16.0.0...v17.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v17.x/UPGRADING.md) when migrating from `v16.x` to `v17.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->
This release adds ICS 2.0 to Gaia. ICS 2.0 -- also known as Partial Set Security (PSS) -- allows each consumer chain to leverage only a subset of the Hub validator set and enables validators to opt-in to validate the consumer chains they want. For more details, check out the [ICS docs](https://cosmos.github.io/interchain-security/features/partial-set-security).   

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v17.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.