# CHANGELOG

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

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

