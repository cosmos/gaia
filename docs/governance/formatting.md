---
order: 4
---

# Formatting a Proposal

Many proposals allow for long form text to be included, usually under the key `description`. These provide the opportunity to include [markdown](https://docs.github.com/en/github/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) if formatted correctly as well as line breaks with `\n`. Beware, however, that if you are using the CLI to create a proposal, and setting `description` using a flag, the text will be [escaped](https://en.wikipedia.org/wiki/Escape_sequences_in_C) which may have undesired effects. If you're using markdown or line breaks it's recommended to put the proposal text into a json file and include that file as part of the CLI proposal, as opposed to individual fields in flags.

## Text

Text proposals are used by delegators to agree to a certain strategy, plan, commitment, future upgrade, or any other statement in the form of text. Aside from having a record of the proposal outcome on the Cosmos Hub chain, a text proposal has no direct effect on the change Cosmos Hub.

### Real example

[Proposal 12](https://hubble.figment.io/cosmos/chains/cosmoshub-4/governance/proposals/12) asked if the Cosmos Hub community of validators charging 0% commission was harmful to the success of the Cosmos Hub.

```
{
  "title": "Are Validators Charging 0% Commission Harmful to the Success of the Cosmos Hub?",
  "description": "This governance proposal is intended to act purely as a signalling proposal. Throughout this history of the Cosmos Hub, there has been much debate about the impact that validators charging 0% commission has on the Cosmos Hub, particularly with respect to the decentralization of the Cosmos Hub and the sustainability for validator operations. \n\n Discussion around this topic has taken place in many places including numerous threads on the Cosmos Forum, public Telegram channels, and in-person meetups.  Because this has been one of the primary discussion points in off-chain Cosmos governance discussions, we believe it is important to get a signal on the matter from the on-chain governance process of the Cosmos Hub. \n\n There have been past discussions on the Cosmos Forum about placing an in-protocol restriction on validators from charging 0% commission.  https://forum.cosmos.network/t/governance-limit-validators-from-0-commission-fee/2182 \n\n This proposal is NOT proposing a protocol-enforced minimum.  It is merely a signalling proposal to query the viewpoint of the bonded Atom holders as a whole. \n\n We encourage people to discuss the question behind this governance proposal in the associated Cosmos Hub forum post here:  https://forum.cosmos.network/t/proposal-are-validators-charging-0-commission-harmful-to-the-success-of-the-cosmos-hub/2505 \n\n Also, for voters who believe that 0% commission rates are harmful to the network, we encourage optionally sharing your belief on what a healthy minimum commission rate for the network using the memo field of their vote transaction on this governance proposal or linking to a longer written explanation such as a Forum or blog post. \n\n The question on this proposal is \u201cAre validators charging 0% commission harmful to the success of the Cosmos Hub?\u201d.  A Yes vote is stating that they ARE harmful to the network's success, and a No vote is a statement that they are NOT harmful.",
  "deposit": "100000umuon"
}
```

## Community Pool Spend

There are five (5) components:
1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Recipient** - the Cosmos Hub (bech32-based) address that will receive funding from the Community Pool
4. **Amount** - the amount of funding that the recipient will receive in micro-ATOMs (uatom)
5. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Examples

In this simple example (below), a network explorer will list the governance proposal as "Community Pool Spend." When an observer selects the proposal, they'll see the description. Not all explorers will show the recipient and amount, so ensure that you verify that the description aligns with the what the governance proposal is programmed to enact. If the description says that a certain address will receive a certain number of ATOMs, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).

The `amount` is `1000000uatom`. 1,000,000 micro-ATOM is equal to 1 ATOM, so `recipient` address `cosmos1qgfdn8h6fkh0ekt4n4d2c93c5gz3cv5gce783m` will receive 1 ATOM if this proposal is passed.

The `deposit 512000000 uatom` results in 512 ATOM being used from the proposal submitter's account. There is a minimum deposit required for a proposal to enter the voting period, and anyone may contribute to this deposit within a 14-day period. If the minimum deposit isn't reach before this time, the deposit amounts will be burned. Deposit amounts will also be burned if quorum isn't met in the vote or if the proposal is vetoed.

```
{
  "title": "Community Pool Spend",
  "description": "This is the summary of the key information about this proposal. Include the URL to a PDF version of your full proposal.",
  "recipient": "cosmos1qgfdn8h6fkh0ekt4n4d2c93c5gz3cv5gce783m",
  "amount": [
    {
      "denom": "uatom",
      "amount": "1000000"
    }
  ],
  "deposit": [
    {
      "denom": "uatom",
      "amount": "512000000"
    }
  ]
}

```

#### Real Example

This is the governance proposal that [Gavin Birch](https://twitter.com/Ether_Gavin) ([Figment Networks](https://figment.network/)) used to create [Prop23, the first successful Cosmos Hub community-spend proposal](https://hubble.figment.network/cosmos/chains/cosmoshub-3/governance/proposals/23). 

You can query the proposal details with the gaiad command-line interface using this command: `gaiad q gov proposal 23 --chain-id cosmoshub-3 --node cosmos-node-1.figment.network:26657`

You use can also use [Hubble](https://hubble.figment.network/cosmos/chains/cosmoshub-3/blocks/424035/transactions/B8E2662DE82413F03919712B18F7B23AF00B50DAEB499DAD8C436514640EFC79?format=json) or gaiad to query the transaction that I sent to create this proposal on-chain in full detail: `gaiad q tx B8E2662DE82413F03919712B18F7B23AF00B50DAEB499DAD8C436514640EFC79 --chain-id cosmoshub-3 --node cosmos-node-1.figment.network:26657`

**Note**: "\n" is used to create a new line.

```
{
  "title": "Cosmos Governance Working Group - Q1 2020",
  "description": "Cosmos Governance Working Group - Q1 2020 funding\n\nCommunity-spend proposal submitted by Gavin Birch (https://twitter.com/Ether_Gavin) of Figment Networks (https://figment.network)\n\n-=-=-\n\nFull proposal: https://ipfs.io/ipfs/QmSMGEoY2dfxADPfgoAsJxjjC6hwpSNx1dXAqePiCEMCbY\n\n-=-=-\n\nAmount to spend from the community pool: 5250 ATOMs\n\nTimeline: Q1 2020\n\nDeliverables:\n1. A governance working group community & charter\n2. A template for community spend proposals\n3. A best-practices document for community spend proposals\n4. An educational wiki for the Cosmos Hub parameters\n5. A best-practices document for parameter changes\n6. Monthly governance working group community calls (three)\n7. Monthly GWG articles (three)\n8. One Q2 2020 GWG recommendations article\n\nMilestones:\nBy end of Month 1, the Cosmos Governance Working Group (GWG) should have been initiated and led by Gavin Birch of Figment Networks.\nBy end of Month 2, Gavin Birch is to have initiated and led GWG’s education, best practices, and Q2 recommendations.\nBy end of Month 3, Gavin Birch is to have led and published initial governance education, best practices, and Q2 recommendations.\n\nDetailed milestones and funding:\nhttps://docs.google.com/spreadsheets/d/1mFEvMSLbiHoVAYqBq8lo3qQw3KtPMEqDFz47ESf6HEg/edit?usp=sharing\n\nBeyond the milestones, Gavin will lead the GWG to engage in and answer governance-related questions on the Cosmos Discourse forum, Twitter, the private Cosmos VIP Telegram channel, and the Cosmos subreddit. The GWG will engage with stake-holders to lower the barriers to governance participation with the aim of empowering the Cosmos Hub’s stakeholders. The GWG will use this engagement to guide recommendations for future GWG planning.\n\nRead more about the our efforts to launch the Cosmos GWG here: https://figment.network/resources/introducing-the-cosmos-governance-working-group/\n\n-=-=-\n\n_Problem_\nPerhaps the most difficult barrier to effective governance is that it demands one of our most valuable and scarce resources: our attention. Stakeholders may be disadvantaged by informational or resource-based asymmetries, while other entities may exploit these same asymmetries to capture value controlled by the Cosmos Hub’s governance mechanisms.\n\nWe’re concerned that without establishing community standards, processes, and driving decentralized delegator-based participation, the Cosmos Hub governance mechanism could be co-opted by a centralized power. As governance functionality develops, potential participants will need to understand how to assess proposals by knowing what to pay attention to.\n\n_Solution_\nWe’re forming a focused, diverse group that’s capable of assessing and synthesizing the key parts of a proposal so that the voting community can get a fair summary of what they need to know before voting.\n\nOur solution is to initiate a Cosmos governance working group that develops decentralized community governance efforts alongside the Hub’s development. We will develop and document governance features and practices, and then communicate these to the broader Cosmos community.\n\n_Future_\nAt the end of Q1, we’ll publish recommendations for the future of the Cosmos GWG, and ideally we’ll be prepared to submit a proposal based upon those recommendations for Q2 2020. We plan to continue our work in blockchain governance, regardless of whether the Hub passes our proposals.\n\n-=-=-\n\nCosmos forum: https://forum.cosmos.network/c/governance\nCosmos GWG Telegram channel: https://t.me/hubgov\nTwitter: https://twitter.com/CosmosGov",
  "recipient": "cosmos1hjct6q7npsspsg3dgvzk3sdf89spmlpfg8wwf7",
  "amount": [
    {
      "denom": "uatom",
      "amount": "5250000000"
    }
  ],
  "deposit":"12000000uatom"
}
```

## Params Change

**Note:** Changes to the [`gov` module](./Governance.md) are different from the other kinds of parameter changes because `gov` has subkeys, [as discussed here](https://github.com/cosmos/cosmos-sdk/issues/5800). Only the `key` part of the JSON file is different for `gov` parameter-change proposals.

For parameter-change proposals, there are seven (7) components:
1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Subspace** - the Cosmos Hub module with the parameter that is being changed
4. **Key** - the parameter that will be changed
5. **Value** - the value of the parameter that will be changed by the governance mechanism
6. **Denom** - `uatom` (micro-ATOM) will be the type of asset used as the deposit
7. **Amount** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Examples

In this simple example ([below](#testnet-example)), a network explorer will list the governance proposal by its title: "Increase the minimum deposit amount for governance proposals." When a user selects the proposal, they'll see the proposal’s description. A nearly identical proposal [can be found on the gaia-13007 testnet here](https://hubble.figment.network/cosmos/chains/gaia-13007/governance/proposals/30).

Not all explorers will show the proposed parameter changes that are coded into the proposal, so ensure that you verify that the description aligns with what the governance proposal is programmed to enact. If the description says that a certain parameter will be increased, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).

You can query the proposal details with the gaiad command-line interface using this command: `gaiad q gov proposal 30 --chain-id gaia-13007 --node 45.77.218.219:26657`

You use can also use [Hubble](https://hubble.figment.network/cosmos/chains/gaia-13007/transactions/B5AB56719ADB7117445F6E191E1FCE775135832AFE6C9922B8703AADBC4B13F3?format=json) or gaiad to query the transaction that I sent to create a similar proposal on-chain in full detail: `gaiad q tx B5AB56719ADB7117445F6E191E1FCE775135832AFE6C9922B8703AADBC4B13F3 --chain-id gaia-13007 --node 45.77.218.219:26657`

#### Testnet Example: changing a parameter from the `gov` module

```
{
  "title": "Increase the minimum deposit amount for governance proposals",
  "description": "If successful, this parameter-change governance proposal that will change the minimum deposit from 0.1 to 0.2 testnet ATOMs.",
  "changes": [
    {
      "subspace": "gov",
      "key": "depositparams",
      "value": {"mindeposit":"200000umuon"}
    }
  ],
  "deposit": "100000umuon"
}
```
The deposit `denom` is `uatom` and `amount` is `100000`. Since 1,000,000 micro-ATOM is equal to 1 ATOM, a deposit of 0.1 ATOM will be included with this proposal. The gaia-13007 testnet currently has a 0.1 ATOM minimum deposit, so this will put the proposal directly into the voting period. There is a minimum deposit required for a proposal to enter the voting period, and anyone may contribute to this deposit within a 14-day period. If the minimum deposit isn't reached before this time, the deposit amounts will be burned. Deposit amounts will also be burned if quorum isn't met in the vote or if the proposal is vetoed.

### Mainnet Example
 
To date, the Cosmos Hub's parameters have not been changed by a parameter-change governance proposal. This is a hypothetical example of the JSON file that would be used with a command line transaction to create a new proposal. This is an example of a proposal that changes two parameters, and both parameters are from the [`slashing` module](./Slashing.md). A single parameter-change governance proposal can reportedly change any number of parameters.

```
{
  "title": "Parameter changes for validator downtime",
  "description": "If passed, this governance proposal will do two things:\n\n1. Increase the slashing penalty for downtime from 0.01% to 0.50%\n2. Decrease the window \n\nIf this proposal passes, validators must sign at least 5% of 5,000 blocks, which is 250 blocks. That means that a validator that misses 4,750 consecutive blocks will be considered by the system to have committed a liveness violation, where previously 9,500 consecutive blocks would need to have been missed to violate these system rules. Assuming 7s block times, validators offline for approximately 9.25 consecutive hours (instead of ~18.5 hours) will be slashed 0.5% (instead of 0.01%).",
  "changes": [
    {
      "subspace": "slashing",
      "key": "SlashFractionDowntime",
      "value": 0.005000000000000000
    }
{
      "subspace": "slashing",
      "key": "SignedBlocksWindow",
      "value": 5000
    }
  ],
  "deposit": "512000000uatom"
}
```
**Note:** in the JSON file, `\n` creates a new line.

It's worth noting that this example proposal doesn't provide reasoning/justification for these changes. Consider consulting the [parameter-change best practices documentation](./best_practices.md) for guidance on the contents of a parameter-change proposal.
