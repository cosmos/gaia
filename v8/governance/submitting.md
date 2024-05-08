---
order: 5
---

# Submitting a Proposal

If you have a final draft of your proposal ready to submit, you may want to push your proposal live on the testnet first. These are the three primary steps to getting your proposal live on-chain.

Interacting with the Cosmos Hub via the command line in order to run queries or submit proposals has several prerequisites:
  - You will need to compile [`gaiad`](https://hub.cosmos.network/main/getting-started/installation.html) from source into a binary file executable by your operating system eg. MacOS, Windows, Linux
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

For information on how to use gaiad (the command line interface) to submit an on-chain proposal through the governance module, please refer to the [gaiad resource](../hub-tutorials/gaiad.md) for the Cosmos Hub documentation.

### Walkthrough example

This is the generic command format for using gaiad (the command-line interface) to submit your proposal on-chain:

```
gaiad tx gov submit-proposal <proposal type>\
   -- <json file> \
   --from <submitter address> \
   --deposit <deposit in uatom> \
   --chain-id <chain id> \
   --gas <max gas allocated> \
   --fees <fees allocated> \
   --node <node address> \

```

A specific example is given here:

```
gaiad tx gov submit-proposal community-pool-spend\
   --~/community_spend_proposal.json \
   --from hypha-dev-wallet \
   --deposit 1000000uatom \
   --chain-id cosmoshub-4 \
   --gas 500000 \
   --fees 7500uatom \
   --node https://rpc.cosmos.network:443 \

```


If `<proposal type>` is left blank, the type will be a Text proposal. Otherwise, it can be set to `param-change` or `community-pool-spend`. Use `--help` to get more info from the tool.


1. `gaiad` is the command-line interface client that is used to send transactions and query the Cosmos Hub.
2. `tx gov submit-proposal community-pool-spend` indicates that the transaction is submitting a community pool spend proposal.
3. `--~/community_spend_proposal.json` indicates the file containing the proposal details.
3. `--from hypha-dev-wallet` is the account key that pays the transaction fee and deposit amount. This account key must be already saved in the keyring on your device and it must be an address you control.
4. `--gas 500000` is the maximum amount of gas permitted to be used to process the transaction.
   - The more content there is in the description of your proposal, the more gas your transaction will consume
   - If this number isn't high enough and there isn't enough gas to process your transaction, the transaction will fail.
   - The transaction will only use the amount of gas needed to process the transaction.
5. `--fees` is a flat-rate incentive for a validator to process your transaction.
   - The network still accepts zero fees, but many nodes will not transmit your transaction to the network without a minimum fee.
   - Many nodes (including the Figment node) use a minimum fee to disincentivize transaction spamming.
   - 7500uatom is equal to 0.0075 ATOM.
6. `--chain-id cosmoshub-4` is Cosmos Hub 4. For current and past chain-id's, please look at the [cosmos/mainnet resource](https://github.com/cosmos/mainnet).
   - The testnet chain ID is `theta-testnet-001`. For current and past testnet information, please look at the [testnet repository](https://github.com/cosmos/testnets).
7. `--node https://rpc.cosmos.network:443` is using an established node to send the transaction to the Cosmos Hub 4 network. For available nodes, please look at the [Chain Registry](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json).

**Note**: be careful what you use for `--fees`. A mistake here could result in spending hundreds or thousands of ATOMs accidentally, which cannot be recovered.

### Verifying your transaction

After posting your transaction, your command line interface (gaiad) will provide you with the transaction's hash, which you can either query using gaiad or by searching the transaction hash using [Mintscan](https://www.mintscan.io/cosmos/txs/0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57). The hash should look something like this: `0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57`

### Troubleshooting a failed transaction

There are a number of reasons why a transaction may fail. Here are two examples:
1. **Running out of gas** - The more data there is in a transaction, the more gas it will need to be processed. If you don't specify enough gas, the transaction will fail.

2. **Incorrect denomination** - You may have specified an amount in 'utom' or 'atom' instead of 'uatom', causing the transaction to fail.

If you encounter a problem, try to troubleshoot it first, and then ask for help on the Cosmos Hub forum: [https://forum.cosmos.network](https://forum.cosmos.network). We can learn from failed attempts and use them to improve upon this guide.

### Depositing funds after a proposal has been submitted
Sometimes a proposal is submitted without having the minimum token amount deposited yet. In these cases you would want to be able to deposit more tokens to get the proposal into the voting stage. In order to deposit tokens, you'll need to know what your proposal ID is after you've submitted your proposal. You can query all proposals by the following command:

```
gaiad q gov proposals
```

If there are a lot of proposals on the chain already, you can also filter by your own address. For the proposal above, that would be:

```
gaiad q gov proposals --depositor cosmos1hxv7mpztvln45eghez6evw2ypcw4vjmsmr8cdx
```

Once you have the proposal ID, this is the command to deposit extra tokens:

```
gaiad tx gov deposit <proposal-id> <deposit> --from <name>
```

In our case above, the `<proposal-id>` would be 59 as queried earlier.
The `<deposit>` is written as `500000uatom`, just like the example above.

### Submitting your proposal to the testnet
Submitting to the testnet is identical to mainnet submissions aside from a few changes:
1. The chain-id is `theta-testnet-001`.
2. The list of usable endpoints can be found [here](https://github.com/cosmos/testnets/tree/master/public#readme).
3. You will need testnet tokens, not ATOM. There is a faucet available in the Developer [Discord](https://discord.gg/W8trcGV).

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons:
1. To see what the proposal description will look like.
2. To signal that your proposal is about to go live on the mainnet.
3. To share what the proposal will look like in advance with stakeholders.
4. To test the functionality of the governance features.


