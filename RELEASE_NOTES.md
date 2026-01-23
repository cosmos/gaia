<!--
  A release notes template that should be adapted for every release
    - release: <v*.*.*>
    - release branch: <v*.x>
    - the last release: <v-last> 
    - the last release branch: <v-last.x>
-->

# Gaia v25.3.2 Release Notes 



## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v25.3.2/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v25.3.0...v25.3.2) from last release.

<!-- Add the following line for major releases -->
<!--Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v25.3.0/UPGRADING.md) when migrating from `v25.2.0` to `v25.3.0`.-->

## 🚀 Highlights

This release bumps the following dependencies:

- Bump [cometbft](https://github.com/cometbft/cometbft) from v0.38.20 to [v0.38.21](https://github.com/cometbft/cometbft/releases/tag/v0.38.21) to address critical security vulnerability in CometBFT detailed [here](https://github.com/cometbft/cometbft/security/advisories/GHSA-c32p-wcqj-j677)

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v25.3.2
make install
```

## ⚡️ Download binaries

Binaries for linux and darwin are available below.
