# CHANGELOG

## v21.0.1

*November 21, 2024*

### BUG FIXES

- Bump [cosmossdk.io/math](https://github.com/cosmos/cosmos-sdk/tree/main/math) to
  [v1.4.0](https://github.com/cosmos/cosmos-sdk/tree/math/v1.4.0) in order to 
  address the the [ASA-2024-010](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-7225-m954-23v7) 
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

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

