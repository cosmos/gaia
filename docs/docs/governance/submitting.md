---
title: Submitting a Proposal
order: 5
---

If you have a final draft of your proposal ready to submit, you may want to push your proposal live on the testnet first. These are the three primary steps to getting your proposal live on-chain.

Interacting with the Cosmos Hub via the command line in order to run queries or submit proposals has several prerequisites:
- You will need to compile [`gaiad`](../getting-started/installation) from source into a binary file executable by your operating system eg. MacOS, Windows, Linux
- You will need to indicate which chain you are querying, and currently this is `--chain-id cosmoshub-4`
- You will need to connect to a full node. You can find a list of available Cosmos Hub endpoints under the [API section](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json) in the [Chain Registry](https://github.com/cosmos/chain-registry).
- More info is in the Walkthrough Example section.

Running a full node can be difficult for those not technically-inclined, so you may choose to use a third-party's full node. In this case, the primary security risk is that of censorship: it's the single place where you have a single gateway to the network, and any messages submitted through an untrusted node could be censored.

## Hosting supplementary materials

In general we try to minimize the amount of data pushed to the blockchain. Hence, detailed documentation about a proposal is usually hosted on a separate censorship resistant data-hosting platform, like IPFS.

Once you have drafted your proposal, ideally as a Markdown file, you
can upload it to the IPFS network:

1. By [running an IPFS node and the IPFS software](https://ipfs.io), or
2. By using a service such as [https://pinata.cloud](https://pinata.cloud)

Ensure that you "pin" the file so that it continues to be available on the network. You should get a URL like this: https://ipfs.io/ipfs/QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD

The value `QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD` is called the `CID` of your file - it is effectively the file's hash.

If you uploaded a markdown file, you can use the IPFS markdown viewer to render the document for better viewing. Links for the markdown viewer look like `https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/<CID>`, where `<CID>` is your CID. For instance the link above would be: https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD

Share the URL with others and verify that your file is publicly accessible.

The reason we use IPFS is that it is a decentralized means of storage, making it resistant to censorship or single points of failure. This increases the likelihood that the file will remain available in the future.

## Formatting the JSON file for the governance proposal

Prior to sending the transaction that submits your proposal on-chain, you must create a JSON file. This file will contain the information that will be stored on-chain as the governance proposal. Begin by creating a new text (.txt) file to enter this information. Use [these best practices](./best-practices.md) as a guide for the contents of your proposal. When you're done, save the file as a .json file. 

Each proposal type is unique in how the JSON should be formatted.
See the relevant section for the type of proposal you are drafting:

- [Text Proposals](./formatting.md#text)
- [Community Pool Spend Proposals](./formatting.md#community-pool-spend)
- [Parameter Change Proposals](./formatting.md#parameter-change)

Once on-chain, most people will rely upon block explorers to interpret this information with a graphical user interface (GUI).

## Sending the transaction that submits your governance proposal

For information on how to use gaiad (the command line interface) to submit an on-chain proposal through the governance module, please refer to the [gaiad CLI tutorials](../hub-tutorials/gaiad) for the Cosmos Hub documentation.

### Proposal types

There are 2 proposal types that can be submitted to the CosmosHub governance module.

#### Legacy proposals (cosmos-sdk < v0.47)
These proposals can be submitted using `gaiad tx gov submit-legacy-proposal`.

Available proposals that can be submitted using this Tx are:
  * cancel-software-upgrade
  * change-reward-denoms
  * consumer-addition
  * consumer-removal
  * ibc-upgrade
  * param-change (does not work for standard cosmos-sdk modules, works on IBC and ICS modules)
  * software-upgrade
  * update-client

You can read more about submitting a legacy proposal in the [cosmos-sdk docs](https://docs.cosmos.network/v0.47/build/modules/gov#submit-legacy-proposal)

#### Proposals (cosmos-sdk >= v0.47)
These proposals can be submitted using `gaiad tx gov submit-proposal`.

Using `gaiad tx gov draft-proposal` can help prepare a proposal. The tool will create a file containing the specified proposal message and it also helps with populating all the required proposal fields.
You can always edit the file after you create it using `draft-proposal`

Most cosmos-sdk modules allow changing their governance gated parameters using a `MsgUpdateParams` which is a new way of updating governance parameters. It is important to note that `MsgUpdateParams` requires **all parameters to be specified** in the proposal message.

You can read more about submitting a proposal in the [cosmos-sdk docs](https://docs.cosmos.network/v0.47/build/modules/gov#submit-proposal)

#### Minimal Deposit amount
:::tip
Please note that cosmoshub-4 uses a minimum initial deposit amount.
:::

Proposals cannot be submitted successfully without providing a minimum initial deposit. In practice, this means that the `deposit` field in your proposal has to meet the `min_initial_deposit` governance parameter.
The minimum deposit is equal to `min_deposit * min_initial_deposit_ratio`. Only `uatom` is supported as deposit denom.
```shell
// checking the min_initial_deposit
gaiad q gov params -o json
{
   ...
   "params": {
      ...
      "min_deposit": [
         {
               "denom": "stake",
               "amount": "10000000"
         }
      ],
      "min_initial_deposit_ratio": "0.000000000000000000"
}
```


### Walkthrough example (changing x/staking params)

Let's illustrate how to change the `x/staking` parameters.

The module has the following parameters (values don't reflect actual on-chain values):
```shell
gaiad q staking params -o json
{
    "unbonding_time": "86400s",
    "max_validators": 100,
    "max_entries": 7,
    "historical_entries": 10000,
    "bond_denom": "stake",
    "min_commission_rate": "0.000000000000000000",
    "validator_bond_factor": "-1.000000000000000000",
    "global_liquid_staking_cap": "1.000000000000000000",
    "validator_liquid_staking_cap": "1.000000000000000000"
}
```

We will use `draft-proposal` to help us create a proposal file that we will later submit.
```shell
gaiad tx gov draft-proposal
// running the command will start a terminal applet allowing you to choose the proposal type

// 1st screen
Use the arrow keys to navigate: ↓ ↑ → ←
? Select proposal type:
    text
    community-pool-spend
    software-upgrade
    cancel-software-upgrade
  ▸ other // choose this

// 2nd screen
✔ other
Use the arrow keys to navigate: ↓ ↑ → ←
? Select proposal message type::
↑   /cosmos.staking.v1beta1.MsgUndelegate
  ▸ /cosmos.staking.v1beta1.MsgUpdateParams // choose this option
    /cosmos.staking.v1beta1.MsgValidatorBond
    /cosmos.upgrade.v1beta1.MsgCancelUpgrade
↓   /cosmos.upgrade.v1beta1.MsgSoftwareUpgrade
```

After choosing the `/cosmos.staking.v1beta1.MsgUpdateParams` message, the applet will allow you to set the message fields and some other proposal details.
Upon completion, the proposal will be available in the directory where you called the `gaiad` command inside the `draft_proposal.json` file.

Here is an example of the `draft_proposal.json` file:
```JSON
{
 "messages": [
  {
   "@type": "/cosmos.staking.v1beta1.MsgUpdateParams",
   "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
   "params": {
    "unbonding_time": "86400s",
    "max_validators": 100,
    "max_entries": 7,
    "historical_entries": 10000,
    "bond_denom": "uatom",
    "min_commission_rate": "0.050000000000000000",  // we are changing this from 0.000000000000000000
    "validator_bond_factor": "-1.000000000000000000",
    "global_liquid_staking_cap": "1.000000000000000000",
    "validator_liquid_staking_cap": "1.000000000000000000"
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "1000000uatom",
 "title": "Updating the staking params (min_comission_rate)",
 "summary": "This proposal will attempt to update the min_commission_rate staking parameter. During proposal creation and submission **all** proposal fields must be specified. Pay attention that you don't unintentionally specify different values for fields that you did not intend to change."
}
```


Finally, we submit the proposal:

```sh
gaiad tx gov submit-proposal <path_to_proposal.json>
   --from <submitter address> \
   --chain-id cosmoshub-4 \
   --gas <max gas allocated> \
   --fees <fees allocated> \
   --node <node address> \
```

Use `gaiad tx gov --help` to get more info about the CLI options, we will explain some options below:

1. `--from` is the account key that pays the transaction fee and deposit amount. This account key must be already saved in the keyring on your device and it must be an address you control (e.g. `--from hypha-dev-wallet`).
5. `--gas` is the maximum amount of gas permitted to be used to process the transaction (e.g. `--gas 500000`).
   - The more content there is in the description of your proposal, the more gas your transaction will consume
   - If this number isn't high enough and there isn't enough gas to process your transaction, the transaction will fail.
   - The transaction will only use the amount of gas needed to process the transaction.
6. `--fees` is a flat-rate incentive for a validator to process your transaction.
   - Many nodes use a minimum fee to disincentivize transaction spamming.
   - 7500uatom is equal to 0.0075 ATOM.
8. `--node` is using an established node to send the transaction to the Cosmos Hub 4 network. For available nodes, please look at the [Chain Registry](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json).

**Note**: be careful what you use for `--fees`. A mistake here could result in spending hundreds or thousands of ATOMs accidentally, which cannot be recovered.

### Verifying your transaction

After posting your transaction, your command line interface (gaiad) will provide you with the transaction's hash, which you can either query using gaiad or by searching the transaction hash using [Mintscan](https://www.mintscan.io/cosmos/txs/0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57). The hash should look something like this: `0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57`.

Alternatively, you can check your Tx status and information using:
```shell
gaiad q tx <hash>
```

### Troubleshooting a failed transaction

There are a number of reasons why a transaction may fail. Here are two examples:
1. **Running out of gas** - The more data there is in a transaction, the more gas it will need to be processed. If you don't specify enough gas, the transaction will fail.

2. **Incorrect denomination** - You may have specified an amount in 'utom' or 'atom' instead of 'uatom', causing the transaction to fail.

If you encounter a problem, try to troubleshoot it first, and then ask for help on the Cosmos Hub forum: [https://forum.cosmos.network](https://forum.cosmos.network). We can learn from failed attempts and use them to improve upon this guide.

### Depositing funds after a proposal has been submitted
Sometimes a proposal is submitted without having the minimum token amount deposited yet. In these cases you would want to be able to deposit more tokens to get the proposal into the voting stage. In order to deposit tokens, you'll need to know what your proposal ID is after you've submitted your proposal. You can query all proposals by the following command:

```sh
gaiad q gov proposals
```

If there are a lot of proposals on the chain already, you can also filter by your own address. For the proposal above, that would be:

```sh
gaiad q gov proposals --depositor cosmos1hxv7mpztvln45eghez6evw2ypcw4vjmsmr8cdx
```

Once you have the proposal ID, this is the command to deposit extra tokens:

```sh
gaiad tx gov deposit <proposal-id> <deposit_amount> --from <name>
```

The amount per deposit is equal to `min_deposit * min_deposit_ratio`. Only `uatom` is supported as deposit denom. Transactions where `deposit_amount < (min_deposit * min_deposit_ratio)` will be rejected.



### Submitting your proposal to the testnet
Submitting to the testnet is identical to mainnet submissions aside from a few changes:
1. The chain-id is `theta-testnet-001`.
2. The list of usable endpoints can be found [here](https://github.com/cosmos/testnets/tree/master/public#readme).
3. You will need testnet tokens, not ATOM. There is a faucet available in the Developer [Discord](https://discord.com/invite/cosmosnetwork).

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons:
1. To see what the proposal description will look like.
2. To signal that your proposal is about to go live on the mainnet.
3. To share what the proposal will look like in advance with stakeholders.
4. To test the functionality of the governance features.
