# Formatting a Community Pool Spend Proposal

There are five (5) components:
1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Recipient** - the Cosmos Hub (bech32-based) address that will receive funding from the Community Pool
4. **Amount** - the amount of funding that the recipient will receive in micro-ATOMs (uatom)
5. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Simple example
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

### Real example
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
