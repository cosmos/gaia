# CHANGELOG

## UNRELEASED

### DEPENDENCIES

- Bump Go to 1.23 [\#3556](https://github.com/cosmos/gaia/pull/3556)
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v10.0.0-beta.2](https://github.com/cosmos/ibc-go/tree/v10.0.0-beta.2) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [ibc-apps/modules/rate-limiting](https://github.com/cosmos/ibc-apps/tree/main/modules/rate-limiting) to [v10.0.0-beta.2](https://github.com/cosmos/ibc-apps/tree/modules/rate-limiting/v10.0.0-beta.2) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [ibc-apps/middleware/packet-forward-middleware](https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware) to [v10.0.0-beta.2](https://github.com/cosmos/ibc-apps/tree/middleware/packet-forward-middleware/v10.0.0-beta.2) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [ibc-go/modules/apps/callbacks](https://github.com/cosmos/ibc-go/tree/main/modules/apps/callbacks) to [v0.3.0+ibc-go-v10.0-beta.2](https://github.com/cosmos/ibc-go/tree/modules/apps/callbacks/v0.3.0%2Bibc-go-v10.0-beta.2/modules/apps/callbacks) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [ibc-go/modules/light-clients/08-wasm](https://github.com/cosmos/ibc-go/tree/main/modules/light-clients/08-wasm) to [v0.6.0+ibc-go-v10.0-wasmvm-v2.2-beta.2](https://github.com/cosmos/ibc-go/tree/modules/light-clients/08-wasm/v0.6.0%2Bibc-go-v10.0-wasmvm-v2.2-beta.2) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [interchain-security](https://github.com/cosmos/interchain-security) to [v7.0.0-rc0](https://github.com/cosmos/interchain-security/tree/v7.0.0-rc0) ([\#3560](https://github.com/cosmos/gaia/pull/3560))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.50.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.12) ([\#3560](https://github.com/cosmos/gaia/pull/3560))

### BUG FIXES
- Export only validators that are participating in consensus
  ([\#3490](https://github.com/cosmos/gaia/pull/3490))
- Fix goreleaser config to generate Linux builds again. ([\#3506](https://github.com/cosmos/gaia/pull/3506))

### IMPROVEMENTS

### FEATURES
- Add [ibc-go/modules/light-clients/08-wasm](https://github.com/cosmos/ibc-go/tree/main/modules/light-clients/08-wasm) ([\#3554](https://github.com/cosmos/gaia/pull/3554))

### STATE BREAKING
- Remove [ibc-go/modules/apps/29-fee](https://github.com/cosmos/ibc-go/tree/v8.5.3/modules/apps/29-fee) ([\#3553](https://github.com/cosmos/gaia/pull/3553))

### API-Breaking

## v22.2.0

*February 12, 2025*

### DEPENDENCIES
- Bump [ibc-apps/middleware/packet-forward-middleware](https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware) to
    [v8.1.1](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv8.1.1)
    ([\#3534](https://github.com/cosmos/gaia/pull/3534))
- Add `v22.2.0` upgrade handler ([\#3538](https://github.com/cosmos/gaia/pull/3538))

### STATE BREAKING
- Bump [ibc-apps/middleware/packet-forward-middleware](https://github.com/cosmos/ibc-apps/tree/main/middleware/packet-forward-middleware) to
    [v8.1.1](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv8.1.1)
    ([\#3534](https://github.com/cosmos/gaia/pull/3534))
- Add `v22.2.0` upgrade handler ([\#3538](https://github.com/cosmos/gaia/pull/3538))

## v22.1.0

February 10, 2025

### DEPENDENCIES
- Bump [wasmvm](https://github.com/CosmWasm/wasmvm) to
  [v2.1.5](https://github.com/CosmWasm/wasmvm/releases/tag/v2.1.5)
  ([\#3519](https://github.com/cosmos/gaia/pull/3519))
- Bump [cometbft](https://github.com/cometbft/cometbft) to
  [v0.38.17](https://github.com/cometbft/cometbft/releases/tag/v0.38.17)
  ([\#3523](https://github.com/cosmos/gaia/pull/3523))

### STATE BREAKING
- Bump [wasmvm](https://github.com/CosmWasm/wasmvm) to
  [v2.1.5](https://github.com/CosmWasm/wasmvm/releases/tag/v2.1.5)
  ([\#3519](https://github.com/cosmos/gaia/pull/3519))

## v22.0.0

*January 10, 2025*

### DEPENDENCIES

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.5.2](https://github.com/cosmos/ibc-go/releases/tag/v8.5.2)
    ([\#3370](https://github.com/cosmos/gaia/pull/3370))
- Bump [cometbft](https://github.com/cometbft/cometbft) to
  [v0.38.15](https://github.com/cometbft/cometbft/releases/tag/v0.38.15)
  ([\#3370](https://github.com/cosmos/gaia/pull/3370))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.50.11-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.11-lsm)
  ([\#3454](https://github.com/cosmos/gaia/pull/3454))
- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
  [v0.53.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.2)
  ([\#3459](https://github.com/cosmos/gaia/pull/3459))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.4.0](https://github.com/cosmos/interchain-security/releases/tag/v6.4.0).
  ([\#3474](https://github.com/cosmos/gaia/pull/3474))

### STATE BREAKING

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.5.2](https://github.com/cosmos/ibc-go/releases/tag/v8.5.2)
    ([\#3370](https://github.com/cosmos/gaia/pull/3370))
- Bump [cometbft](https://github.com/cometbft/cometbft) to
  [v0.38.15](https://github.com/cometbft/cometbft/releases/tag/v0.38.15)
  ([\#3370](https://github.com/cosmos/gaia/pull/3370))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.50.11-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.11-lsm)
  ([\#3454](https://github.com/cosmos/gaia/pull/3454))
- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
  [v0.53.2](https://github.com/CosmWasm/wasmd/releases/tag/v0.53.2)
  ([\#3459](https://github.com/cosmos/gaia/pull/3459))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.4.0](https://github.com/cosmos/interchain-security/releases/tag/v6.4.0).
  ([\#3474](https://github.com/cosmos/gaia/pull/3474))

## v21.0.1

*November 21, 2024*

### BUG FIXES

- Bump [cosmossdk.io/math](https://github.com/cosmos/cosmos-sdk/tree/main/math) to
  [v1.4.0](https://github.com/cosmos/cosmos-sdk/tree/math/v1.4.0) in order to 
  address the [ASA-2024-010](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-7225-m954-23v7) 
  security advisory. 
  ([\#3418](https://github.com/cosmos/gaia/pull/3418))

### DEPENDENCIES

- Bump [cosmossdk.io/math](https://github.com/cosmos/cosmos-sdk/tree/main/math) to
  [v1.4.0](https://github.com/cosmos/cosmos-sdk/tree/math/v1.4.0)
  ([\#3418](https://github.com/cosmos/gaia/pull/3418))

## v21.0.0

*October 14, 2024*

### BUG FIXES

- Fix submission of broken invariants
  ([\#3346](https://github.com/cosmos/gaia/pull/3346))
- Fix creation of multiple temp directories
  ([\#3349](https://github.com/cosmos/gaia/pull/3349))
- Initialize uninitialized governance params
  ([\#3387](https://github.com/cosmos/gaia/pull/3387))

### DEPENDENCIES

- Update wasmvm to v2.1.3 - security patch
  ([\#3366](https://github.com/cosmos/gaia/pull/3366))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.3.0](https://github.com/cosmos/interchain-security/releases/tag/v6.3.0).
  ([\#3395](https://github.com/cosmos/gaia/pull/3395))

### STATE BREAKING

- Distribute all the unaccounted known denoms from  the
 consumer rewards pool.
([\#3361](https://github.com/cosmos/gaia/pull/3361))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v6.3.0](https://github.com/cosmos/interchain-security/releases/tag/v6.3.0).
  ([\#3395](https://github.com/cosmos/gaia/pull/3395))

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

## v19.2.0

*September 04, 2024*

### DEPENDENCIES

- Bump [ICS](https://github.com/cosmos/interchain-security) to
    [v5.2.0](https://github.com/cosmos/interchain-security/releases/tag/v5.2.0)
    ([\#3310](https://github.com/cosmos/gaia/pull/3310))

### STATE BREAKING

- Bump [ICS](https://github.com/cosmos/interchain-security) to
    [v5.2.0](https://github.com/cosmos/interchain-security/releases/tag/v5.2.0)
    ([\#3310](https://github.com/cosmos/gaia/pull/3310))

## v19.1.0

*August 21, 2024*

### BUG FIXES

- Bump [feemarket](https://github.com/skip-mev/feemarket) to
  [v1.1.0](https://github.com/skip-mev/feemarket/releases/tag/v1.1.0)
  ([92a2a88](https://github.com/cosmos/gaia/commit/92a2a88da512a1d8102817c61bd23cd65dda93c8))

### DEPENDENCIES

- Bump [feemarket](https://github.com/skip-mev/feemarket) to
  [v1.1.0](https://github.com/skip-mev/feemarket/releases/tag/v1.1.0)
  ([92a2a88](https://github.com/cosmos/gaia/commit/92a2a88da512a1d8102817c61bd23cd65dda93c8))
- Bump [cometbft](https://github.com/cometbft/cometbft) to
   [v0.38.11](https://github.com/cometbft/cometbft/releases/tag/v0.38.11)
   ([\#3270](https://github.com/cosmos/gaia/pull/3270))

## v19.0.0

*August 1st, 2024*

### DEPENDENCIES

- Bump [cometbft](https://github.com/cometbft/cometbft) to
    [v0.38.9](https://github.com/cometbft/cometbft/releases/tag/v0.38.9)
    ([\#3171](https://github.com/cosmos/gaia/pull/3171))
- Bump [feemarket](https://github.com/skip-mev/feemarket) to
    [v1.0.4](https://github.com/skip-mev/feemarket/releases/tag/v1.0.4)
    ([\#3221](https://github.com/cosmos/gaia/pull/3221))
- Bump [ibc-rate-limiting](https://github.com/cosmos/ibc-apps/blob/main/modules/rate-limiting) to
    [v8](https://github.com/cosmos/ibc-apps/releases/tag/modules/rate-limiting/v8.0.0)
    ([\#3227](https://github.com/cosmos/gaia/pull/3227))
- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
    [v0.51.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.51.0)
    ([\#3230](https://github.com/cosmos/gaia/pull/3230))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.4.0](https://github.com/cosmos/ibc-go/releases/tag/v8.4.0)
    ([\#3233](https://github.com/cosmos/gaia/pull/3233))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
    [v5.1.1](https://github.com/cosmos/interchain-security/releases/tag/v5.1.1)
    ([\#3237](https://github.com/cosmos/gaia/pull/3237))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
    [v0.50.9-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.9-lsm)
    ([\#3249](https://github.com/cosmos/gaia/pull/3249))

### IMPROVEMENTS

- Reintroduce docker builds for gaia and make them compatible with
  interchaintest ([\#3199](https://github.com/cosmos/gaia/pull/3199))

### STATE BREAKING

- Bump [cometbft](https://github.com/cometbft/cometbft) to
    [v0.38.9](https://github.com/cometbft/cometbft/releases/tag/v0.38.9)
    ([\#3171](https://github.com/cosmos/gaia/pull/3171))
- Bump [feemarket](https://github.com/skip-mev/feemarket) to
    [v1.0.4](https://github.com/skip-mev/feemarket/releases/tag/v1.0.4)
    ([\#3221](https://github.com/cosmos/gaia/pull/3221))
- Bump [ibc-rate-limiting](https://github.com/cosmos/ibc-apps/blob/main/modules/rate-limiting) to
    [v8](https://github.com/cosmos/ibc-apps/releases/tag/modules/rate-limiting/v8.0.0)
    ([\#3227](https://github.com/cosmos/gaia/pull/3227))
- Bump [wasmd](https://github.com/CosmWasm/wasmd) to
    [v0.51.0](https://github.com/CosmWasm/wasmd/releases/tag/v0.51.0)
    ([\#3230](https://github.com/cosmos/gaia/pull/3230))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
    [v8.4.0](https://github.com/cosmos/ibc-go/releases/tag/v8.4.0)
    ([\#3233](https://github.com/cosmos/gaia/pull/3233))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
    [v5.1.1](https://github.com/cosmos/interchain-security/releases/tag/v5.1.1)
    ([\#3237](https://github.com/cosmos/gaia/pull/3237))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
    [v0.50.9-lsm](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.9-lsm)
    ([\#3249](https://github.com/cosmos/gaia/pull/3249))

## v18.0.0

*June 25, 2024*

### API BREAKING

- Remove both the globalfee module and the auth module 'DeductFeeDecorator'.
  ([\#3028](https://github.com/cosmos/gaia/pull/3028))
- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v4.3.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.0-lsm).
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))

### DEPENDENCIES

- Bump go version to 1.22
  ([\#3028](https://github.com/cosmos/gaia/pull/3028))
- Add the wasmd module.
  ([\#3051](https://github.com/cosmos/gaia/pull/3051))
- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v4.3.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.0-lsm).
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.6.0](https://github.com/cosmos/ibc-go/releases/tag/v7.6.0)
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.16-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.16-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))

### FEATURES

- Add the [feemarket module](https://github.com/skip-mev/feemarket) and set the initial params to the following values. ([\#3028](https://github.com/cosmos/gaia/pull/3028) and [\#3164](https://github.com/cosmos/gaia/pull/3164))
  ```
  FeeDenom = "uatom"
  DistributeFees = false // burn base fees
  MinBaseGasPrice = 0.005 // same as previously enforced by `x/globalfee`
  MaxBlockUtilization = 30_000_000 // the default value 
  ```
  
- Add the wasmd module.
  ([\#3051](https://github.com/cosmos/gaia/pull/3051))
- Enable both `MsgSoftwareUpgrade` and `MsgCancelUpgrade` to be expedited. 
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))

### STATE BREAKING

- Remove both the globalfee module and the auth module 'DeductFeeDecorator'.
  ([\#3028](https://github.com/cosmos/gaia/pull/3028))
- Add the [feemarket module](https://github.com/skip-mev/feemarket).
  ([\#3028](https://github.com/cosmos/gaia/pull/3028))
- Add the wasmd module.
  ([\#3051](https://github.com/cosmos/gaia/pull/3051))
- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v4.3.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.0-lsm).
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))
- Enable both `MsgSoftwareUpgrade` and `MsgCancelUpgrade` to be expedited. 
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.6.0](https://github.com/cosmos/ibc-go/releases/tag/v7.6.0)
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.16-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.16-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3149](https://github.com/cosmos/gaia/pull/3149))

## v17.2.0

*June 5, 2024*

### DEPENDENCIES

- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.15-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.15-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3134](https://github.com/cosmos/gaia/pull/3134))

### STATE BREAKING

- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.15-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.15-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3134](https://github.com/cosmos/gaia/pull/3134))

## v17.1.0

*June 4, 2024*

### DEPENDENCIES

- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.14-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.14-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3125](https://github.com/cosmos/gaia/pull/3125))

### STATE BREAKING

- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.14-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.14-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3125](https://github.com/cosmos/gaia/pull/3125))

## v17.0.0

*May 17, 2024*

### DEPENDENCIES

- Bump [CometBFT](https://github.com/cometbft/cometbft)
  to [v0.37.6](https://github.com/cometbft/cometbft/releases/tag/v0.37.6)
  ([\#3103](https://github.com/cosmos/gaia/pull/3103))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.2.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.2.0-lsm)
  ([\#3103](https://github.com/cosmos/gaia/pull/3103))

### FEATURES

- Add ICS 2.0 aka Partial Set Security (PSS). 
  See the [PSS docs](https://cosmos.github.io/interchain-security/features/partial-set-security) for more details.
  ([\#3103](https://github.com/cosmos/gaia/pull/3103))

### STATE BREAKING

- Add ICS 2.0 aka Partial Set Security (PSS)
  ([\#3103](https://github.com/cosmos/gaia/pull/3103))

## v16.0.0

*23rd April, 2024*

### DEPENDENCIES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware)
  to [v7.1.3](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv7.1.3).
  ([\#3021](https://github.com/cosmos/gaia/pull/3021))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.4.0](https://github.com/cosmos/ibc-go/releases/tag/v7.4.0)
  ([\#3039](https://github.com/cosmos/gaia/pull/3039))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.0-lsm)
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.1-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.1-lsm)
  ([\#3071](https://github.com/cosmos/gaia/pull/3071))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.13-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.13-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3078](https://github.com/cosmos/gaia/pull/3078))
- Bump [CometBFT](https://github.com/cometbft/cometbft)
  to [v0.37.5](https://github.com/cometbft/cometbft/releases/tag/v0.37.5)
  ([\#3078](https://github.com/cosmos/gaia/pull/3078))

### FEATURES

- Add ICA Controller sub-module
  ([\#3001](https://github.com/cosmos/gaia/pull/3001))
- Add the [IBC Rate Limit module](https://github.com/Stride-Labs/ibc-rate-limiting).
  ([\#3002](https://github.com/cosmos/gaia/pull/3002))
- Add the [IBC Fee Module](https://ibc.cosmos.network/v7/middleware/ics29-fee/overview).
  ([\#3038](https://github.com/cosmos/gaia/pull/3038))
- Add rate limits to IBC transfer channels cf.
  https://www.mintscan.io/cosmos/proposals/890.
  ([\#3042](https://github.com/cosmos/gaia/pull/3042))
- Initialize ICS epochs by adding a consumer validator set for every existing consumer chain.
  ([\#3079](https://github.com/cosmos/gaia/pull/3079))

### STATE BREAKING

- Add ICA Controller sub-module
  ([\#3001](https://github.com/cosmos/gaia/pull/3001))
- Add the [IBC Rate Limit module](https://github.com/Stride-Labs/ibc-rate-limiting).
  ([\#3002](https://github.com/cosmos/gaia/pull/3002))
- Add the [IBC Fee Module](https://ibc.cosmos.network/v7/middleware/ics29-fee/overview).
  ([\#3038](https://github.com/cosmos/gaia/pull/3038))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.4.0](https://github.com/cosmos/ibc-go/releases/tag/v7.4.0)
  ([\#3039](https://github.com/cosmos/gaia/pull/3039))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.0-lsm)
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.13-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.13-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3078](https://github.com/cosmos/gaia/pull/3078))
- Bump [CometBFT](https://github.com/cometbft/cometbft)
  to [v0.37.5](https://github.com/cometbft/cometbft/releases/tag/v0.37.5)
  ([\#3078](https://github.com/cosmos/gaia/pull/3078))
- Initialize ICS epochs by adding a consumer validator set for every existing consumer chain.
  ([\#3079](https://github.com/cosmos/gaia/pull/3079))

## v15.2.0

*March 29, 2024*

### BUG FIXES

- Increase x/gov metadata fields legth to 10200 ([\#3025](https://github.com/cosmos/gaia/pull/3025))
- Fix parsing of historic Txs with TxExtensionOptions ([\#3032](https://github.com/cosmos/gaia/pull/3032))

### STATE BREAKING

- Increase x/gov metadata fields legth to 10200 ([\#3025](https://github.com/cosmos/gaia/pull/3025))
- Fix parsing of historic Txs with TxExtensionOptions ([\#3032](https://github.com/cosmos/gaia/pull/3032))

## v15.1.0

*March 15, 2024*

### DEPENDENCIES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) to `v7.1.3-0.20240228213828-cce7f56d000b`.
  ([\#2982](https://github.com/cosmos/gaia/pull/2982))

### FEATURES

- Add gaiad snapshots command set ([\#2974](https://github.com/cosmos/gaia/pull/2974))

### STATE BREAKING

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) to `v7.1.3-0.20240228213828-cce7f56d000b`.
  ([\#2982](https://github.com/cosmos/gaia/pull/2982))
- Mint and transfer missing assets in escrow accounts
 to reach parity with counterparty chain supply.
 ([\#2993](https://github.com/cosmos/gaia/pull/2993))

## v15.0.0

*February 20, 2024*

### API BREAKING

- Reject `MsgVote` messages from accounts with less than 1 atom staked. 
  ([\#2912](https://github.com/cosmos/gaia/pull/2912))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.10-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10-ics-lsm).
  As compared to [v0.47.10](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10), 
  this special branch of cosmos-sdk has the following API-breaking changes:
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))
  - Limit the accepted deposit coins for a proposal to the minimum proposal deposit denoms (e.g., `uatom` for Cosmos Hub). ([sdk-#19302](https://github.com/cosmos/cosmos-sdk/pull/19302))
  - Add denom check to reject denoms outside of those listed in `MinDeposit`. A new `MinDepositRatio` param is added (with a default value of `0.01`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted. ([sdk-#19312](https://github.com/cosmos/cosmos-sdk/pull/19312))
  - Disable the `DenomOwners` query. ([sdk-#19266](https://github.com/cosmos/cosmos-sdk/pull/19266))
- The consumer CCV genesis state obtained from the provider chain needs to be 
  transformed to be compatible with older versions of consumer chains 
  (see [ICS docs](https://cosmos.github.io/interchain-security/consumer-development/consumer-genesis-transformation)). 
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))

### BUG FIXES

- Add ante handler that only allows `MsgVote` messages from accounts with at least
  1 atom staked. ([\#2912](https://github.com/cosmos/gaia/pull/2912))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.10-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10-ics-lsm).
  This special branch of cosmos-sdk backports a series of fixes for issues found 
  during the [Oak Security audit of SDK 0.47](https://github.com/oak-security/audit-reports/blob/master/Cosmos%20SDK/2024-01-23%20Audit%20Report%20-%20Cosmos%20SDK%20v1.0.pdf).
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))
  - Backport [sdk-#18146](https://github.com/cosmos/cosmos-sdk/pull/18146): Add denom check to reject denoms outside of those listed in `MinDeposit`. A new `MinDepositRatio` param is added (with a default value of `0.01`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted. ([sdk-#19312](https://github.com/cosmos/cosmos-sdk/pull/19312))
  - Partially backport [sdk-#18047](https://github.com/cosmos/cosmos-sdk/pull/18047): Add a limit of 200 grants pruned per `EndBlock` in the feegrant module. ([sdk-#19314](https://github.com/cosmos/cosmos-sdk/pull/19314))
  - Partially backport [skd-#18737](https://github.com/cosmos/cosmos-sdk/pull/18737): Add a limit of 200 grants pruned per `BeginBlock` in the authz module. ([sdk-#19315](https://github.com/cosmos/cosmos-sdk/pull/19315))
  - Backport [sdk-#18173](https://github.com/cosmos/cosmos-sdk/pull/18173): Gov Hooks now returns error and are "blocking" if they fail. Expect for `AfterProposalFailedMinDeposit` and `AfterProposalVotingPeriodEnded` that will log the error and continue. ([sdk-#19305](https://github.com/cosmos/cosmos-sdk/pull/19305))
  - Backport [sdk-#18189](https://github.com/cosmos/cosmos-sdk/pull/18189): Limit the accepted deposit coins for a proposal to the minimum proposal deposit denoms. ([sdk-#19302](https://github.com/cosmos/cosmos-sdk/pull/19302))
  - Backport [sdk-#18214](https://github.com/cosmos/cosmos-sdk/pull/18214) and [sdk-#17352](https://github.com/cosmos/cosmos-sdk/pull/17352): Ensure that modifying the argument to `NewUIntFromBigInt` and `NewIntFromBigInt` doesn't mutate the returned value. ([sdk-#19293](https://github.com/cosmos/cosmos-sdk/pull/19293))
  

### DEPENDENCIES

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.3.1](https://github.com/cosmos/ibc-go/releases/tag/v7.3.1)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v7.1.2](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv7.1.2)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Bump [CometBFT](https://github.com/cometbft/cometbft)
  to [v0.37.4](https://github.com/cometbft/cometbft/releases/tag/v0.37.4)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.10-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))
- Bump [ICS](https://github.com/cosmos/interchain-security) to 
  [v3.3.3-lsm](https://github.com/cosmos/interchain-security/releases/tag/v3.3.3-lsm) 
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))

### FEATURES

- Add support for metaprotocols using Tx extension options. 
  ([\#2960](https://github.com/cosmos/gaia/pull/2960))

### STATE BREAKING

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.3.1](https://github.com/cosmos/ibc-go/releases/tag/v7.3.1)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v7.1.2](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv7.1.2)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Bump [CometBFT](https://github.com/cometbft/cometbft)
  to [v0.37.4](https://github.com/cometbft/cometbft/releases/tag/v0.37.4)
  ([\#2852](https://github.com/cosmos/gaia/pull/2852))
- Set min commission rate staking parameter to `5%`
 ([prop 826](https://www.mintscan.io/cosmos/proposals/826))
 and update the commission rate for all validators that have a commission
 rate less than `5%`. ([\#2855](https://github.com/cosmos/gaia/pull/2855))
- Migrate the signing infos of validators for which the consensus address is missing. 
([\#2886](https://github.com/cosmos/gaia/pull/2886))
- Migrate vesting funds from "cosmos145hytrc49m0hn6fphp8d5h4xspwkawcuzmx498"
 to community pool according to signal prop [860](https://www.mintscan.io/cosmos/proposals/860).
 ([\#2891](https://github.com/cosmos/gaia/pull/2891))
- Add ante handler that only allows `MsgVote` messages from accounts with at least
  1 atom staked. ([\#2912](https://github.com/cosmos/gaia/pull/2912))
- Remove `GovPreventSpamDecorator` and initialize the `MinInitialDepositRatio` gov
  param to `10%`. 
  ([\#2913](https://github.com/cosmos/gaia/pull/2913))
- Add support for metaprotocols using Tx extension options. 
  ([\#2960](https://github.com/cosmos/gaia/pull/2960))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.10-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))
  - Skip running `addDenomReverseIndex` in `bank/v3` migration as it is prohibitively expensive to run on the Cosmos Hub. ([sdk-#19266](https://github.com/cosmos/cosmos-sdk/pull/19266))
- Bump [ICS](https://github.com/cosmos/interchain-security) to 
  [v3.3.3-lsm](https://github.com/cosmos/interchain-security/releases/tag/v3.3.3-lsm) 
  ([\#2967](https://github.com/cosmos/gaia/pull/2967))

## v14.2.0

*March 6, 2024*

**This is an emergency release.**

### DEPENDENCIES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) to `v4.1.2-0.20240228222021-455757bb5771`.
  ([\#2980](https://github.com/cosmos/gaia/pull/2980))

### STATE BREAKING

- Emergency patch for [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware).
  ([\#2980](https://github.com/cosmos/gaia/pull/2980))

## v14.1.0

*November 21, 2023*

### API BREAKING

- Deprecate equivocation proposals of ICS provider module ([\#2825](https://github.com/cosmos/gaia/pull/2825))

### DEPENDENCIES

- Bump [ICS] to [v2.4.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.4.0-lsm) ([\#2825](https://github.com/cosmos/gaia/pull/2825))

### FEATURES

- Set in the v14 upgrade handler the min evidence height for `neutron-1` 
  at `4552189` and for `stride-1` at `6375035`. 
  ([\#2821](https://github.com/cosmos/gaia/pull/2821))
- Introducing the cryptographic verification of equivocation feature to the ICS provider module ([\#2825](https://github.com/cosmos/gaia/pull/2825))

### STATE BREAKING

- Bump [ICS] to [v2.4.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.4.0-lsm) ([\#2825](https://github.com/cosmos/gaia/pull/2825))

## v14.0.0

*November 15, 2023*

❗***This release is deprecated and should not be used in production. Use v14.1.0 instead.***

### API BREAKING

- Deprecate equivocation proposals of ICS provider module ([\#2814](https://github.com/cosmos/gaia/pull/2814))

### DEPENDENCIES

- Bump [ICS] to [v2.3.0-provider-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.3.0-provider-lsm) ([\#2814](https://github.com/cosmos/gaia/pull/2814))

### FEATURES

- Introducing the cryptographic verification of equivocation feature to the ICS provider module ([\#2814](https://github.com/cosmos/gaia/pull/2814))

### STATE BREAKING

- Bump [ICS] to [v2.3.0-provider-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.3.0-provider-lsm) ([\#2814](https://github.com/cosmos/gaia/pull/2814))

## v13.0.2

*November 7, 2023*

### BUG FIXES

- Bump [cosmos/ledger-cosmos-go](https://github.com/cosmos/ledger-cosmos-go) to
  [v0.12.4](https://github.com/cosmos/ledger-cosmos-go/releases/tag/v0.12.4) 
  to fix signing with ledger through the binary on newest versions of macOS and Xcode
  ([\#2763](https://github.com/cosmos/gaia/pull/2763))

## v13.0.1

*October 25, 2023*

### BUG FIXES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v4.1.1](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.1)
  ([\#2771](https://github.com/cosmos/gaia/pull/2771))

### DEPENDENCIES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v4.1.1](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.1)
  ([\#2771](https://github.com/cosmos/gaia/pull/2771))

## v13.0.0

*September 18, 2023*

### DEPENDENCIES

- Remove [Liquidity](https://github.com/Gravity-Devs/liquidity)
  ([\#2716](https://github.com/cosmos/gaia/pull/2716))
- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v2.1.0-provider-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.1.0-provider-lsm)
  ([\#2732](https://github.com/cosmos/gaia/pull/2732))

### STATE BREAKING

- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v2.1.0-provider-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.1.0-provider-lsm)
  ([\#2732](https://github.com/cosmos/gaia/pull/2732))

## v12.0.0

*August 18, 2023*

### API BREAKING

- Add Liquid Staking Module (LSM) and initialize the LSM params: 
  ValidatorBondFactor, ValidatorLiquidStakingCap, GlobalLiquidStakingCap
  ([\#2643](https://github.com/cosmos/gaia/pull/2643))

### BUG FIXES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v4.1.0](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.0)
  ([\#2677](https://github.com/cosmos/gaia/pull/2677))

### DEPENDENCIES

- Bump [interchain-security](https://github.com/cosmos/interchain-security) to
  [v2.0.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v2.0.0-lsm)
  ([\#2643](https://github.com/cosmos/gaia/pull/2643))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.45.16-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.45.16-ics-lsm)
  ([\#2643](https://github.com/cosmos/gaia/pull/2643))
- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v4.1.0](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.0)
  ([\#2677](https://github.com/cosmos/gaia/pull/2677))

### FEATURES

- Add Liquid Staking Module (LSM) and initialize the LSM params: 
  ValidatorBondFactor, ValidatorLiquidStakingCap, GlobalLiquidStakingCap
  ([\#2643](https://github.com/cosmos/gaia/pull/2643))

### STATE BREAKING

- Add Liquid Staking Module (LSM) and initialize the LSM params: 
  ValidatorBondFactor, ValidatorLiquidStakingCap, GlobalLiquidStakingCap
  ([\#2643](https://github.com/cosmos/gaia/pull/2643))
- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware) 
  to [v4.1.0](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv4.1.0)
  ([\#2677](https://github.com/cosmos/gaia/pull/2677))

## v11.0.0

*July 18, 2023*

### API BREAKING

- [GlobalFee](x/globalfee)
  - Add `bypass-min-fee-msg-types` and `maxTotalBypassMinFeeMsgGagUsage` to
    globalfee params. `bypass-min-fee-msg-types` in `config/app.toml` is
    deprecated ([\#2424](https://github.com/cosmos/gaia/pull/2424))

### BUG FIXES

- Fix logic bug in `GovPreventSpamDecorator` that allows bypassing the 
  `MinInitialDeposit` requirement 
  ([a759409](https://github.com/cosmos/gaia/commit/a759409c9da2780663244308b430a7847b95139b))

### DEPENDENCIES

- Bump [PFM](https://github.com/strangelove-ventures/packet-forward-middleware) to 
  [v4.0.5](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v4.0.5)
  ([\#2185](https://github.com/cosmos/gaia/issues/2185))
- Bump [Interchain-Security](https://github.com/cosmos/interchain-security) to
  [v2.0.0](https://github.com/cosmos/interchain-security/releases/tag/v2.0.0)
  ([\#2616](https://github.com/cosmos/gaia/pull/2616))
- Bump [Liquidity](https://github.com/Gravity-Devs/liquidity) to 
  [v1.6.0-forced-withdrawal](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.6.0-forced-withdrawal) 
  ([\#2652](https://github.com/cosmos/gaia/pull/2652))

### STATE BREAKING

- General
  - Fix logic bug in `GovPreventSpamDecorator` that allows bypassing the
    `MinInitialDeposit` requirement
    ([a759409](https://github.com/cosmos/gaia/commit/a759409c9da2780663244308b430a7847b95139b))
  - Bump [Interchain-Security](https://github.com/cosmos/interchain-security) to
    [v2.0.0](https://github.com/cosmos/interchain-security/releases/tag/v2.0.0)
    ([\#2616](https://github.com/cosmos/gaia/pull/2616))
  - Bump [Liquidity](https://github.com/Gravity-Devs/liquidity) to
    [v1.6.0-forced-withdrawal](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.6.0-forced-withdrawal)
    ([\#2652](https://github.com/cosmos/gaia/pull/2652))
- [GlobalFee](x/globalfee)
  - Create the upgrade handler and params migration for the new Gloabal Fee module
    parameters introduced in [#2424](https://github.com/cosmos/gaia/pull/2424)
    ([\#2352](https://github.com/cosmos/gaia/pull/2352))
  - Add `bypass-min-fee-msg-types` and `maxTotalBypassMinFeeMsgGagUsage` to
    globalfee params ([\#2424](https://github.com/cosmos/gaia/pull/2424))
  - Update Global Fee's AnteHandler to check tx fees against the network min gas
    prices in DeliverTx mode ([\#2447](https://github.com/cosmos/gaia/pull/2447))

## v10.0.2

*July 03, 2023*

This release bumps several dependencies and enables extra queries. 

### DEPENDENCIES

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v4.4.2](https://github.com/cosmos/ibc-go/releases/tag/v4.4.2)
  ([\#2554](https://github.com/cosmos/gaia/pull/2554))
- Bump [CometBFT](https://github.com/cometbft/cometbft) to
  [v0.34.29](https://github.com/cometbft/cometbft/releases/tag/v0.34.29)
  ([\#2594](https://github.com/cosmos/gaia/pull/2594))

### FEATURES

- Register NodeService to enable query `/cosmos/base/node/v1beta1/config`
  gRPC query to disclose node operator's configured minimum-gas-price.
  ([\#2629](https://github.com/cosmos/gaia/issues/2629))

## [v10.0.1] 2023-05-25

* (deps) [#2543](https://github.com/cosmos/gaia/pull/2543) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.1](https://github.com/cosmos/ibc-go/releases/tag/v4.4.1).

## [v10.0.0] 2023-05-19

* (deps) [#2498](https://github.com/cosmos/gaia/pull/2498) Bump multiple dependencies. 
  * Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.16-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16-ics). See the [v0.45.16 release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16) for details. 
  * Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.0](https://github.com/cosmos/ibc-go/releases/tag/v4.4.0).
  * Bump [CometBFT](https://github.com/cometbft/cometbft) to [v0.34.28](https://github.com/cometbft/cometbft/releases/tag/v0.34.28).
* (gaia) Bump Golang prerequisite from 1.18 to 1.20. See (https://go.dev/blog/go1.20) for details.

## [v9.1.1] - 2023-05-25

* (deps) [#2542](https://github.com/cosmos/gaia/pull/2542) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.2.1](https://github.com/cosmos/ibc-go/releases/tag/v4.2.1).

## [v9.1.0] - 2023-05-08

* (fix) [#2474](https://github.com/cosmos/gaia/pull/2474) Multisig and distribution fix in [Interchain-Security](https://github.com/cosmos/interchain-security). Bump Interchain-Security to [v1.1.0-multiden](https://github.com/cosmos/interchain-security/tree/v1.1.0-multiden).

This release combines two fixes that we judged were urgent to get onto the Cosmos Hub before the launch of the first ICS consumer chain. _Please note that user funds were not at risk and these fixes pertain to the liveness of the Hub and consumer chains_.

The first fix is to enable the use of multisigs and Ledger devices when assigning keys for consumer chains. The second is to prevent a possible DOS vector involving the reward distribution system.

### Multisig fix

On April 25th (a week and a half ago), we began receiving reports that validators using multisigs and Ledger devices were getting errors reading Error: unable to resolve type URL /interchain_security.ccv.provider.v1.MsgAssignConsumerKey: tx parse error when attempting to assign consensus keys for consumer chains. 

This was surprising because we had never seen this error before, even though we have done many testnets. The reason for this is probably because people don’t bother to use high security key management techniques in testnets.

We quickly narrowed the problem down to issues having to do with using the PubKey type directly in the MsgAssignConsumerKey transaction, and Amino (a deprecated serialization library still used in Ledger devices and multisigs) not being able to handle this. We attempted to fix this with the assistance of the Cosmos-SDK team, but after making no headway for a few days, we decided to simply use a JSON representation of the PubKey in the transaction. This is how it is usually represented anyway. We have verified that this fixes the problem.

### Distribution fix

The ICS distribution system works by allowing consumer chains to send rewards to a module address on the Hub called the FeePoolAddress. From here they are automatically distributed to all validators and delegators through the distribution system that already exists to distribute Atom staking rewards. The FeePoolAddress is usually blocked so that no tokens can be sent to it, but to enable ICS distribution we had to unblock it.

We recently realized that unblocking the FeePoolAddress could enable an attacker to send a huge number of different denoms into the distribution system. The distribution system would then attempt to distribute them all, leading to out of memory errors. Fixing a similar attack vector that existed in the distribution system before ICS led us to this realization.

To fix this problem, we have re-blocked the FeePoolAddress and created a new address called the ConsumerRewardsPool. Consumer chains now send rewards to this new address. There is also a new transaction type called RegisterConsumerRewardDenom. This transaction allows people to register denoms to be used as rewards from consumer chains. It costs 10 Atoms to run this transaction.The Atoms are transferred to the community pool. Only denoms registered with this command are then transferred to the FeePoolAddress and distributed out to delegators and validators.

Note: The fee of 10 Atoms was originally intended to be a parameter that could be changed by governance (10 Atoms might cost too much in the future). However, we ran into some problems creating a new parameter as part of an emergency upgrade. After consulting with the Cosmos-SDK team, we learned that creating new parameters is only supported as part of a scheduled upgrade. So in the current code, the number of Atoms is hardcoded. It will turn into a parameter in the next scheduled upgrade.

## [v9.0.3] - 2023-04-19
* (deps) [#2399](https://github.com/cosmos/gaia/pull/2399) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.15-ics](https://github.com/cosmos/cosmos
sdk/releases/tag/v0.45.15-ics) and migrate to [CometBFT](https://github.com/cometbft/cometbft).

## [v9.0.2] - 2023-04-03
* (feat) Bump [Interchain-Security](https://github.com/cosmos/interchain-security) [v1.1.0](https://github.com/cosmos/interchain-security/releases/tag/v1.1.0) provider module. See the [release notes](https://github.com/cosmos/interchain-security/releases/tag/v1.1.0) for details.
* (feat) Add two more msg types `/ibc.core.channel.v1.MsgTimeout` and `/ibc.core.channel.v1.MsgTimeoutOnClose` to default `bypass-min-fee-msg-types`.
* (feat) Change the bypassing gas usage criteria. Instead of requiring 200,000 gas per `bypass-min-fee-msg`, we will now allow a maximum total usage of 1,000,000 gas for all bypassed messages in a transaction. Note that all messages in the transaction must be the `bypass-min-fee-msg-types` for the bypass min fee to take effect, otherwise, fee payment will still apply.
* (fix) [#2087](https://github.com/cosmos/gaia/issues/2087) Fix `bypass-min-fee-msg-types` parsing in `app.toml`. Parsing of `bypass-min-fee-types` is changed to allow node operators to use empty bypass list. Removing the `bypass-min-fee-types` from `app.toml` applies the default message types. See [#2092](https://github.com/cosmos/gaia/pull/2092) for details.

## [v9.0.1] - 2023-03-09

* (feat) [Add spam prevention antehandler](https://github.com/cosmos/gaia/pull/2262) to alleviate recent governance spam issues.

## [v9.0.0] - 2023-02-21

* (feat) Add [Interchain-Security](https://github.com/cosmos/interchain-security) [v1.0.0](https://github.com/cosmos/interchain-security/releases/tag/v1.0.0) provider module. See the [ICS Spec](https://github.com/cosmos/ibc/blob/main/spec/app/ics-028-cross-chain-validation/README.md) for more details.
* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.13-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.13-ics). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/releases/tag/v0.45.13-ics) for details.
* (gaia) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.2.0](https://github.com/cosmos/ibc-go/blob/release/v4.2.x/CHANGELOG.md). See [v4.2 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v4.2.0) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.26](https://github.com/informalsystems/tendermint/tree/v0.34.26). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.26/CHANGELOG.md#v03426) for details.
* (gaia) Bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v4.0.4](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v4.0.4).
* (tests) Add [E2E ccv tests](https://github.com/cosmos/gaia/blob/main/tests/e2e/e2e_gov_test.go#L138). Tests covering new functionality introduced by the provider module to add and remove a consumer chain via governance proposal.
* (tests) Add [integration ccv tests](https://github.com/cosmos/gaia/blob/main/tests/ics/interchain_security_test.go). Imports Interchain-Security's `TestCCVTestSuite` and implements Gaia as the provider chain.
* (fix) [#2017](https://github.com/cosmos/gaia/issues/2017) Fix Gaiad binary build tag for ubuntu system. See [#2018](https://github.com/cosmos/gaia/pull/2018) for details.

## [v8.0.1] - 2023-02-17

* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.14). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.26](https://github.com/informalsystems/tendermint/tree/v0.34.26). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.26/CHANGELOG.md) for details.

## [v8.0.0] - 2023-01-31

* (gaia) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v3.4.0](https://github.com/cosmos/ibc-go/blob/v3.4.0/CHANGELOG.md) to fix a vulnerability in ICA. See [v3.4.0 CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.9/CHANGELOG.md) and [v3.2.1 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v3.2.1) for details.
* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.12). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.24](https://github.com/informalsystems/tendermint/tree/v0.34.24). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.24/CHANGELOG.md) for details.
* (gaia) Bump [liquidity](https://github.com/Gravity-Devs/liquidity) to [v1.5.3](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.3).
* (gaia) Bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v3.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v3.1.1).
* (feat) Add [globalfee](https://github.com/cosmos/gaia/tree/main/x/globalfee) module. See [globalfee docs](https://github.com/cosmos/gaia/blob/main/docs/modules/globalfee.md) for more details.
* (feat) [#1845](https://github.com/cosmos/gaia/pull/1845) Add bech32-convert command to gaiad.
* (fix) [#2080](https://github.com/cosmos/gaia/issues/2074) Reintroduce deleted configuration for client rpc endpoints, transaction routes, and module REST routes in app.go.
* (fix) [Add new fee decorator](https://github.com/cosmos/gaia/pull/1961) to change `MaxBypassMinFeeMsgGasUsage` so importers of x/globalfee can change `MaxGas`.
* (fix) [#1870](https://github.com/cosmos/gaia/issues/1870) Fix bank denom metadata in migration. See [#1892](https://github.com/cosmos/gaia/pull/1892) for more details.
* (fix) [#1976](https://github.com/cosmos/gaia/pull/1976) Fix Quicksilver ICA exploit in migration. See [the bug fix forum post](https://forum.cosmos.network/t/upcoming-interchain-accounts-bugfix-release/8911) for more details.
* (tests) Add [E2E tests](https://github.com/cosmos/gaia/tree/main/tests/e2e). The tests cover transactions/queries tests of different modules, including Bank, Distribution, Encode, Evidence, FeeGrant, Global Fee, Gov, IBC, packet forwarding middleware, Slashing, Staking, and Vesting module.
* (tests) [#1941](https://github.com/cosmos/gaia/pull/1941) Fix packet forward configuration for e2e tests.
* (tests) Use gaiad to swap out [Ignite](https://github.com/ignite/cli) in [liveness tests](https://github.com/cosmos/gaia/blob/main/.github/workflows/test.yml).

## [v7.1.1] - 2023-02-06

* (gaia) bump [tendermint](https://github.com/tendermint/tendermint) to [0.34.25](https://github.com/informalsystems/tendermint/releases/tag/v0.34.25) to patch p2p issue. See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.25/CHANGELOG.md#v03425) for details.

## [v7.1.0] - 2022-10-14
* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.9) to fix the security vulnerability! See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.9/CHANGELOG.md) for details.

## [v7.0.3] - 2022-08-03
* (gaia) update go to 1.18.
* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.6). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.6/CHANGELOG.md) for details.
* (gaia) bump [Liquidity](https://github.com/Gravity-Devs/liquidity) module to [v1.5.1](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.1).
* (gaia) bump [cosmos ledger](https://github.com/cosmos/ledger-go) to [v0.9.3](https://github.com/cosmos/ledger-go/releases/tag/v0.9.3) to fix issue [#1573](https://github.com/cosmos/gaia/issues/1573) - Ledger Nano S Plus not detected by gaiad.
* 
## [v7.0.2] -2022-05-09

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.4). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.4/CHANGELOG.md#v0454---2022-04-25) for details.
* (gaia) [#1447](https://github.com/cosmos/gaia/pull/1447) Support custom message types to bypass minimum fee checks for.
  If a transaction contains only bypassed message types, the transaction will not have minimum fee
  checks performed during `CheckTx`. Operators can supply these message types via the `bypass-min-fee-msg-types`
  configuration in `app.toml`. Note, by default they include various IBC message types.

## [v7.0.1] -2022-04-13

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.3). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/CHANGELOG.md#v0453---2022-04-12) for details.
* (gaia) bump [tendermint](https://github.com/tendermint/tendermint) to [0.34.19](https://github.com/tendermint/tendermint/tree/v0.34.19). See [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/v0.34.19/CHANGELOG.md#v03419) for details.
* (gaia) bump [tm-db](https://github.com/tendermint/tm-db) to [v0.6.7](https://github.com/tendermint/tm-db/tree/v0.6.7). See [CHANGELOG.md](https://github.com/tendermint/tm-db/blob/v0.6.7/CHANGELOG.md#067) for details.

## [v7.0.0] - 2022-03-24

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.1/CHANGELOG.md#v0451---2022-02-03) for details.
* (gaia) bump [ibc-go](https://github.com/cosmos/ibc-go) module to [v3.0.0](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0). See [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v3.0.0/CHANGELOG.md#v300---2022-03-15) for details.
* (gaia) add [interchain account](https://github.com/cosmos/ibc-go/tree/main/modules/apps/27-interchain-accounts) module (interhchain-account module is part of ibc-go module).
* (gaia) bump [liquidity](https://github.com/gravity-devs/liquidity) module to [v1.5.0](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.0). See [CHANGELOG.md](https://github.com/Gravity-Devs/liquidity/blob/v1.5.0/CHANGELOG.md#v150---20220223) for details.
* (gaia) bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) module to [v2.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v2.1.1).
* (gaia) add migration logs for upgrade process.

## [v6.0.4] - 2022-03-10

* (gaia) Bump [Liquidity](https://github.com/gravity-devs/liquidity) module to [v1.4.6](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.4.6).
* (gaia) Bump [IBC](https://github.com/cosmos/ibc-go) module to [2.0.3](https://github.com/cosmos/ibc-go/releases/tag/v2.0.3).
* (gaia) [#1230](https://github.com/cosmos/gaia/pull/1230) Fix: update gRPC Web Configuration in `contrib/testnets/test_platform`.
* (gaia) [#1135](https://github.com/cosmos/gaia/pull/1135) Fix rocksdb build tag usage.
* (gaia) [#1160](https://github.com/cosmos/gaia/pull/1160) Improvement: update state sync configs.
* (gaia) [#1208](https://github.com/cosmos/gaia/pull/1208) Update statesync.bash.
  * * (gaia) Bump [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) to [v0.44.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.6)
* (gaia) Bump [Versions](https://github.com/cosmos/gaia/pull/1100) of various smaller dependencies, remove the Cosmos SDK replace statement, update `initiClientCtx` params, ensure `stdout` and `stderr` are handled correctly in the CLI.

## [v6.0.3] - 2022-02-18

* This is a reverted release that is the same as v6.0.0

## [v6.0.2] - 2022-02-17

* Unusable release

## [v6.0.1] - 2022-02-10

* Unusable release

## [v6.0.0] - 2021-11-24

* (gaia) Add NewSetUpContextDecorator to anteDecorators
* (gaia) Reconfigure SetUpgradeHandler to ensure vesting is configured after auth and new modules have InitGenesis run.
* (golang) Bump golang prerequisite to 1.17.
* (gaia) Bump [Liquidity](https://github.com/gravity-devs/liquidity) module to [v1.4.2](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.4.2).
* (gaia) Bump [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) to [v0.44.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.3). See the [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/CHANGELOG.md#v0443---2021-10-21) for details.
* (gaia) Add [IBC](https://github.com/cosmos/ibc-go) as a standalone module from the Cosmos SDK using version [v2.0.0](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0). See the [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v2.0.0/CHANGELOG.md) for details.
* (gaia) Add [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) [v1.0.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v1.0.1).
* (gaia) [#969](https://github.com/cosmos/gaia/issues/969) Remove legacy migration code.

## [v5.0.8] - 2021-10-14

* (gaia) This release includes a new AnteHandler that rejects redundant IBC transactions to save relayers fees.

## [v5.0.7] - 2021-09-30

* (gaia) Bump Cosmos SDK to 0.42.10

## [v5.0.6] - 2021-09-16

* (gaia) Bump tendermint to 0.34.13

## [v5.0.5] - 2021-08-05

* (gaia) Bump SDK to [0.42.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.9) to resolve IBC channel restart issue ([9800](https://github.com/cosmos/cosmos-sdk/issues/9800)).

## [v5.0.4] - 2021-07-31

* (chore) Fix release to include intended items from `v5.0.3`.

## [v5.0.3] - 2021-07-30

* (gaia) Bump SDK to [0.42.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.8) to resolve tx query issues.
* (gaia) Bump SDK to [0.42.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.7) to resolve state sync issues.

## [v5.0.2] - 2021-07-15

* (gaia) Fix misspelling in RELEASE.md
* (gaia) Add releases to .gitignore

## [v5.0.1] - 2021-07-15

* (gaia) Configure gaiad command to add back `config` capabilities.

## [v5.0.0] - 2021-06-28

* (golang) Bump golang prerequisite from 1.15 to 1.16.
* (gaia) Add [Liquidity](https://github.com/gravity-devs/liquidity) module [v1.2.9](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.2.9).
* (sdk)  Bump SDK version to [v0.42.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.6).
* (tendermint) Bump Tendermint version to [v0.34.11](https://github.com/tendermint/tendermint/releases/tag/v0.34.11).

## [v4.2.1] - 2021-04-08

A critical security vulnerability was identified in Tendermint Core, which impacts Tendermint Lite Client.

This release fixes the identified security vulnerability.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.4)
* (tendermint) Bump Tendermint version to [v0.34.9](https://github.com/tendermint/tendermint/releases/tag/v0.34.9).

## [v4.2.0] - 2021-03-25

A critical security vulnerability has been identified in Gaia v4.1.x.
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This release fixes the identified security vulnerability.

If the chain halts before or during the upgrade, validators with sufficient voting power need to upgrade
and come online in order for the chain to resume.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.3)

## [v4.1.2] - 2021-03-22

This release removes unnecessary dependencies.

### Bug Fixes

* (gaia)  [\#781](https://github.com/cosmos/gaia/pull/781) Remove unnecessary dependencies

## [v4.1.1] - 2021-03-19

This release bring improvements to keyring UX, tx search results, and multi-sig account migrations.
See the Cosmos SDK [release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.2) for details.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.2)

## [v4.1.0] - 2021-03-10

### Bug Fixes

A critical security vulnerability has been identified in Gaia v4.0.x.
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This release fixes the identified security vulnerability.

If the chain halts before or during the upgrade, validators with sufficient voting power need to upgrade
and come online in order for the chain to resume.

## [v4.0.6] - 2021-03-09

### Bug Fixes

This release bumps the Cosmos SDK, which includes an important security fix for all non
Cosmos Hub chains (e.g. any chain that does not use the default cosmos bech32 prefix),
and a few performance improvements. The SDK also applies a security fix for validator
address conversion in evidence handling, and the full header is now emitted on an
IBC UpdateClient message event.

* (sdk)  Bump SDK version to [v0.42.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.0)

## [v4.0.5] - 2021-03-02

### Bug Fixes

* (tendermint) Bump Tendermint version to [v0.34.8](https://github.com/tendermint/tendermint/releases/tag/v0.34.8).
* (sdk)  Bump SDK version to [v0.41.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.4), which reduces startup time with the cosmoshub-4 mainnet genesis without invariant checks.

## [v4.0.4] - 2021-02-19

### Bug Fixes

This release applies a patched version to grpc dependencies in order to resolve some queries; no explicit version bumps are included.

## [v4.0.3] - 2021-02-18

### Bug Fixes

This release fixes build failures caused by a small API breakage introduced in tendermint v0.34.7.

* (sdk)  Bump SDK version to [v0.41.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.3).

## [v4.0.2] - 2021-02-18

### Bug Fixes

This release fixes a downstream security issue which impacts Cosmos SDK users.
See the [Tendermint v0.34.7 SDK changelog](https://github.com/tendermint/tendermint/blob/v0.34.x/CHANGELOG.md#v0347) for details.

* (sdk) [\#640](https://github.com/cosmos/gaia/pull/640) Bump SDK version to [v0.41.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.2).
* (tendermint) [\#640](https://github.com/cosmos/gaia/pull/640) Bump Tendermint version to [v0.34.7](https://github.com/tendermint/tendermint/releases/tag/v0.34.7).

## [v4.0.1] - 2021-02-17

### Bug Fixes

* (sdk) [\#579](https://github.com/cosmos/gaia/pull/635) Bump SDK version to [v0.41.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.1).
* (tendermint) [\#622](https://github.com/cosmos/gaia/pull/622) Bump Tendermint version to [v0.34.4](https://github.com/tendermint/tendermint/releases/tag/v0.34.4).

## [v4.0.0] - 2021-01-26

### Improvements

* (app) [\#564](https://github.com/cosmos/gaia/pull/564) Add client denomination metadata for atoms.

### Bug Fixes

* (cmd) [\#563](https://github.com/cosmos/gaia/pull/563) Add balance coin to supply when adding a new genesis account
* (sdk) [\#579](https://github.com/cosmos/gaia/pull/579) Bump SDK version to [v0.41.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.0).

## [v3.0.1] - 2021-01-19

### Improvements

* (protobuf) [\#553](https://github.com/cosmos/gaia/pull/553) Bump gogo protobuf deps to v1.3.3
* (github) [\#543](https://github.com/cosmos/gaia/pull/543) Add docker deployment
* (starport) [\#535](https://github.com/cosmos/gaia/pull/535) Add config.yml
* (docker) [\#534](https://github.com/cosmos/gaia/pull/534) Update to python3

### Bug Fixes

* (sdk) Bump SDK version to [v0.40.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.1).
* (tendermint) Bump Tendermint version to [v0.34.3](https://github.com/tendermint/tendermint/releases/tag/v0.34.3).
* (github) [\#544](https://github.com/cosmos/gaia/pull/544) Deploy from main not master
* (docs) [\#550](https://github.com/cosmos/gaia/pull/550) Bump vuepress-theme-cosmos to 1.0.180
* (docker) [\#537](https://github.com/cosmos/gaia/pull/537) Fix single-node.sh setup script

## [v3.0.0] - 2021-01-09

### Improvements

* (sdk) Bump SDK version to [v0.40.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.0).
* (tendermint) Bump Tendermint version to [v0.34.1](https://github.com/tendermint/tendermint/releases/tag/v0.34.1).

## [v2.0.14] - 2020-12-10

* (sdk) Bump SDK version to [v0.37.15](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.15).
* (tendermint) Bump Tendermint version to [v0.32.14](https://github.com/tendermint/tendermint/releases/tag/v0.32.14).

## [v2.0.13] - 2020-08-13

* (sdk) Bump SDK version to [v0.37.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.14).

## [v2.0.12] - 2020-08-13

* This version did not contain the update to [v0.37.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.14). Please use v2.0.13

## [v2.0.11] - 2020-05-06

* (sdk) Bump SDK version to [v0.37.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.13).

## [v2.0.10] - 2020-05-06

* (sdk) Bump SDK version to [v0.37.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.12).

## [v2.0.9] - 2020-04-23

* (sdk) Bump SDK version to [v0.37.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.11).

## [v2.0.8] - 2020-04-09

### Improvements

* (sdk) Bump SDK version to [v0.37.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.9).

## [v2.0.7] - 2020-03-11

### Improvements

* (sdk) Bump SDK version to [v0.37.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.8).

## [v2.0.6] - 2020-02-10

### Improvements

* (sdk) Bump SDK version to [v0.37.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.7).

## [v2.0.5] - 2020-01-21

### Improvements

* (sdk) Bump SDK version to [v0.37.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.6).
* (tendermint) Bump Tendermint version to [v0.32.9](https://github.com/tendermint/tendermint/releases/tag/v0.32.9).

## [v2.0.4] - 2020-01-09

### Improvements

* (sdk) Bump SDK version to [v0.37.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.5).
* (tendermint) Bump Tendermint version to [v0.32.8](https://github.com/tendermint/tendermint/releases/tag/v0.32.8).

### Bug Fixes

* (cli) Fixed `gaiacli query txs` to use `events` instead of `tags`. Events take the form of `'{eventType}.{eventAttribute}={value}'`. Please
  see the [events doc](https://github.com/cosmos/cosmos-sdk/blob/master/docs/core/events.md#events-1)
  for further documentation.

## [v2.0.3] - 2019-11-04

### Improvements

* (sdk) Bump SDK version to [v0.37.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.4).
* (tendermint) Bump Tendermint version to [v0.32.7](https://github.com/tendermint/tendermint/releases/tag/v0.32.7).

## [v2.0.2] - 2019-10-12

### Improvements

* (sdk) Bump SDK version to [v0.37.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.3).
* (tendermint) Bump Tendermint version to [v0.32.6](https://github.com/tendermint/tendermint/releases/tag/v0.32.6).

## [v2.0.1] - 2019-09-20

### Features

* (gaiad) [\#119](https://github.com/cosmos/gaia/pull/119) Add support for the `--halt-time` CLI flag and configuration.

### Improvements

* [\#119](https://github.com/cosmos/gaia/pull/119) Refactor and upgrade Circle CI
  configuration.
* (sdk) Update SDK version to v0.37.1

## [v2.0.0] - 2019-08-22

### Bug Fixes

* [\#104](https://github.com/cosmos/gaia/issues/104) Fix `ModuleAccountAddrs` to
  not rely on the `x/supply` keeper to get module account addresses for blacklisting.

### Breaking Changes

* (sdk) Update SDK version to v0.37.0

## [v1.0.0] - 2019-08-13

### Bug Fixes

* (gaiad) [\#4113](https://github.com/cosmos/cosmos-sdk/issues/4113) Fix incorrect `$GOBIN` in `Install Go`
* (gaiacli) [\#3945](https://github.com/cosmos/cosmos-sdk/issues/3945) There's no check for chain-id in TxBuilder.SignStdTx
* (gaiacli) [\#4190](https://github.com/cosmos/cosmos-sdk/issues/4190) Fix redelegations-from by using the correct params and query endpoint.
* (gaiacli) [\#4219](https://github.com/cosmos/cosmos-sdk/issues/4219) Return an error when an empty mnemonic is provided during key recovery.
* (gaiacli) [\#4345](https://github.com/cosmos/cosmos-sdk/issues/4345) Improved Ledger Nano X detection

### Breaking Changes

* (sdk) Update SDK version to v0.36.0
* (gaiad) [\#3985](https://github.com/cosmos/cosmos-sdk/issues/3985) ValidatorPowerRank uses potential consensus power
* (gaiad) [\#4027](https://github.com/cosmos/cosmos-sdk/issues/4027) gaiad version command does not return the checksum of the go.sum file shipped along with the source release tarball.
  Go modules feature guarantees dependencies reproducibility and as long as binaries are built via the Makefile shipped with the sources, no dependendencies can break such guarantee.
* (gaiad) [\#4159](https://github.com/cosmos/cosmos-sdk/issues/4159) use module pattern and module manager for initialization
* (gaiad) [\#4272](https://github.com/cosmos/cosmos-sdk/issues/4272) Merge gaiareplay functionality into gaiad replay.
  Drop `gaiareplay` in favor of new `gaiad replay` command.
* (gaiacli) [\#3715](https://github.com/cosmos/cosmos-sdk/issues/3715) query distr rewards returns per-validator
  rewards along with rewards total amount.
* (gaiacli) [\#40](https://github.com/cosmos/cosmos-sdk/issues/40) rest-server's --cors option is now gone.
* (gaiacli) [\#4027](https://github.com/cosmos/cosmos-sdk/issues/4027) gaiacli version command dooes not return the checksum of the go.sum file anymore.
* (gaiacli) [\#4142](https://github.com/cosmos/cosmos-sdk/issues/4142) Turn gaiacli tx send's --from into a required argument.
  New shorter syntax: `gaiacli tx send FROM TO AMOUNT`
* (gaiacli) [\#4228](https://github.com/cosmos/cosmos-sdk/issues/4228) Merge gaiakeyutil functionality into gaiacli keys.
  Drop `gaiakeyutil` in favor of new `gaiacli keys parse` command. Syntax and semantic are preserved.
* (rest) [\#3715](https://github.com/cosmos/cosmos-sdk/issues/3715) Update /distribution/delegators/{delegatorAddr}/rewards GET endpoint
  as per new specs. For a given delegation, the endpoint now returns the
  comprehensive list of validator-reward tuples along with the grand total.
* (rest) [\#3942](https://github.com/cosmos/cosmos-sdk/issues/3942) Update pagination data in txs query.
* (rest) [\#4049](https://github.com/cosmos/cosmos-sdk/issues/4049) update tag MsgWithdrawValidatorCommission to match type
* (rest) The `/auth/accounts/{address}` now returns a `height` in the response. The
  account is now nested under `account`.

### Features

* (gaiad) Add `migrate` command to `gaiad` to provide the ability to migrate exported
  genesis state from one version to another.
* (gaiad) Update Gaia for community pool spend proposals per Cosmos Hub governance proposal [\#7](https://github.com/cosmos/cosmos-sdk/issues/7) "Activate the Community Pool"

### Improvements

* (gaiad) [\#4042](https://github.com/cosmos/cosmos-sdk/issues/4042) Update docs and scripts to include the correct `GO111MODULE=on` environment variable.
* (gaiad) [\#4066](https://github.com/cosmos/cosmos-sdk/issues/4066) Fix 'ExportGenesisFile() incorrectly overwrites genesis'
* (gaiad) [\#4064](https://github.com/cosmos/cosmos-sdk/issues/4064) Remove `dep` and `vendor` from `doc` and `version`.
* (gaiad) [\#4080](https://github.com/cosmos/cosmos-sdk/issues/4080) add missing invariants during simulations
* (gaiad) [\#4343](https://github.com/cosmos/cosmos-sdk/issues/4343) Upgrade toolchain to Go 1.12.5.
* (gaiacli) [\#4068](https://github.com/cosmos/cosmos-sdk/issues/4068) Remove redundant account check on `gaiacli`
* (gaiacli) [\#4227](https://github.com/cosmos/cosmos-sdk/issues/4227) Support for Ledger App v1.5
* (rest) [\#2007](https://github.com/cosmos/cosmos-sdk/issues/2007) Return 200 status code on empty results
* (rest) [\#4123](https://github.com/cosmos/cosmos-sdk/issues/4123) Fix typo, url error and outdated command description of doc clients.
* (rest) [\#4129](https://github.com/cosmos/cosmos-sdk/issues/4129) Translate doc clients to chinese.
* (rest) [\#4141](https://github.com/cosmos/cosmos-sdk/issues/4141) Fix /txs/encode endpoint

<!-- Release links -->

[v10.0.1]: https://github.com/cosmos/gaia/releases/tag/v10.0.1
[v10.0.0]: https://github.com/cosmos/gaia/releases/tag/v10.0.0
[v9.1.1]: https://github.com/cosmos/gaia/releases/tag/v9.1.1
[v9.1.0]: https://github.com/cosmos/gaia/releases/tag/v9.1.0
[v9.0.3]: https://github.com/cosmos/gaia/releases/tag/v9.0.3
[v9.0.2]: https://github.com/cosmos/gaia/releases/tag/v9.0.2
[v9.0.1]: https://github.com/cosmos/gaia/releases/tag/v9.0.1
[v9.0.0]: https://github.com/cosmos/gaia/releases/tag/v9.0.0
[v8.0.1]: https://github.com/cosmos/gaia/releases/tag/v8.0.1
[v8.0.0]: https://github.com/cosmos/gaia/releases/tag/v8.0.0
[v7.1.1]: https://github.com/cosmos/gaia/releases/tag/v7.1.1
[v7.1.0]: https://github.com/cosmos/gaia/releases/tag/v7.1.0
[v7.0.3]: https://github.com/cosmos/gaia/releases/tag/v7.0.3
[v7.0.2]: https://github.com/cosmos/gaia/releases/tag/v7.0.2
[v7.0.1]: https://github.com/cosmos/gaia/releases/tag/v7.0.1
[v7.0.0]: https://github.com/cosmos/gaia/releases/tag/v7.0.0
[v6.0.4]: https://github.com/cosmos/gaia/releases/tag/v6.0.4
[v6.0.3]: https://github.com/cosmos/gaia/releases/tag/v6.0.3
[v6.0.2]: https://github.com/cosmos/gaia/releases/tag/v6.0.2
[v6.0.1]: https://github.com/cosmos/gaia/releases/tag/v6.0.1
[v6.0.0]: https://github.com/cosmos/gaia/releases/tag/v6.0.0
[v5.0.8]: https://github.com/cosmos/gaia/releases/tag/v5.0.8
[v5.0.7]: https://github.com/cosmos/gaia/releases/tag/v5.0.7
[v5.0.6]: https://github.com/cosmos/gaia/releases/tag/v5.0.6
[v5.0.5]: https://github.com/cosmos/gaia/releases/tag/v5.0.5
[v5.0.4]: https://github.com/cosmos/gaia/releases/tag/v5.0.4
[v5.0.3]: https://github.com/cosmos/gaia/releases/tag/v5.0.3
[v5.0.2]: https://github.com/cosmos/gaia/releases/tag/v5.0.2
[v5.0.1]: https://github.com/cosmos/gaia/releases/tag/v5.0.1
[v5.0.0]: https://github.com/cosmos/gaia/releases/tag/v5.0.0
[v4.2.1]: https://github.com/cosmos/gaia/releases/tag/v4.2.1
[v4.2.0]: https://github.com/cosmos/gaia/releases/tag/v4.2.0
[v4.1.2]: https://github.com/cosmos/gaia/releases/tag/v4.1.2
[v4.1.1]: https://github.com/cosmos/gaia/releases/tag/v4.1.1
[v4.1.0]: https://github.com/cosmos/gaia/releases/tag/v4.1.0
[v4.0.6]: https://github.com/cosmos/gaia/releases/tag/v4.0.6
[v4.0.5]: https://github.com/cosmos/gaia/releases/tag/v4.0.5
[v4.0.4]: https://github.com/cosmos/gaia/releases/tag/v4.0.4
[v4.0.3]: https://github.com/cosmos/gaia/releases/tag/v4.0.3
[v4.0.2]: https://github.com/cosmos/gaia/releases/tag/v4.0.2
[v4.0.1]: https://github.com/cosmos/gaia/releases/tag/v4.0.1
[v4.0.0]: https://github.com/cosmos/gaia/releases/tag/v4.0.0
[v3.0.1]: https://github.com/cosmos/gaia/releases/tag/v3.0.1
[v3.0.0]: https://github.com/cosmos/gaia/releases/tag/v3.0.0
[v2.0.14]: https://github.com/cosmos/gaia/releases/tag/v2.0.14
[v2.0.13]: https://github.com/cosmos/gaia/releases/tag/v2.0.13
[v2.0.12]: https://github.com/cosmos/gaia/releases/tag/v2.0.12
[v2.0.11]: https://github.com/cosmos/gaia/releases/tag/v2.0.11
[v2.0.10]: https://github.com/cosmos/gaia/releases/tag/v2.0.10
[v2.0.9]: https://github.com/cosmos/gaia/releases/tag/v2.0.9
[v2.0.8]: https://github.com/cosmos/gaia/releases/tag/v2.0.8
[v2.0.7]: https://github.com/cosmos/gaia/releases/tag/v2.0.7
[v2.0.6]: https://github.com/cosmos/gaia/releases/tag/v2.0.6
[v2.0.5]: https://github.com/cosmos/gaia/releases/tag/v2.0.5
[v2.0.4]: https://github.com/cosmos/gaia/releases/tag/v2.0.4
[v2.0.3]: https://github.com/cosmos/gaia/releases/tag/v2.0.3
[v2.0.2]: https://github.com/cosmos/gaia/releases/tag/v2.0.2
[v2.0.1]: https://github.com/cosmos/gaia/releases/tag/v2.0.1
[v2.0.0]: https://github.com/cosmos/gaia/releases/tag/v2.0.0
[v1.0.0]: https://github.com/cosmos/gaia/releases/tag/v1.0.0

