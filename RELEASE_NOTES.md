<!--
  A release notes template that should be adapted for every release
    - release: <v*.*.*>
    - release branch: <v*.x>
    - the last release: <v-last> 
    - the last release branch: <v-last.x>
-->

# Gaia v26.0.0-rc0 Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v26.0.0-rc0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v25.3.2...v26.0.0-rc0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v26.0.0-rc0/UPGRADING.md) when migrating from `v25.3.2` to `v26.0.0`.

## 🚀 Highlights

* This upgrade introduces the `x/tokenfactory` module to the Hub, as per signalling proposal [1010](https://www.mintscan.io/cosmos/proposals/1010).
  * The following parameters will be set during the upgrade and can be adjusted through future governance proposals:
    * Denom creation fee: `10 ATOM`
    * Denom creation gas consumption: `2_000_000`


## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v26.0.0-rc0
make install
```

## ⚡️ Download binaries

Binaries for linux and darwin are available below.