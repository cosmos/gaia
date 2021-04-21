<!--
order: 5
-->

# Service Providers

We define 'Service Providers' as entities providing services for end-users that involve some form of interaction with the Cosmos Hub. More specifically, this document will be focused around interactions with tokens.

This section does not concern wallet builders that want to provide Light-Client functionalities. Service Providers are expected to act as trusted point of contact to the blockchain for their end-users.

## Connection Options

There are four main technologies to consider, connecting to the Cosmos Hub:

- Full Nodes: To interact with the blockchain. 
- Rest Server: This acts as a relayer for HTTP calls.
- Rest API: Define available endpoints for the Rest Server.
- GRPC: Connect to the Cosmos Hub via gRPC.

## Running a Full-Node

### What is a Full Node?

A Full Node is a network node that syncs up with the state of the blockchain. It provides blockchain data to others by using RESTful APIs, a replica of the database by exposing data with interfaces. A Full Node keeps in syncs with the rest of the blockchain nodes and stores the state on disk. If the full node does not have the queried block on disk the full node can go find the blockchain where the queried data lives. 

### Installation and configuration

We will describe the steps to run and interact with a Full Node for the Cosmos Hub.

First, you need to [install the software](../gaia-tutorials/installation.md).

Then, you can start running a [Cosmos Hub Full Node](../gaia-tutorials/join-mainnet.md).

### Command-Line interface

## Remote Access to gaiad

When choosing to remote access a Full Node and gaiad, you need a Full Node running.
You can either connect to an existing Full Node, or learn how to setup a [Cosmos Hub Full Node](../gaia-tutorials/join-mainnet.md).

::: warning
**Please check that you are always using the latest stable release of `gaiad`**
:::

`gaiad` is the tool that enables you to interact with the node that runs on the Cosmos Hub network, whether you run it yourself or not. Let us set it up properly.

In order to set up `gaiad` on a local machine and connect to an existing Full Node, use the following command:

```bash
gaiad config <flag> <value>
```

First, set up the address of the full-node you want to connect to:

```bash
gaiad config node <host>:<port

// example: gaiad config node https://77.87.106.33:26657
```

If you run your own Full Node locally, use `tcp://localhost:26657` as the address. 

Set the default value of the `--trust-node` flag:

```bash
gaiad config trust-node false

// Set to true if you run a light-client node
```

Finally, set the `chain-id` of the blockchain you want to interact with:

```bash
gaiad config chain-id cosmoshub-4
```

Next you will find a few useful CLI commands to interact with the Full-Node.

### How to create a key-pair

To generate a new key (default secp256k1 elliptic curve):

```bash
gaiad keys add <your_key_name>
```

You will be asked to create a password (at least 8 characters) for this key-pair. This will return the information listed below:

- `NAME`: Name of your key
- `TYPE`: Type of your key, always `local`. 
- `ADDRESS`: Your address. Used to receive funds.
- `PUBKEY`: Your public key. Useful for validators.
- `MNEMONIC`: 24-words phrase. **Save this mnemonic somewhere safe**. It is used to recover your private key in case you forget the password.

You can see all available keys by typing:

```bash
gaiad keys list
```

#### Check your balance

After receiving tokens to your address, you can view your account's balance by typing:

```bash
gaiad query account <YOUR_ADDRESS>
```

*Note: When you query an account balance with zero tokens, you will get this error: No account with address <YOUR_ADDRESS> was found in the state. This is expected! We're working on improving our error messages.*

#### Send coins using the CLI

Here is the command to send coins via the CLI:

```bash
gaiad tx send <from_key_or_address> <to_address> <amount> \
    --chain-id=<your_chain_id> 
```

Parameters:

- `<from_key_or_address>`: Key name or address of sending account.
- `<to_address>`: Address of the recipient.
- `<amount>`: This parameter accepts the format `<value|coinName>`, such as `10faucetToken`.

Flags:

- `--chain-id`: This flag allows you to specify the id of the chain. There will be different ids for different testnet chains and mainnet chains.

#### Help

If you need to do something else, the best command you can run is:

```bash
gaiad 
```

It will display all the available commands. For each command, you can use the `--help` flag to get further information. 

## REST API

The [REST API documents](https://cosmos.network/rpc/) list all the available endpoints that you can use to interact
with your Full Node.

To give more flexibility to developers, we have included the ability to
generate unsigned transactions, [sign](https://cosmos.network/rpc/#/ICS20/post_tx_sign)
and [broadcast](https://cosmos.network/rpc/#/ICS20/post_tx_broadcast) them with
different API endpoints. This allows service providers to use their own signing
mechanism for instance.

In order to generate an unsigned transaction (example with
[coin transfer](https://cosmos.network/rpc/#/ICS20/post_bank_accounts__address__transfers)),
you need to use the field `generate_only` in the body of `base_req`.

### Listen for incoming transaction

The recommended way to listen for incoming transaction is to periodically query the blockchain through the following endpoint of the LCD:

[`/cosmos/bank/v1beta1/balances/{address}`](https://cosmos.network/rpc/)

## Cosmos SDK Transaction Signing

Cosmos SDK transaction signing is a fairly simple process.

Every Cosmos SDK transaction has a canonical JSON representation. The `gaiad`
and Stargate REST interfaces provide canonical JSON representations of transactions
and their "broadcast" functions will provide compact Amino (a protobuf-like wire format)
encoding translations.

Things to know when signing messages:

The format is as follows

```json
{
  "account_number": XXX,
  "chain_id": XXX,
  "fee": XXX,
  "sequence": XXX,
  "memo": XXX,
  "msgs": XXX
}
```

The signer must supply `"chain_id"`, `"account number"` and `"sequence number"`.

The `"fee"`, `"msgs"` and `"memo"` fields will be supplied by the transaction
composer interface.

The `"account_number"` and `"sequence"` fields can be queried directly from the
blockchain or cached locally. Getting these numbers wrong, along with the chainID,
is a common cause of invalid signature error. You can load the mempool of a full
node or validator with a sequence of uncommitted transactions with incrementing
sequence numbers and it will mostly do the correct thing.  

Before signing, all keys are lexicographically sorted and all white space is
removed from the JSON output.

The signature encoding is the 64-byte concatenation of ECDSArands (i.e. `r || s`),
where `s` is lexicographically less than its inverse in order to prevent malleability.
This is like Ethereum, but without the extra byte for PubKey recovery, since
Tendermint assumes the PubKey is always provided anyway.

Signatures and public key examples in a signed transaction:

``` json
{
  "type": "auth/StdTx",
  "value": {
    "msg": [...],
    "signatures": [
      {
        "pub_key": {
          "type": "tendermint/PubKeySecp256k1",
          "value": XXX
        },
        "signature": XXX
      }
    ],
  }
}
```

Once signatures are properly generated, insert the JSON into into the generated
transaction and then use the broadcast endpoint.
