---
title: On-Chain Proposal Process
order: 2
---

## Governance Parameters

Several of the numbers involved in governance are parameters and can thus be changed by passing a parameter change proposal. 
- Minimum deposit: 250 ATOM
- Maximum deposit period: 14 days
- Voting period: 14 days
- Quorum: 40% of participating voting power
- Pass threshold: 50% of participating voting power
- Veto threshold: 33.40% of participating voting power

## 1. Deposit Period

The deposit period lasts either 14 days or until the proposal deposit totals 250 ATOMs, whichever happens first. 

### Deposits

Prior to a governance proposal entering the voting period (i.e., for the proposal to be voted upon), there must be at least a minimum number of ATOMs deposited (250). Anyone may contribute to this deposit, though it is usually filled by the proposal maker. Deposits of passed and failed proposals are returned to the contributors.

In the past, different people have considered contribution amounts differently. There is some consensus that this should be a personal choice. There is also some consensus that this can be an opportunity for supporters to signal their support by adding to the deposit amount, so a proposer may choose to leave contribution room (i.e., a deposit below 250 ATOMs) so that others may participate. It is important to remember that any contributed ATOMs are at risk of being burned.

### Burned deposits

Deposits are burned only when proposals are vetoed as documented in the [Cosmos SDK gov module spec](https://docs.cosmos.network/main/modules/gov#deposit-refund-and-burn). Deposits are not burned for failing to meet quorum or for being rejected. 

## 2. Voting Period

The voting period is currently a fixed 14-day period. During the voting period, participants may select a vote of either 'Yes', 'No', 'Abstain', or 'NoWithVeto'. Voters may change their vote at any time before the voting period ends. 

### What do the voting options mean?

1. **Abstain:** The voter wishes to contribute to quorum without voting for or against a proposal.
2. **Yes:** Approval of the proposal in its current form.
3. **No:** Disapproval of the proposal in its current form.
4. **NoWithVeto:** A ‘NoWithVeto’ vote indicates a proposal either (1) is deemed to be spam, i.e., irrelevant to Cosmos Hub, (2) disproportionately infringes on minority interests, or (3) violates or encourages violation of the rules of engagement as currently set out by Cosmos Hub governance.

As accepted by the community in [Proposal 75](https://ipfs.io/ipfs/QmVHVH9WeGy9tTNN9dViqvDn7N79XJJUseKXD1rpyLVckK), voters are expected to vote 'NoWithVeto' for proposals that are spam, infringe on minority interests, or violate the rules of engagement (i.e., Social protocols which have passed governance and thus been accepted as rules on the Hub). This proposal was an extension of the ideas put forward in [Proposal 6](https://ipfs.io/ipfs/QmRtR7qkeaZCpCzHDwHgJeJAZdTrbmHLxFDYXhw7RoF1pp).

Voting 'NoWithVeto' has no immediate additional financial cost to the voter - you do not directly risk your ATOM by using this option.

### What determines whether or not a governance proposal passes?

There are four criteria:

1. Deposit is filled: A minimum deposit of 250 ATOM is required for the proposal to enter the voting period
   - anyone may contribute to this deposit
   - the deposit must be reached within 14 days (this is the deposit period)
2. Quorum is reached: A minimum of 40% of the network's total voting power (staked ATOM) is required to participate 
3. Simple majority of 'Yes' votes: Greater than 50% of the participating voting power must back the 'Yes' vote by the end of the 14-day voting period
4. Not vetoed: Less than 33.4% of participating voting power must have backed 'NoWithVeto' by the end of the 14-day voting period

Currently, the criteria for submitting and passing/failing all proposal types is the same.

### How is quorum determined?

Voting power, whether backing a vote of 'Yes', 'Abstain', 'No', or 'NoWithVeto', counts toward quorum. Quorum is required for the outcome of a governance proposal vote to be considered valid and for deposit contributors to recover their deposit amounts. 

### How is voting tallied?

- **Total voting power** refers to all staked ATOM at the end of the 14-day voting period. Liquid ATOMs are not part of the total voting power and thus cannot participate in voting. 
- **Participating voting power** refers to only the ATOM which have been used to cast a vote on a particular proposal. Quorum is set to 40% of the **participating** voting power.

Validators not in the active set can cast a vote, but their voting power (including the backing of their delegators) will not count toward the vote if they are not in the active set **when the voting period ends**. That means that if ATOM is delegated to a validator that is jailed, tombstoned, or outside of the active set at the time that the voting period ends, that ATOM's stake-weight will not count in the vote.

Though a simple majority 'Yes' vote (ie. 50% of participating voting power) is required for a governance proposal vote to pass, a 'NoWithVeto' vote of 33.4% of participating voting power or greater can override this outcome and cause the proposal to fail. This enables a minority group representing greater than 1/3 of participating voting power to fail a proposal that would otherwise pass.
