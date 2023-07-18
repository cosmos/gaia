# Gaia v11.0.0  Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v11.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/cosmos-sdk/compare/release/v10.0.2...v11.0.0) from last release.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release contains the following changes to Gaia:

- Bumps Interchain-Security to v2.0.0.
- Refactors the Global Fee module.
- Removes all states of the Liquidity module with forced withdrawal for all pool 
  coins as a preparation to remove the module.

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v11.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.