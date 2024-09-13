# Gaia v20.0.0 Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v20.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/<v-last>...v20.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v20.x/UPGRADING.md) when migrating from `v19.2.x` to `v20.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release bumps Interchain Security (ICS) to [v6.0.0](https://github.com/cosmos/interchain-security/releases/tag/v6.0.0) which brings the following major features:

- ICS with Inactive Validators (as per [prop 930](https://www.mintscan.io/cosmos/proposals/930)) enables validators from outside the Hub’s active set to validate on Consumer Chains.
- Permissionless ICS (as per [prop 945](https://www.mintscan.io/cosmos/proposals/945)) enables users to permissionlessly launch opt-in Consumer Chains on the Cosmos Hub.

It also bumps CosmWasm/wasmd to [v0.53.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.0) and ibc-go to [v8.5.1](https://github.com/cosmos/ibc-go/releases/tag/v8.5.1).

## 🔨 Build from source

❗***You must use Golang v1.22 if building from source.***

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v20.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.