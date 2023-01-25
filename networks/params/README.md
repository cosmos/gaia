## Genesis Parameters

Many genesis fields are self-evident, null, or uncontroversial (e.g. gas prices, which are chosen for spam prevention).

Here the more subjective parameter choices are documented with the reasons behind their recommendation.

Note that all durations are specified in nanoseconds.

### Staking Module

- `"unbonding_time": "1814400000000000"`. The unbonding time determines the duration for which bonded stake is
  held accountable for any discovered equivocations, specified in nanoseconds. 3 weeks was chosen to balance
  the concerns of a sufficient unbonding period for lite client safety and a modicum of staking token liquidity.
- `"max_validators": "100"`. The maximum validator count is the total number of validators which can be bonded
  and voting in consensus for any given block - which validators are in this set is dynamically determined
  to be the top hundred validator candidates sorted by delegated stake. The value of `100` was specified in the Cosmos whitepaper.
  It is expected to grow over time, but automatic increases aren't yet
  implemented.

### Minting Module

- `"inflation": "0.07"`. The initial annual inflation rate will be 7%, as specified in the Cosmos whitepaper.
- `"inflation_max": "0.2"`. The maximum annual inflation rate will be 20%, as specified in the Cosmos whitepaper.
- `"inflation_min": "0.07"`. The minimum annual inflation rate will be 7%, as specified in the Cosmos whitepaper.
- `"inflation_rate_change": "0.13"`. The rate at which the inflation rate changes (second derivative of inflation),
  per year squared, will be 13%, as specified in the Cosmos whitepaper.

### Distribution Module

- `"community_tax": "0.02"`. The tax on inflation and fees levied to fund the public goods pool will be 2%,
  as specified in the Cosmos whitepaper.
- `"base_proposer_reward": "0.01"`. 1% of inflation and fees (flat) will be allocated to the block proposer. This provides an incentive for 
    validators to be good proposers by being available when it's their turn to propose, including lots of transactions in their proposed block, and
    gossiping the proposed block quickly.
- `"bonus_proposer_reward": "0.04"`. 4% of inflation and fees (varying according to the fraction of precommits included)
  will be allocated to the block proposer to incentivize them to include as many
  precommits from other validators as possible.
- `"withdraw_addr_enabled": false`. Changing reward withdrawal addresses will be initially disabled. It may later be enabled via a hard fork.

### Governance Module

- `"min_deposit": 512atom`. The minimum deposit to bring a proposal up for a vote is 512 ATOMs. Because the price of ATOMs is uncertain at launch, this value should be high enough to prevent spam proposals, while not being too expensive. As a note, the deposit can be crowd-funded, so the proposer doesn't have to provide the whole thing. Proposals which pass refund all deposits.
- `"max_deposit_period": "1209600000000000"`. The duration in which a proposal can collect deposits is 14 days. This value should be long enough for a proposal to have time to gain support from the community.
- `"voting_period": "1209600000000000"`. The duration in which a proposal can be voted upon is 14 days. The voting period should be long enough that all staked ATOM holders had time to participate.
- `"quorum": "0.4"`. A minimum quorum of 40% of bonded stake must vote on a proposal in order for it to be considered for passage. This is to ensure that proposals don't pass that have support from only a small segment of the community.
- `"threshold": "0.5"`. Over half the voting stake must vote in favor of a proposal in order for it to pass.
- `"veto": "0.334"`. 1/3 of voting stake vetoing a proposal prevents it from passing. This is necessitated by the 1/3 BFT safety bound,
  since 1/3 of stake could also elect to halt the chain or compromise safety.

### Slashing Module

- `"max_evidence_age": "1814400000000000"`. The maximum age of evidence possibly considered valid is 3 weeks
  (it must be the same as the unbonding period).
- `"signed_blocks_window": "10000"`. The rolling window for uptime measurement is 10,000 blocks.
- `"min_signed_per_window": "0.05"`. A minimum of 5% of the blocks in the last window must have been signed or
  else a validator will be slashed for downtime. To nurture network launch, a lenient uptime requirement is recommended that can later be increased by governance.
- `"downtime_jail_duration": "600000000000"`. Validators slashed for downtime are jailed for ten minutes. This provides a disincentive for validator downtime.
- `"slash_fraction_double_sign": "0.05"`. Validators who equivocate (double-sign a block, and thereby compromise safety)
  and are caught are slashed by 5% of their bonded stake.
- `"slash_fraction_downtime": "0.0001"`. Validators who are slashed for downtime and thereby compromise the availability
  of the network are slashed by 0.01% of their bonded stake. This is to provide additional disincentive for validator downtime.
