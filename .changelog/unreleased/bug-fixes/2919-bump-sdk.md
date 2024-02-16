- Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to
  [v0.47.8-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.8-ics-lsm).
  This special branch of cosmos-sdk backports a series of fixes for issues found 
  during the [Oak Security audit of SDK 0.47](https://github.com/oak-security/audit-reports/blob/master/Cosmos%20SDK/2024-01-23%20Audit%20Report%20-%20Cosmos%20SDK%20v1.0.pdf).
  ([\#2919](https://github.com/cosmos/gaia/pull/2919))
  - Backport [sdk-#18146](https://github.com/cosmos/cosmos-sdk/pull/18146): Add denom check to reject denoms outside of those listed in `MinDeposit`. A new `MinDepositRatio` param is added (with a default value of `0.01`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted. ([sdk-#2919](https://github.com/cosmos/cosmos-sdk/pull/19312))
  - Partially backport [sdk-#18047](https://github.com/cosmos/cosmos-sdk/pull/18047): Add a limit of 200 grants pruned per `EndBlock` in the feegrant module. ([sdk-#19314](https://github.com/cosmos/cosmos-sdk/pull/19314))
  - Partially backport [skd-#18737](https://github.com/cosmos/cosmos-sdk/pull/18737): Add a limit of 200 grants pruned per `BeginBlock` in the authz module. ([sdk-#19315](https://github.com/cosmos/cosmos-sdk/pull/19315))
  - Backport [sdk-#18173](https://github.com/cosmos/cosmos-sdk/pull/18173): Gov Hooks now returns error and are "blocking" if they fail. Expect for `AfterProposalFailedMinDeposit` and `AfterProposalVotingPeriodEnded` that will log the error and continue. ([sdk-#19305](https://github.com/cosmos/cosmos-sdk/pull/19305))
  - Backport [sdk-#18189](https://github.com/cosmos/cosmos-sdk/pull/18189): Limit the accepted deposit coins for a proposal to the minimum proposal deposit denoms. ([sdk-#19302](https://github.com/cosmos/cosmos-sdk/pull/19302))
  - Backport [sdk-#18214](https://github.com/cosmos/cosmos-sdk/pull/18214) and [sdk-#17352](https://github.com/cosmos/cosmos-sdk/pull/17352): Ensure that modifying the argument to `NewUIntFromBigInt` and `NewIntFromBigInt` doesn't mutate the returned value. ([sdk-#19293](https://github.com/cosmos/cosmos-sdk/pull/19293))
  