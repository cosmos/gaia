---
order: 4
---

# Formatting a Proposal
<!-- markdown-link-check-disable -->
Many proposals allow for long form text to be included, usually under the key `description`. These provide the opportunity to include [markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) if formatted correctly, as well as line breaks with `\n`. Beware, however, that if you are using the CLI to create a proposal, and setting `description` using a flag, the text will be [escaped](https://en.wikipedia.org/wiki/Escape_sequences_in_C) which may have undesired effects. If you're using markdown or line breaks, you put the proposal text into a json file and include that file as part of the CLI proposal, as opposed to individual fields in flags.
<!-- markdown-link-check-enable -->

The examples shown below are simply of the text in a `json` file packaged into a `submit-proposal` transaction sent on-chain. More details about how to submit a proposal are in the [Submitting a Governance Proposal](./submitting.md) section.

## Text

Text proposals are used by delegators to agree to a certain strategy, plan, commitment, future upgrade, or any other statement in the form of text. Aside from having a record of the proposal outcome on the Cosmos Hub chain, a text proposal has no direct effect on the change Cosmos Hub.

There are three components:

1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Real example

[Proposal 12](https://hubble.figment.io/cosmos/chains/cosmoshub-4/governance/proposals/12) asked if the Cosmos Hub community of validators charging 0% commission was harmful to the success of the Cosmos Hub.

```
{
  "title": "Are Validators Charging 0% Commission Harmful to the Success of the Cosmos Hub?",
  "description": "This governance proposal is intended to act purely as a signalling proposal. Throughout this history of the Cosmos Hub, there has been much debate about the impact that validators charging 0% commission has on the Cosmos Hub, particularly with respect to the decentralization of the Cosmos Hub and the sustainability for validator operations. Discussion around this topic has taken place in many places including numerous threads on the Cosmos Forum, public Telegram channels, and in-person meetups. Because this has been one of the primary discussion points in off-chain Cosmos governance discussions, we believe it is important to get a signal on the matter from the on-chain governance process of the Cosmos Hub. There have been past discussions on the Cosmos Forum about placing an in-protocol restriction on validators from charging 0% commission. https://forum.cosmos.network/t/governance-limit-validators-from-0-commission-fee/2182 This proposal is NOT proposing a protocol-enforced minimum. It is merely a signalling proposal to query the viewpoint of the bonded Atom holders as a whole. We encourage people to discuss the question behind this governance proposal in the associated Cosmos Hub forum post here: https://forum.cosmos.network/t/proposal-are-validators-charging-0-commission-harmful-to-the-success-of-the-cosmos-hub/2505 Also, for voters who believe that 0% commission rates are harmful to the network, we encourage optionally sharing your belief on what a healthy minimum commission rate for the network using the memo field of their vote transaction on this governance proposal or linking to a longer written explanation such as a Forum or blog post. The question on this proposal is “Are validators charging 0% commission harmful to the success of the Cosmos Hub?”. A Yes vote is stating that they ARE harmful to the network's success, and a No vote is a statement that they are NOT harmful.",
  "deposit": "100000uatom"
}
```


## Community Pool Spend

There are five (5) components:

1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Recipient** - the Cosmos Hub (bech32-based) address that will receive funding from the Community Pool
4. **Amount** - the amount of funding that the recipient will receive in micro-ATOMs (uatom)
5. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Real example

In this simple example (below), a network explorer will list the governance proposal as "Community Pool Spend." When an observer selects the proposal, they'll see the description. Not all explorers will show the recipient and amount, so ensure that you verify that the description aligns with the what the governance proposal is programmed to enact. If the description says that a certain address will receive a certain number of ATOMs, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).

The `amount` is `1000000000uatom`. 1,000,000 micro-ATOM is equal to 1 ATOM, so `recipient` address `cosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze` will receive 1000 ATOM if this proposal is passed.

The `deposit": "1000000uatom` results in 1 ATOM being used from the proposal submitter's account. There is a minimum deposit required for a proposal to enter the voting period, and anyone may contribute to this deposit within a 14-day period. If the minimum deposit isn't reach before this time, the proposal will not enter voting period.

```
{
  "title": "Activate governance discussions on the Discourse forum using community pool funds",
  "description": "## Summary\nProposal to request for 1000 ATOM from the community spending pool to be sent to a multisig who will put funds towards stewardship of the Discourse forum to make it an authoritative record of governance decisions as well as a vibrant space to draft and discuss proposals.\n## Details\nWe are requesting 1000 ATOM from the community spending pool to activate and steward the Cosmos Hub (Discourse) forum for the next six months.\n\nOff-chain governance conversations are currently highly fragmented, with no shared public venue for discussing proposals as they proceed through the process of being drafted and voted on. It means there is no record of discussion that voters can confidently point to for context, potentially leading to governance decisions becoming delegitimized by stakeholders.\n\nThe requested amount will be sent to a multisig comprising individuals (members listed below) who can ensure that the tokens are spent judiciously. We believe stewardship of the forum requires:\n\n* **Moderation**: Format, edit, and categorize posts; Standardize titles and tags; Monitor and approve new posts; Archive posts.\n* **Facilitation**: Ask clarifying questions in post threads; Summarize discussions; Provide historical precedence to discussions.\n* **Engagement**: Circulate important posts on other social channels to increase community participation; Solicit input from key stakeholders.\n* **Guidance**: Orient and assist newcomers; Guide proposers through governance process; Answer questions regarding the forum or Cosmos ecosystem.\nThe work to steward the forum will be carried out by members of [Hypha Worker Co-op](https://hypha.coop/) and individuals selected from the community to carry out scoped tasks in exchange for ATOM from this budget.\n## Multisig Members\n* Hypha: Mai Ishikawa Sutton (Hypha Co-op)\n* Validator: Daniel Hwang (Stakefish)\n* Cosmos Hub developer: Lauren Gallinaro (Interchain Berlin)\n\nWe feel the membership of the multisig should be rotated following the six-month pilot period to preserve insight from the distinct specializations (i.e., Cosmos Hub validators and developers).\n## Timeline and Deliverables\nWe estimate the total work to take 250-300 hours over six months where we hope to produce:\n* **Moving summaries:** Provide succinct summaries of the proposals and include all publicly stated reasons why various entities are choosing to vote for/against a given proposal. These summaries will be written objectively, not siding with any one entity.\n* **Validator platforms:** Create a section of the Forum where we collate all validators' visions for Cosmos Hub governance to allow them to state their positions publicly. We will work with the smaller validators to ensure they are equally represented.\n* **Regular check-ins with the Cosmonaut DAO:** Collaborate with the future Cosmonaut DAO to ensure maximal accessibility and engagement. Community management is a critical, complementary aspect of increasing participation in governance.\n* **Announcement channel:** Create a read-only announcement channel in the Cosmos Community Discord, so that new proposals and major discussions can be easily followed.\n* **Tooling friendly posts:** Tag and categorize posts so that they can be easily ingested into existing tooling that validators have setup.\n* **Neutral moderation framework:** Document and follow transparent standards for how the forum is moderated.\n\nAt the end of the period, we will produce a report reflecting on our successes and failures, and recommendations for how the work of maintaining a governance venue can be continuously sustained (e.g., through a DAO). We see this initiative as a process of discovery, where we are learning by doing.\n\nFor more context, you can read through the discussions on this [proposal on the Discourse forum](https://forum.cosmos.network/t/proposal-draft-activate-governance-discussions-on-the-discourse-forum-using-community-pool-funds/5833).\n\n## Governance Votes\nThe following items summarize the voting options and what it means for this proposal:\n**YES** - You approve this community spend proposal to deposit 1000 ATOM to a multisig that will spend them to improve governance discussions in the Discourse forum.\n**NO** - You disapprove of this community spend proposal in its current form (please indicate why in the Cosmos Forum).\n**NO WITH VETO** - You are strongly opposed to this change and will exit the network if passed.\n**ABSTAIN** - You are impartial to the outcome of the proposal.\n## Recipient\ncosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze\n## Amount\n1000 ATOM\n\n***Disclosure**: Hypha has an existing contract with the Interchain Foundation focused on the testnet program and improving documentation. This work is beyond the scope of that contract and is focused on engaging the community in governance.*\n\nIPFS pin of proposal on-forum: (https://ipfs.io/ipfs/Qmaq7ftqWccgYCo8U1KZfEnjvjUDzSEGpMxcRy61u8gf2Y)",
  "recipient": "cosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze", 
  "amount": "1000000000uatom",
  "deposit": "1000000uatom"
}

```

## Params Change

**Note:** Changes to the [`gov` module](https://docs.cosmos.network/main/modules/gov/) are different from the other kinds of parameter changes because `gov` has subkeys, [as discussed here](https://github.com/cosmos/cosmos-sdk/issues/5800). Only the `key` part of the JSON file is different for `gov` parameter-change proposals.

For parameter-change proposals, there are seven (7) components:

1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Subspace** - the Cosmos Hub module with the parameter that is being changed
4. **Key** - the parameter that will be changed
5. **Value** - the value of the parameter that will be changed by the governance mechanism
6. **Denom** - `uatom` (micro-ATOM) will be the type of asset used as the deposit
7. **Amount** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Real example

In this example, a network explorer will list the governance proposal by its title: "STaking param change" When a user selects the proposal, they'll see the proposal’s description. 

Not all explorers will show the proposed parameter changes that are coded into the proposal, so ensure that you verify that the description aligns with what the governance proposal is programmed to enact. If the description says that a certain parameter will be increased, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).


The deposit `denom` is `uatom` and `amount` is `100000`. Since 1,000,000 micro-ATOM is equal to 1 ATOM, a deposit of 0.1 ATOM will be included with this proposal. The gaia-13007 testnet currently has a 0.1 ATOM minimum deposit, so this will put the proposal directly into the voting period. There is a minimum deposit required for a proposal to enter the voting period, and anyone may contribute to this deposit within a 14-day period. If the minimum deposit isn't reached before this time, the deposit amounts will be burned. Deposit amounts will also be burned if quorum isn't met in the vote or if the proposal is vetoed.


```json
{
  EXAMPLE MISSING
}
```


It's worth noting that this example proposal doesn't provide reasoning/justification for these changes. Consider consulting the [parameter-change best practices documentation](./best-practices.md) for guidance on the contents of a parameter-change proposal.
