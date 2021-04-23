<!--
order: 5
-->

# Service Providers

'Service Providers' are defined as entities providing services for end-users that involve some form of interaction with the Cosmos Hub. More specifically, this document will be focused around interactions with tokens.

This section does not concern wallet builders that want to provide Light-Client functionalities. Service Providers are expected to act as trusted point of contact to the blockchain for their end-users.

## Connection Options

There are four main technologies to consider, connecting to the Cosmos Hub:

- Full Nodes: To interact with the blockchain. 
- REST Server: This acts as a relayer for HTTP calls.
- REST API: Define available endpoints for the REST Server.
- GRPC: Connect to the Cosmos Hub via gRPC.

## Running a Full-Node

### What is a Full Node?

A Full Node is a network node that syncs up with the state of the blockchain. It provides blockchain data to others by using RESTful APIs, a replica of the database by exposing data with interfaces. A Full Node keeps in syncs with the rest of the blockchain nodes and stores the state on disk. If the full node does not have the queried block on disk the full node can go find the blockchain where the queried data lives. 

### Installation and configuration

We will describe the steps to run and interact with a Full Node for the Cosmos Hub.

First, you need to [install the software](../gaia-tutorials/installation.md).

Consider running your own [Cosmos Hub Full Node](../gaia-tutorials/join-mainnet.md).

## Command-Line interface

The Command-Line Interface (CLI) is the most powerful tool to access the Cosmos Hub and use gaia.
You need to install the latest version of `gaia` on your machine in order to use the Command-Line Interface.

Compare your version with the [latest release version](https://github.com/cosmos/gaia/releases)

```bash
gaiad version --long
```

#### Help

All available CLI commands will be shown if you just execute `gaiad`:

```bash
gaiad 
```

```bash
Stargate Cosmos Hub App

Usage:
  gaiad [command]

Available Commands:


  add-genesis-account Add a genesis account to genesis.json
  collect-gentxs      Collect genesis txs and output a genesis.json file
  debug               Tool for helping with debugging your application
  export              Export state to JSON
  gentx               Generate a genesis tx carrying a self delegation
  help                Help about any command
  init                Initialize private validator, p2p, genesis, and application configuration files
  keys                Manage your application's keys
  migrate             Migrate genesis to a specified target version
  query               Querying subcommands
  start               Run the full node
  status              Query remote node for status
  tendermint          Tendermint subcommands
  testnet             Initialize files for a simapp testnet
  tx                  Transactions subcommands
  unsafe-reset-all    Resets the blockchain database, removes address book files, and resets data/priv_validator_state.json to the genesis state
  validate-genesis    validates the genesis file at the default location or at the location passed as an arg
  version             Print the application binary version information

Flags:
  -h, --help                help for gaiad
      --home string         directory for config and data (default "/Users/tobias/.gaia")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "gaiad [command] --help" for more information about a command.
```

For each displayed command, you can use the `--help` flag to get further information. 

### Remote Access to gaiad

When choosing to remote access a Full Node and gaiad, you need a Full Node running and gaia installed on your local machine.

`gaiad` is the tool that enables you to interact with the node that runs on the Cosmos Hub network, whether you run it yourself or not.

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

Next you will learn useful CLI commands to interact with the Full Node.
You can run these commands as remote control or when you are running it on your local machine.

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

If you want to add a key to your keyring that imports a mnemonic, use the `--recover` flag.

```bash
gaiad keys add <your_key_name> --recover
```

#### Check your balance

After receiving tokens to your address, you can view your account by typing:

```bash
gaiad query account <YOUR_ADDRESS>
```

Query the account balance with the command:

```bash
gaiad query bank balances <YOUR_ADDRESS>
```

The response contains keys `balances` and `pagination`.
Each `balances` entry contains an `amount` held, connected to a `denom` identifier.
The typical $ATOM token is identified by the denom `uatom`. Where 1 `uatom` is 0.000001 ATOM.

```bash
balances: 
- amount: "12345678"
  denom: uatom
pagination:
  next_key: null
  total: "0"
```

When you query an account that has not received any token yet, you will see the `balances` entry as empty array.

```bash
balances: []
pagination:
  next_key: null
  total: "0"
```

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
- `--gas-prices`: This flag allows you to specify gas prices you pay for the transaction. The format is used as `0.025uatom`

## REST API

The [REST API documents](https://cosmos.network/rpc/) list all the available endpoints that you can use to interact
with your Full Node. Learn [how to enable the REST API](../gaia-tutorials/join-mainnet.md#enable-the-rest-api) on your Full Node.

### Listen for incoming transaction

The recommended way to listen for incoming transaction is to periodically query the blockchain through the following http endpoint:

[`/cosmos/bank/v1beta1/balances/{address}`](https://cosmos.network/rpc/)

## Cosmos SDK Transaction Signing

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

Before signing, all keys are lexicographically sorted and all white space are
removed from the JSON output.

The signature encoding is the 64-byte concatenation of ECDSArands (i.e. `r || s`),
where `s` is lexicographically less than its inverse in order to prevent malleability.
This is similar to Ethereum signing, but without the extra byte for PubKey recovery, since
Tendermint assumes the PubKey is always provided.

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

POST [`/txs`](https://cosmos.network/rpc/)