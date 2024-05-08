# Parameter change: lower minimum proposal deposit amount

## Summary

The current deposit amount of 512 ATOMs prohibits valuable governance activity from small holders or those with most of their ATOM staked. We propose lowering the requirement to 64 ATOMS.

## Objectives

1. Enable community members with good ideas but little capital to participate in governance and request resources from the community pool treasury.
2. Improve the governance UX for holders who keep most of their ATOM staked.
3. Increase utilisation of treasury (currently 666,457 ATOM, approximately $14 MM USD, at time of writing).
4. Accelerate Cosmos Hub development and growth.

## Background

Current deposit is 512 ATOMs (approximately $10k USD today). The ATOM price when the community treasury was activated (2019-05-03) was $4.99 (source: CoinMarketCap), meaning the total required deposit to submit a proposal was $2,555. Today, most proposers must coordinate with large ATOM holders to request additional funds in order to meet the minimum deposit requirements. This also applies to large ATOM holders who want to be active in governance but do not have enough liquid ATOM to meet the deposit requirements, as staked ATOM cannot be used to post deposits.

## Proposers

Federico Kunze Küllmer (Tharsis) and Sam Hart (Interchain).

Credit to Gavin Birch (Figment Networks) and the Cosmos Governance Working Group (GWG) for initiating a recent conversation that motivated this proposal.

## Proposed Parameter Change

Change the minimum proposal deposit requirement from 512 ATOMs (aprox. $10,000 USD) to 64 ATOMs (aprox. $1,300 USD).

Note: Parameters are denominated in micro-ATOMs, as described in the [governance parameter list](https://github.com/cosmos/governance/blob/master/params-change/Governance.md).

## Risks

__This change makes it easier to submit spam proposals.__

While this is true, in order to fully mitigate spam the Cosmos Hub must increase the minimum deposit required for proposal [submission](https://cosmoscan.net/proposal/28).

__By increasing the number of submissions, voter participation or the level of consideration given to each proposal may decrease.__

We believe this is a justifiable trade-off for promoting more community-driven initiatives and enthusiasm for advancing Cosmos. As we lower the barrier to entry for governance participation, we invite community members to take this opportunity to enact more effective and efficient governance practices. The upcoming Groups, Authz, and Interchain Accounts modules will provide powerful abstractions to this end.

## Alternatives

__Wait for the Cosmos Hub to adopt proposed changes to the Governance module for variable deposit amounts, quorom thresholds, and voting periods.__

These initiatives should not be mutually exclusive. While research and development of these features is ongoing, the Cosmos Hub will benefit from this parameter change today, as well as the precedent it sets for self-improving governance.

__Since the ATOM price fluctuates with respect to USD, make proposal thresholds reference a stable price oracle__

This is an interesting design space, however it becomes more plausible if and when the Cosmos Hub adds a decentralized exchange that can be used to produce a reference rate. Lowering the proposal threshold is a temporary solution that will help in the short-term.

## Governance Votes

The following items summarise the voting options and what it means for this proposal.

- **YES**: You approve the parameter change proposal to decrease the governance proposal deposit requirements from 512 to 64 ATOMs.
- **NO**: You disapprove of the parameter change in its current form (please indicate in the Cosmos Forum why this is the case).
- **NO WITH VETO**: You are strongly opposed to this change and will exit the network if passed.
- **ABSTAIN**: You are impartial to the outcome of the proposal.