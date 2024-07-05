# CHANGELOG

## v18.1.0

*July 5, 2024*

### DEPENDENCIES

- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.3.1-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.1-lsm)
  ([\#3187](https://github.com/cosmos/gaia/pull/3187))

### STATE BREAKING

- Bump [ICS](https://github.com/cosmos/interchain-security) to
  [v4.3.1-lsm](https://github.com/cosmos/interchain-security/releases/tag/v4.3.1-lsm)
  ([\#3187](https://github.com/cosmos/gaia/pull/3187))

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

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

