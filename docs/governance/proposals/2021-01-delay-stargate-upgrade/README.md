# Delay of Hub Stargate Upgrade
 
The Stargate team is recommending that the Cosmos Hub reschedule the next upgrade to a new commit hash. The new commit hash is expected to be available on Tuesday Jan 26th with a new upgrade proposal immediately after.
 
This governance proposal will signal that [proposal 35](https://www.mintscan.io/cosmos/proposals/35) will not be executed. The Hub governance will vote on the forthcoming proposal aiming for a final upgrade. The earliest target date would be February 11th. Given that Lunar New Year is on Feb 12th. The next best date is Feb 18th 06:00UTC.
 
We are recommending the delay for the following reasons.
 
* Bugs have been identified in the Proposal 29 implementation.  They are resolved in this pull request[Additional review of prop 29 and migration testing by zmanian · Pull Request #559 · cosmos/gaia · GitHub](https://github.com/cosmos/gaia/pull/559)
* A balance validation regression was identified during Prop 29 code review. [x/bank: balance and metadata validation by fedekunze · Pull Request #8417 · cosmos/cosmos-sdk · GitHub](https://github.com/cosmos/cosmos-sdk/pull/8417)
* The IBC Go To Market Working Group has [identified Ledger hardware wallet](https://github.com/cosmos/cosmos-sdk/issues/8266) support as a necessary feature for the initial launch of IBC on the Hub. We have an opportunity to provide this support in this upgrade. The SDK believes this can be quickly remediated in the time available with merged PRs on Monday.
* The number of Stargate related support requests from integrators has increased significantly since the governance proposal went live but some teams have already announced a period of reduced $ATOM support while they upgrade like <https://twitter.com/Ledger_Support/status/1352247403605356551?s=20>. The additional time should minimize the disruption for $ATOM holders. Thank so much to the $IRIS team who is fielding a similar request volume among our non-English community.