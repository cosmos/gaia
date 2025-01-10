# Gaia v22.0.0 Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v22.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v21.0.1...v22.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v22.x/UPGRADING.md) when migrating from `v21.x` to `v22.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release bumps Interchain Security (ICS) to [v6.4.0](https://github.com/cosmos/interchain-security/releases/tag/v6.4.0) which brings the following improvements to the provider module:

- Add a priority list to the power shaping parameters. This enables consumer chains to further customize their validator set by setting a list of validators that will have priority regardless of their voting power. 
- Remove the governance proposal whitelisting from consumer chains. See [this discussion](https://github.com/cosmos/interchain-security/issues/1194) for details on the motivation for this change. 
- Enable the chain ID of a consumer chain to be updated after creation, but before launch. As a result, it brings more flexibility to consumer chain deployment, especially to chains that are not yet sure about their final chain ID. 
- Enable querying the provider module for the genesis time for a consumer chain. As a result, consumer chains can set an accurate genesis time in the genesis file when bootstrapping their chain. See [this issue](https://github.com/cosmos/interchain-security/issues/2280) for more details. 
- Enable consumer chains to have customizable slashing and jailing (as per [ADR 020](https://cosmos.github.io/interchain-security/adrs/adr-020-cutomizable_slashing_and_jailing)). As a result, consumer chain owners will be able to customize both the jailing period and slashing factor for different types of infractions. Note that the default params will maintain the current behavior, especially the no slashing for downtime.
- Simplify the changeover from standalone to consumer chains. Basically, this feature enables consumer chains to reused the existing IBC connection to the provider (the one created as standalone chains). This simplifies the process as new clients no longer need to be created, which means the initial height initialization parameter is no longer needed. 

This release also bumps the following dependencies:

- Cosmos SDK to [v0.50.11-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.11-lsm). This includes the changes introduced by Cosmos SDK [v0.50.10](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.10) and [v0.50.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.11), as well as fixes for issues identified by OtterSec and Zellic during their audit of LSM. Note that none of the identified issues are critical.  
- CometBFT to [v0.38.15](https://github.com/cometbft/cometbft/releases/tag/v0.38.15).
- IBC to [v8.5.2](https://github.com/cosmos/ibc-go/releases/tag/v8.5.2).
- wasmd to [v0.53.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.2).

## 🔨 Build from source

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v22.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.