- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.8-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.8-ics-lsm).
  This is a special cosmos-sdk branch with support for both ICS and LSM.
  ([\#2919](https://github.com/cosmos/gaia/pull/2919))
  - Skip running `addDenomReverseIndex` in `bank/v3` migration as it is prohibitively expensive to run on the Cosmos Hub. ([sdk-#19266](https://github.com/cosmos/cosmos-sdk/pull/19266))