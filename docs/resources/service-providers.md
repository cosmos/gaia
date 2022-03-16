<!--
order: 5
-->

# Service Providers

'Service Providers' are defined as entities that provide services for end-users that involve some form of interaction with the Cosmos Hub. More specifically, this document is focused on interactions with tokens.

Service Providers are expected to act as trusted points of contact to the blockchain for their end-users. This Service Providers section does not apply to wallet builders that want to provide Light Client functionalities. 

This document describes:

- [Connection Options](#connection-options)
- [Running a Full Node](#running-a-full-node)
  - [What is a Full Node?](#what-is-a-full-node)
  - [Installation and Configuration](#installation-and-configuration)
- [Command-Line Interface](#command-line-interface)
  - [Available Commands](#available-commands)
  - [Remote Access to gaiad](#remote-access-to-gaiad)
  - [Create a Key pair](#create-a-key-pair)
  - [Check your Account](#check-your-account)
  - [Check your Balance](#check-your-balance)
  - [Send coins using the CLI](#send-coins-using-the-cli)
- [REST API](#rest-api)
  - [Listen for incoming transactions](#listen-for-incoming-transaction)


## Connection Options

There are four main technologies to consider to connect to the Cosmos Hub:

- Full Nodes: Interact with the blockchain. 
- REST Server: Serves for HTTP calls.
- REST API: Use available endpoints for the REST Server.
- GRPC: Connect to the Cosmos Hub using gRPC.

## Running a Full Node

### What is a Full Node?

A Full Node is a network node that syncs up with the state of the blockchain. It provides blockchain data to others by using RESTful APIs, a replica of the database by exposing data with interfaces. A Full Node keeps in syncs with the rest of the blockchain nodes and stores the state on disk. If the full node does not have the queried block on disk the full node can go find the blockchain where the queried data lives. 

### Installation and Configuration

This section describes the steps to run and interact with a full node for the Cosmos Hub.

First, you need to [install the software](../getting-started/installation.md).

Consider running your own [Cosmos Hub Full Node](../hub-tutorials/join-mainnet.md).

## Command-Line Interface

The command-line interface (CLI) is the most powerful tool to access the Cosmos Hub and use gaia.
To use the CLI, you must install the latest version of `gaia` on your machine.

Compare your version with the [latest release version](https://github.com/cosmos/gaia/releases)

```bash
gaiad version --long
```

### Available Commands

All available CLI commands are shown when you run the `gaiad` command:

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

```bash
gaiad query --help
Usage:
  gaiad query [flags]
  gaiad query [command]

Aliases:
  query, q

Available Commands:
  account                  Query for account by address
  auth                     Querying commands for the auth module
  bank                     Querying commands for the bank module
  block                    Get verified data for a the block at given height
  distribution             Querying commands for the distribution module
  evidence                 Query for evidence by hash or for all (paginated) submitted evidence
  gov                      Querying commands for the governance module
  ibc                      Querying commands for the IBC module
  ibc-transfer             IBC fungible token transfer query subcommands
  mint                     Querying commands for the minting module
  params                   Querying commands for the params module
  slashing                 Querying commands for the slashing module
  staking                  Querying commands for the staking module
  tendermint-validator-set Get the full tendermint validator set at given height
  tx                       Query for a transaction by hash in a committed block
  txs                      Query for paginated transactions that match a set of events
  upgrade                  Querying commands for the upgrade module

Flags:
      --chain-id string   The network chain ID
  -h, --help              help for query

Global Flags:
      --home string         directory for config and data (default "/Users/tobias/.gaia")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "gaiad query [command] --help" for more information about a command.
```

### Remote Access to gaiad

When choosing to remote access a Full Node and gaiad, you need a Full Node running and gaia installed on your local machine.

`gaiad` is the tool that enables you to interact with the node that runs on the Cosmos Hub network, whether you run it yourself or not.

To set up `gaiad` on a local machine and connect to an existing full node, use the following command:

```bash
gaiad config <flag> <value>
```

First, set up the address of the full node you want to connect to:

```bash
gaiad config node <host>:<port

// example: gaiad config node https://77.87.106.33:26657
```

If you run your own full node locally, use `tcp://localhost:26657` as the address. 

Set the default value of the `--trust-node` flag:

```bash
gaiad config trust-node false

// Set to true if you run a light client node
```

Finally, set the `chain-id` of the blockchain you want to interact with:

```bash
gaiad config chain-id cosmoshub-4
```

Next, learn to use CLI commands to interact with the full node.
You can run these commands as remote control or when you are running it on your local machine.

### Create a Key Pair

The default key is `secp256k1 elliptic curve`. Use the `gaiad keys` command to list the keys and generate a new key.



```bash
gaiad keys add <your_key_name>
```

You will be asked to create a password (at least 8 characters) for this key-pair. This will return the information listed below:

- `NAME`: Name of your key
- `TYPE`: Type of your key, always `local`. 
- `ADDRESS`: Your address. Used to receive funds.
- `PUBKEY`: Your public key. Useful for validators.
- `MNEMONIC`: 24-word phrase. **Save this mnemonic somewhere safe**. This phrase is required to recover your private key in case you forget the password. The mnemonic is displayed at the end of the output.

You can see all available keys by typing:

```bash
gaiad keys list
```

Use the `--recover` flag to add a key that imports a mnemonic to your keyring.

```bash
gaiad keys add <your_key_name> --recover
```

#### Check your Account

You can view your account by using the `query account` command.

```bash
gaiad query account <YOUR_ADDRESS>
```

It will display your account type, account number, public key and current account sequence.

```bash
'@type': /cosmos.auth.v1beta1.BaseAccount
account_number: "xxxx"
address: cosmosxxxx
pub_key:
  '@type': /cosmos.crypto.secp256k1.PubKey
  key: xxx
sequence: "x"
```

### Check your Balance

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

When you query an account that has not received any token yet, the `balances` entry is shown as an empty array.

```bash
balances: []
pagination:
  next_key: null
  total: "0"
```

#### Send Coins Using the CLI

To send coins using the CLI:

```bash
gaiad tx send <from_key_or_address> <to_address> <amount> \
    --chain-id=<your_chain_id> 
```

Parameters:

- `<from_key_or_address>`: Key name or address of sending account.
- `<to_address>`: Address of the recipient.
- `<amount>`: This parameter accepts the format `<value|coinName>`, such as `1000000uatom`.

Flags:

- `--chain-id`: This flag allows you to specify the id of the chain. There are different ids for different testnet chains and mainnet chains.
- `--gas-prices`: This flag allows you to specify the gas prices you pay for the transaction. The format is used as `0.0025uatom`

## REST API

The [REST API documents](https://cosmos.network/rpc/) list all the available endpoints that you can use to interact
with your full node. Learn [how to enable the REST API](../hub-tutorials/join-mainnet.md#enable-the-rest-api) on your full node.

### Listen for Incoming Transactions

The recommended way to listen for incoming transactions is to periodically query the blockchain by using the following HTTP endpoint:

[`/cosmos/bank/v1beta1/balances/{address}`](https://cosmos.network/rpc/)
