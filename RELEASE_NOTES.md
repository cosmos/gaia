<!--
  A release notes template that should be adapted for every release
    - release: <v*.*.*>
    - release branch: <v*.x>
    - the last release: <v-last> 
    - the last release branch: <v-last.x>
-->

# Gaia v25.3.0 Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v25.3.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v25.2.0...v25.3.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v25.3.0/UPGRADING.md) when migrating from `v25.2.0` to `v25.3.0`.

## 🚀 Highlights

This release bumps the following dependencies:

- Bump [cometbft](https://github.com/cometbft/cometbft) from v0.38.19 to [v0.38.20](https://github.com/cometbft/cometbft/releases/tag/v0.38.20)
- Bump [ibc-go](https://github.com/cosmos/ibc-go) from 10.3.0 to [10.5.0](https://github.com/cosmos/ibc-go/releases/tag/v10.5.0)
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) from 0.53.3 to [0.53.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.4)

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v25.3.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.