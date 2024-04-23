# CHANGELOG

## v16.0.0

*23rd April, 2024*

### DEPENDENCIES

- Bump [PFM](https://github.com/cosmos/ibc-apps/tree/main/middleware)
  to [v7.1.3](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv7.1.3).
  ([\#3021](https://github.com/cosmos/gaia/pull/3021))
- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v7.4.0](https://github.com/cosmos/ibc-go/releases/tag/v7.4.0)
  ([\#3039](https://github.com/cosmos/gaia/pull/3039))
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.12-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.12-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.0-lsm)
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.1-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.1-lsm)
  ([\#3071](https://github.com/cosmos/gaia/pull/3071))

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
- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.12-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.12-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.1.0-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.1.0-lsm)
  ([\#3062](https://github.com/cosmos/gaia/pull/3062))
- Initialize ICS epochs by adding a consumer validator set for every existing consumer chain.
  ([\#3079](https://github.com/cosmos/gaia/pull/3079))

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

