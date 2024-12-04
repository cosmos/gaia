---
order: 2
parent:
  order: 1
---

# Text (Signaling)

## What are signaling proposals currently used for?
Signaling proposals are used to make an on-chain record of support or agreement on a certain topic or ideas. Text proposals do not contain any code. That is, they do not directly cause any changes to the Hub once passed.

Past signalling proposals have been used for a variety of reasons:
* Agreement to adopt (or not adopt) a feature in a future release ([7](https://www.mintscan.io/cosmos/proposals/7), [31](https://www.mintscan.io/cosmos/proposals/31),  [38](https://www.mintscan.io/cosmos/proposals/38), [49](https://www.mintscan.io/cosmos/proposals/49), [69](https://www.mintscan.io/cosmos/proposals/69))
* A high-signal alert to validators ([8](https://www.mintscan.io/cosmos/proposals/8))
* On-chain record of community opinion ([12](https://www.mintscan.io/cosmos/proposals/12))
* Ratification of a social norm ([75](https://www.mintscan.io/cosmos/proposals/75))

### A note on historical text proposals
In the early days of the Cosmos Hub, 'text' was the only proposal type. If you read old proposals, you will find 'text' proposals being used for things we use other proposal types now, such as changing a parameter ([10](https://www.mintscan.io/cosmos/proposals/10)) or upgrading the software ([19](https://www.mintscan.io/cosmos/proposals/19)).

The process for these historical proposals was that an on-chain signal was used to give permission for development or changes to be made off-chain and included in the Cosmos Hub code. With the addition of new proposal types, these development or spending choices can now be executed by the Gaia code immediately after the vote is tallied.

## Why make a signaling proposal?
Signaling proposals are a great way to take an official, public poll of community sentiment before investing more resources into a project. The most common way for text proposals to be used is to confirm that the community is actually interested in what the proposer wants to develop, without asking for money to fund development that might not be concrete enough to have a budget yet. 

Because the results of signaling proposals remain on-chain and are easily accessible to anyone, they are also a good way to formalize community opinions. Information contained in documentation or Github repos can be hard to find for new community members but signaling proposals in a block explorer or wallet is very accessible. 

You might make a signaling proposal to gather opinions for work you want to do for the Hub, or because you think it's important to have a record of some perspective held by the community at large. 

## What happens when a signaling proposal passes?
Technically, nothing happens on-chain. No code executes, and this 'unenforceable' property of text proposals is one of the biggest criticisms of the format. Regardless of whether the results of a signaling proposal are enforced by code, there can still be value from having a proposal on-chain and subject to discussion. Whether a proposal passes or fails, we all get information from it having been considered.

* The community might have had a thorough, thoughtful discussion about a topic that they otherwise wouldn't have had.
* A dev team interested in a feature might have a better idea of how their work will be received by the community.
* The community might be more informed about a topic than they previously were.
* The community might feel confident that we are aligned on a particular definition or social norm. 

## Submitting a text proposal

Follow the instructions below to create a text proposal and submit it to the blockchain.

```sh
➜ gaiad tx gov draft-proposal

Use the arrow keys to navigate: ↓ ↑ → ←
? Select proposal type:
  ▸ text  # choose this
    community-pool-spend
    software-upgrade
    cancel-software-upgrade
    other
```

Choose `text` from the `draft-proposal` menu and populate all the available fields.
```sh
✔ text
Enter proposal title: Title
Enter proposal authors: Author
Enter proposal summary: Proposal summary
Enter proposal details: Details, all the details
Enter proposal forum url: /
Enter proposal vote option context: Vote yes if <...>
Enter proposal deposit: 100001uatom
```

Check `draft_proposal.json`, your result should be similar to this:
```json
{
 "metadata": "ipfs://CID",
 "deposit": "100001uatom",
 "title": "Title",
 "summary": "Proposal summary"
}
```

Upload your `draft_metadata.json` to a distribution platform of your choice. `draft_proposal.json` is used to submit a governance proposal using `submit-proposal`.

```sh
gaiad tx gov submit-proposal <path_to_proposal.json>
   --from <submitter address> \
   --chain-id cosmoshub-4 \
   --gas <max gas allocated> \
   --fees <fees allocated> \
   --node <node address> \
```

Additional instructions with debugging information is available on the [submitting](../submitting.md) page.
