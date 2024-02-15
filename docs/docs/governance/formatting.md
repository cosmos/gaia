---
title: Formatting a Proposal
order: 4
---

<!-- markdown-link-check-disable -->
Many proposals allow for long form text to be included, usually under the key `description`. These provide the opportunity to include [markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) if formatted correctly, as well as line breaks with `\n`. 

Beware, however, that if you are using the CLI to create a proposal, and setting `description` using a flag, the text will be [escaped](https://en.wikipedia.org/wiki/Escape_sequences_in_C) which may have undesired effects. 

Formatting a proposal can be a trial-and-error process, which is why first submitting to the [testnet](submitting.md#submitting-your-proposal-to-the-testnet) is recommended. 
<!-- markdown-link-check-enable -->

The examples shown below are of the text in a `json` file packaged into a `submit-proposal` transaction sent on-chain. More details about how to submit a proposal are in the [Submitting a Governance Proposal](./submitting.md) section, but for now just be aware that the examples are the contents of a file separate from the transaction. As a general rule, any flags specific to a proposal (e.g., Title, description, deposit, parameters, recipient) can be placed in a `json` file, while flags general to a transaction of any kind (e.g., chain-id, node-id, gas, fees) can remain in the CLI.

## Text

Text proposals are used by delegators to agree to a certain strategy, plan, commitment, future upgrade, or any other statement in the form of text. Aside from having a record of the proposal outcome on the Cosmos Hub chain, a text proposal has no direct effect on the change Cosmos Hub.

There are three components:

1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Text Proposal Example

[Proposal 12](https://www.mintscan.io/cosmos/proposals/12) asked if the Cosmos Hub community of validators charging 0% commission was harmful to the success of the Cosmos Hub.

```json
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

If the description says that a certain address will receive a certain number of ATOMs, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise). Check that the description aligns with the 'recipient' address.

### Community Pool Spend Example

The `amount` is `1000000000uatom`. 1,000,000 micro-ATOM is equal to 1 ATOM, so `recipient` address `cosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze` will receive 1000 ATOM if this proposal is passed.

The `deposit": "1000000uatom` results in 1 ATOM being used from the proposal submitter's account. 

```json
{
  "title": "Activate governance discussions on the Discourse forum using community pool funds",
  "description": "## Summary\nProposal to request for 1000 ATOM from the community spending pool to be sent to a multisig who will put funds towards stewardship of the Discourse forum to make it an authoritative record of governance decisions as well as a vibrant space to draft and discuss proposals.\n## Details\nWe are requesting 1000 ATOM from the community spending pool to activate and steward the Cosmos Hub (Discourse) forum for the next six months.\n\nOff-chain governance conversations are currently highly fragmented, with no shared public venue for discussing proposals as they proceed through the process of being drafted and voted on. It means there is no record of discussion that voters can confidently point to for context, potentially leading to governance decisions becoming delegitimized by stakeholders.\n\nThe requested amount will be sent to a multisig comprising individuals (members listed below) who can ensure that the tokens are spent judiciously. We believe stewardship of the forum requires:\n\n* **Moderation**: Format, edit, and categorize posts; Standardize titles and tags; Monitor and approve new posts; Archive posts.\n* **Facilitation**: Ask clarifying questions in post threads; Summarize discussions; Provide historical precedence to discussions.\n* **Engagement**: Circulate important posts on other social channels to increase community participation; Solicit input from key stakeholders.\n* **Guidance**: Orient and assist newcomers; Guide proposers through governance process; Answer questions regarding the forum or Cosmos ecosystem.\nThe work to steward the forum will be carried out by members of [Hypha Worker Co-op](https://hypha.coop/) and individuals selected from the community to carry out scoped tasks in exchange for ATOM from this budget.\n## Multisig Members\n* Hypha: Mai Ishikawa Sutton (Hypha Co-op)\n* Validator: Daniel Hwang (Stakefish)\n* Cosmos Hub developer: Lauren Gallinaro (Interchain Berlin)\n\nWe feel the membership of the multisig should be rotated following the six-month pilot period to preserve insight from the distinct specializations (i.e., Cosmos Hub validators and developers).\n## Timeline and Deliverables\nWe estimate the total work to take 250-300 hours over six months where we hope to produce:\n* **Moving summaries:** Provide succinct summaries of the proposals and include all publicly stated reasons why various entities are choosing to vote for/against a given proposal. These summaries will be written objectively, not siding with any one entity.\n* **Validator platforms:** Create a section of the Forum where we collate all validators' visions for Cosmos Hub governance to allow them to state their positions publicly. We will work with the smaller validators to ensure they are equally represented.\n* **Regular check-ins with the Cosmonaut DAO:** Collaborate with the future Cosmonaut DAO to ensure maximal accessibility and engagement. Community management is a critical, complementary aspect of increasing participation in governance.\n* **Announcement channel:** Create a read-only announcement channel in the Cosmos Community Discord, so that new proposals and major discussions can be easily followed.\n* **Tooling friendly posts:** Tag and categorize posts so that they can be easily ingested into existing tooling that validators have setup.\n* **Neutral moderation framework:** Document and follow transparent standards for how the forum is moderated.\n\nAt the end of the period, we will produce a report reflecting on our successes and failures, and recommendations for how the work of maintaining a governance venue can be continuously sustained (e.g., through a DAO). We see this initiative as a process of discovery, where we are learning by doing.\n\nFor more context, you can read through the discussions on this [proposal on the Discourse forum](https://forum.cosmos.network/t/proposal-draft-activate-governance-discussions-on-the-discourse-forum-using-community-pool-funds/5833).\n\n## Governance Votes\nThe following items summarize the voting options and what it means for this proposal:\n**YES** - You approve this community spend proposal to deposit 1000 ATOM to a multisig that will spend them to improve governance discussions in the Discourse forum.\n**NO** - You disapprove of this community spend proposal in its current form (please indicate why in the Cosmos Forum).\n**NO WITH VETO** - You are strongly opposed to this change and will exit the network if passed.\n**ABSTAIN** - You are impartial to the outcome of the proposal.\n## Recipient\ncosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze\n## Amount\n1000 ATOM\n\n***Disclosure**: Hypha has an existing contract with the Interchain Foundation focused on the testnet program and improving documentation. This work is beyond the scope of that contract and is focused on engaging the community in governance.*\n\nIPFS pin of proposal on-forum: (https://ipfs.io/ipfs/Qmaq7ftqWccgYCo8U1KZfEnjvjUDzSEGpMxcRy61u8gf2Y)",
  "recipient": "cosmos1xf2qwf6g6xvuttpf37xwrgp08qq984244952ze", 
  "amount": "1000000000uatom",
  "deposit": "1000000uatom"
}

```

## Param Change Proposal

**Note:** Changes to the [`gov` module](https://docs.cosmos.network/main/modules/gov) are different from the other kinds of parameter changes because `gov` has subkeys, [as discussed here](https://github.com/cosmos/cosmos-sdk/issues/5800). Only the `key` part of the JSON file is different for `gov` parameter-change proposals.

For parameter-change proposals, there are arguably seven (7) components, though three are nested beneath 'Changes':

1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Changes** - a component containing:
   1. **Subspace** - the Cosmos Hub module with the parameter that is being changed
   2. **Key** - the parameter that will be changed
   3. **Value** - the value of the parameter that will be changed by the governance mechanism
4. **Deposit** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

The components must be presented as shown in the example.

### Param Change Example

This example is 'real', because it was put on-chain using the Theta testnet and can be seen in the block explorer [here](https://explorer.theta-testnet.polypore.xyz/proposals/87).

Not all explorers will show the proposed parameter changes that are coded into the proposal, so ensure that you verify that the description aligns with what the governance proposal is programmed to enact. If the description says that a certain parameter will be increased, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).

```json
 {
  "title": "Doc update test: Param change for MaxValidators",
  "description": "Testing the proposal format for increasing the MaxValidator param",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 200
    }
  ],
  "deposit": "100000uatom"
}
```

## Software Upgrade Proposal

For software upgrade proposals, there are a number of items need to be decided before making a proposal.

1. **Network**. Testnet or Mainnet? When making a software upgrade proposal, you have to target a running network. This means either targeting the [testnet](https://github.com/cosmos/testnets) or [mainnet](https://github.com/cosmos/mainnet).
2. **Chain ID**. The correct chain ID should be included in the proposal. The chain ID for the Cosmos Hub mainnet is `cosmoshub-4`. Check the [testnet](https://github.com/cosmos/testnets) github for more information about the testnet setup.
3. **Upgrade block height**. The block height at which the upgrade needs to be calculated, bearing in mind the time required for the voting period and any extra time buffer.
4. **Binary URL's**. The proposal must have references to the binary that will be used for the upgrade. Validators may or may not use supplied binaries, typically, a number of validators will build from source and package the application and configuration as required by their operating environment.
5. **Upgrade Name**. The upgrade naming convention for mainnet is quite simple: `vXXX`, where `XXX` is the version number. The testnet maybe have different conventions.
6. **Proposer Account**. The proposer account is the Cosmos Hub account that is making the upgrade proposal to the network. A proposer Cosmos network address, is in the form `cosmos....`
7. **Node Address**. A running CosmosHub node, that is running on the selected network (mainnet, testnet), that will be used to submit the proposal. The required node can specified using the `--node flag`.

## Formatting the Proposal

When using the commandline, creating and formatting a proposal requires a number of steps. Online tools are also available, such as [Mintscan](https://www.mintscan.io/wallet/create-proposal) which forego some of this effort. Here we show proposal creation using the commandline.

1. Generate the unsigned proposal JSON using the `--offline` flag. Add a placeholder for the `description` text. A mainnet example from a [past upgrade proposal](https://www.mintscan.io/cosmos/proposals/825) is shown below.

   ```bash
   gaiad tx gov submit-proposal software-upgrade v13 \
   --title "v13 Software Upgrade" \
   --deposit 250000000uatom \
   --upgrade-height 17380000 \
   --upgrade-info '{"binaries":{"darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-darwin-amd64?checksum=sha256:58b9ca70f47376590aa4b4a9d75c69428c07bf6d0ad1481e669022bd470ec2c4","darwin/arm64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-darwin-arm64?checksum=sha256:c104c29c5af8e6da8308774f6fed4be3b09544a085c960359ca6ae0ec0045e5f","linux/amd64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-linux-amd64?checksum=sha256:ef34857554888de598edfa411d43d4dc0a820bad285cf044f5d8f91769d5599e","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-linux-arm64?checksum=sha256:114b085fb1bcd5a0de1aaf2d120ad0c102498ed3729cd31318f0baf865f5262a","windows/amd64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-windows-amd64.exe?checksum=sha256:87814b1072465cac7aa48c601bd11ea7563f900eaee15d38a17414cb642fc03d","windows/arm64":"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-windows-arm64.exe?checksum=sha256:c31c4700c073a76d18cfd85cf840a8b140ad1a95ff6788559d05ad3b4fba723c"}}' \
   --description "placeholder" \
   --fees 5000uatom \
   --from <proposer-address> \
   --chain-id cosmoshub-4 \
   --generate-only \
   --offline | jq -r '.' > unsigned-proposal.json
   ```

2. The proposal description needs to be written and formatted. The proposal description is in Markdown and it must be less than 10k characters in length.

3. Paste the Markdown for the proposal description into a tools such as https://www.cescaper.com/. This will convert the description into an escaped format that converts new lines into `\r\n`.

## Submitting the Proposal

`unsigned-v13-proposal.json`
```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.gov.v1beta1.MsgSubmitProposal",
        "content": {
          "@type": "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
          "title": "v13 Software Upgrade",
          "description": "# Gaia v13 Upgrade\r\n\r\n## Background\r\n\r\nThe **Gaia v13** release is a major release that will follow the standard governance process by initially submitting this post on the Cosmos Hub forum. After collecting forum feedback (~ 1 week) and adapting the proposal as required, a governance proposal will be sent to the Cosmos Hub for voting. The on-chain voting period typically lasts 2 weeks.\r\n\r\nOn governance vote approval, validators will be required to update the Cosmos Hub binary at the halt-height specified in the on-chain proposal.\r\n\r\n## Release Contents\r\n\r\nThe relevant Github epic for this release is: [Gaia v13 Release 3](https://github.com/cosmos/gaia/issues/2700)\r\n\r\nThere are two changes that are planned for the v13 release:\r\n\r\nThe **first change** is a housekeeping one, with the removal of the old Gravity Dex Liquidity module code from the Gaia codebase. The Gravity Dex Liquidity module was disabled earlier this year and any remaining funds have been returned to depositors or have been transferred to the CosmosHub community pool. The v13 release will now remove all redundant code related to this module from the Gaia codebase.\r\n\r\nThe **second change** is an update to the Interchain Security provider module for the Cosmos Hub.\r\n\r\nThe release candidate can be found [here](https://github.com/cosmos/gaia/releases/tag/v13.0.0-rc0). \r\nThe roadmap for this and future releases can be found [here](https://informal.systems/blog/informal-hub-team-jan-2023-update#2023-cosmos-hub-roadmap).\r\n\r\n\r\n\r\n## Testing and Testnets\r\n\r\nThe v13 release will go through rigorous testing, including e2e tests, integration tests, and differential tests. Differential tests are similar to integration tests, but they compare the system state to an expected state generated from a model implementation. In addition, v13 will be independently tested by the team at Hypha co-op.\r\n\r\nValidators and node operators can join a public testnet to participate in a test upgrade to a release candidate before the Cosmos Hub upgrades to the final release. You can find the relevant information (genesis file, peers, etc.) to join the release testnet (theta-testnet-001) [here](https://github.com/cosmos/testnets/tree/master/public), or the Replicated Security testnet (provider) [here](https://github.com/cosmos/testnets/blob/master/replicated-security/provider/README.md).\r\n\r\n## Potential risk factors\r\n\r\nAlthough very extensive testing and simulation will have taken place there always exists a risk that the Cosmos Hub might experience problems due to potential bugs or errors from the new features. In the case of serious problems, validators should stop operating the network immediately.\r\n\r\nCoordination with validators will happen in the #cosmos-hub-validators-verified channel of the Cosmos Network Discord to create and execute a contingency plan. Likely this will be an emergency release with fixes or the recommendation to consider the upgrade aborted and revert back to the previous release of gaia (v12).\r\n\r\n## Governance votes\r\n\r\nThe following items summarize the voting options and what it means for this proposal:\r\n\r\nYES - You agree that the Cosmos Hub should be updated with this release.\r\n\r\nNO - You disagree that the Cosmos Hub should be updated with this release.\r\n\r\nNO WITH VETO - A ‘NoWithVeto’ vote indicates a proposal either (1) is deemed to be spam, i.e., irrelevant to Cosmos Hub, (2) disproportionately infringes on minority interests, or (3) violates or encourages violation of the rules of engagement as currently set out by Cosmos Hub governance. If the number of ‘NoWithVeto’ votes is greater than a third of total votes, the proposal is rejected and the deposits are burned.\r\n\r\nABSTAIN - You wish to contribute to the quorum but you formally decline to vote either for or against the proposal.\r\n",
          "plan": {
            "name": "v13",
            "time": "0001-01-01T00:00:00Z",
            "height": "17380000",
            "info": "{\"binaries\":{\"darwin/amd64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-darwin-amd64?checksum=sha256:58b9ca70f47376590aa4b4a9d75c69428c07bf6d0ad1481e669022bd470ec2c4\",\"darwin/arm64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-darwin-arm64?checksum=sha256:c104c29c5af8e6da8308774f6fed4be3b09544a085c960359ca6ae0ec0045e5f\",\"linux/amd64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-linux-amd64?checksum=sha256:ef34857554888de598edfa411d43d4dc0a820bad285cf044f5d8f91769d5599e\",\"linux/arm64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-linux-arm64?checksum=sha256:114b085fb1bcd5a0de1aaf2d120ad0c102498ed3729cd31318f0baf865f5262a\",\"windows/amd64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-windows-amd64.exe?checksum=sha256:87814b1072465cac7aa48c601bd11ea7563f900eaee15d38a17414cb642fc03d\",\"windows/arm64\":\"https://github.com/cosmos/gaia/releases/download/v13.0.0/gaiad-v13.0.0-windows-arm64.exe?checksum=sha256:c31c4700c073a76d18cfd85cf840a8b140ad1a95ff6788559d05ad3b4fba723c\"}}",
            "upgraded_client_state": null
          }
        },
        "initial_deposit": [
          {
            "denom": "uatom",
            "amount": "250000000"
          }
        ],
        "proposer": "cosmos..."
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [],
    "fee": {
      "amount": [
        {
          "denom": "uatom",
          "amount": "5000"
        }
      ],
      "gas_limit": "2000000",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": []
}
```

5. Sign the proposal to generate `signed-v13-proposal.json`.
   ```bash
   gaiad tx sign unsigned-v13-proposal.json \
   --chain-id cosmoshub-4 \
   --from <cosmos-proposer-account> --node <node-address> &> signed-v13-proposal.json
   ```

5. Submit the `signed-v13-proposal.json` proposal.
   ```bash
   gaiad tx broadcast signed-v13-proposal.json --node <node-address> -b block
   ```
