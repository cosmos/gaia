# CHANGELOG

## v20.0.0

*September 13, 2024*

### API BREAKING

- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.1.0](https://github.com/cosmos/interchain-security/releases/tag/v6.1.0).
  This release of ICS introduces several API breaking changes. 
  See the [ICS changelog](https://github.com/cosmos/interchain-security/blob/v6.1.0/CHANGELOG.md#api-breaking) for details.
  ([\#3350](https://github.com/cosmos/gaia/pull/3350))

### BUG FIXES

- Migrate consensus params - initialize Version field
  ([\#3333](https://github.com/cosmos/gaia/pull/3333))

### DEPENDENCIES

- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
  [v0.53.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.0)
  ([\#3304](https://github.com/cosmos/gaia/pull/3304))
- Bump [feemarket](https://github.com/skip-mev/feemarket) to
  [v1.1.1](https://github.com/skip-mev/feemarket/releases/tag/v1.1.1)
  ([3306](https://github.com/cosmos/gaia/pull/3306))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.5.1](https://github.com/cosmos/ibc-go/releases/tag/v8.5.1)
    ([\#3338](https://github.com/cosmos/gaia/pull/3338))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.1.0](https://github.com/cosmos/interchain-security/releases/tag/v6.1.0).
  ([\#3350](https://github.com/cosmos/gaia/pull/3350))

### FEATURES

- Set the `MaxProviderConsensusValidators` parameter of the provider module to 180. 
  This parameter will be used to govern the number of validators participating in consensus,
  and takes over this role from the `MaxValidators` parameter of the staking module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))
- Set the `MaxValidators` parameter of the staking module to 200, which is the current number of 180 plus 20.
  This is done as a result of introducing the inactive-validators feature of Interchain Security, 
  which entails that the number of validators participating in consensus will be governed by the 
  `MaxProviderConsensusValidators` parameter in the provider module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))
- Set the metadata for launched ICS consumer chains.
  ([\#3308](https://github.com/cosmos/gaia/pull/3308))
- Migrate active ICS gov proposal to the new messages
  introduced by the permissionless ICS feature.
  ([\#3316](https://github.com/cosmos/gaia/pull/3316))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.1.0](https://github.com/cosmos/interchain-security/releases/tag/v6.1.0).
  This release of ICS enables the permissionless creation of consumer chains 
  and allows validators outside the active validator set to opt in to validate 
  on consumer chains.
  ([\#3350](https://github.com/cosmos/gaia/pull/3350))

### STATE BREAKING

- Set the `MaxProviderConsensusValidators` parameter of the provider module to 180. 
  This parameter will be used to govern the number of validators participating in consensus,
  and takes over this role from the `MaxValidators` parameter of the staking module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))
- Set the `MaxValidators` parameter of the staking module to 200, which is the current number of 180 plus 20.
  This is done as a result of introducing the inactive-validators feature of Interchain Security, 
  which entails that the number of validators participating in consensus will be governed by the 
  `MaxProviderConsensusValidators` parameter in the provider module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))
- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
  [v0.53.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.0)
  ([\#3304](https://github.com/cosmos/gaia/pull/3304))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.5.0](https://github.com/cosmos/ibc-go/releases/tag/v8.5.0)
    ([\#3305](https://github.com/cosmos/gaia/pull/3305))
- Set the metadata for launched ICS consumer chains.
  ([\#3308](https://github.com/cosmos/gaia/pull/3308))
- Migrate active ICS gov proposal to the new messages
  introduced by the permissionless ICS feature.
  ([\#3316](https://github.com/cosmos/gaia/pull/3316))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.1.0](https://github.com/cosmos/interchain-security/releases/tag/v6.1.0).
  ([\#3350](https://github.com/cosmos/gaia/pull/3350))

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

