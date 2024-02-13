- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.8-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.8-ics-lsm).
  As compared to [v0.47.8](https://github.com/cosmos/cosmos-sdk/tree/v0.47.8), 
  this special branch of cosmos-sdk has the following API-breaking changes:
  ([\#2919](https://github.com/cosmos/gaia/pull/2919))
  - Limit the accepted deposit coins for a proposal to the minimum proposal deposit denoms (e.g., `uatom` for Cosmos Hub). ([sdk-#19302](https://github.com/cosmos/cosmos-sdk/pull/19302))
  - Add denom check to reject denoms outside of those listed in `MinDeposit`. A new `MinDepositRatio` param is added (with a default value of `0.01`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted. ([sdk-#2919](https://github.com/cosmos/cosmos-sdk/pull/19312))
  - Disable the `DenomOwners` query. ([sdk-#19266](https://github.com/cosmos/cosmos-sdk/pull/19266))