---
order: 3
parent:
  order: 1
---

# Community Pool Spend

Cosmos Hub launched with community-spend capabilities on December 11, 2019, effectively unlocking the potential for token-holders to vote to approve spending from the Community Pool.

🇪🇸 Esta página también está [disponible en español](https://github.com/raquetelio/CosmosCommunitySpend/blob/master/README%5BES_es%5D.md).

## Learn About the Community Pool

### How is the Community Pool funded?

2% of all staking rewards generated (via block rewards & transaction fees) are continually transferred to and accrue within the Community Pool. For example, from Dec 19, 2019 until Jan 20, 2020 (32 days), 28,726 ATOM were generated and added to the pool.

### How can funding for the Community Pool change?

Though the rate of funding is currently fixed at 2% of staking rewards, the effective rate is dependent upon the Cosmos Hub's staking rewards, which can change with inflation and block times.

The current parameter `Community Tax` parameter of 2% may be modified with a governance proposal and enacted immediately after the proposal passes.

### How much money is in the Community Pool?

You may directly query the Cosmos Hub 4 for the balance of the Community Pool:

```gaiad q distribution community-pool --chain-id cosmoshub-4 --node <rpc-node-address> ```

Alternatively, popular Cosmos explorers such as [Big Dipper](https://cosmos.bigdipper.live) and [Mintscan](https://www.mintscan.io/cosmos) display the ongoing Community Pool balance.

### How can funds from the Community Pool be spent?

Funds from the Cosmos Community Pool may be spent via successful governance proposal.

### How should funds from the Community Pool be spent?

We don't know 🤷

The prevailing assumption is that funds should be spent in a way that brings value to the Cosmos Hub. However, there is debate about how to keep the fund sustainable. There is also some debate about who should receive funding. For example, part of the community believes that the funds should only be used for those who need funding most. Other topics of concern include:

- retroactive grants
- price negotiation
- fund disbursal (eg. payments in stages; payments pegged to reduce volatility)
- radical overhaul of how the community-spend mechanism functions

We can expect this to take shape as proposals are discussed, accepted, and rejected by the Cosmos Hub community.

### How are funds disbursed after a community-spend proposal is passed?

If a community-spend proposal passes successfully, the number of ATOM encoded in the proposal will be transferred from the community pool to the address encoded in the proposal, and this will happen immediately after the voting period ends.

## Why create a proposal to use Community Pool funds?

There are other funding options, most notably the Interchain Foundation's grant program. Why create a community-spend proposal?

**As a strategy: you can do both.** You can submit your proposal to the Interchain Foundation, but also consider submitting your proposal publicly on-chain. If the Hub votes in favour, you can withdraw your Interchain Foundation application.

**As a strategy: funding is fast.** Besides the time it takes to push your proposal on-chain, the only other limiting factor is a fixed 14-day voting period. As soon as the proposal passes, your account will be credited the full amount of your proposal request.

**To build rapport.** Engaging publicly with the community is the opportunity to develop relationships with stakeholders and to educate them about the importance of your work. Unforeseen partnerships could arise, and overall the community may value your work more if they are involved as stakeholders.

**To be more independent.** The Interchain Foundation (ICF) may not always be able to fund work. Having a more consistently funded source and having a report with its stakeholders means you can use your rapport to have confidence in your ability to secure funding without having to be dependent upon the ICF alone.
