# icamauth module

## Introduction to Interchain Accounts
**Interchain Accounts** (ICA) is a standard that allows an account on a *controller* chain to create and securely control an address on a different *host* chain using the Inter Blockchain Protocol (IBC). Transactions native to the host chain are wrapped inside an IBC packet and sent from the Interchain Account Owner on the controller chain to be executed on the host chain.

The benefit of ICA is that there is no need to create a custom IBC implementation for each unique transaction that a sovereign blockchain might have (trading on a DEX, executing a specific smart contract, etc.). Instead, a **generic** implementation allows blockchains to speak to each other, much like contracts can interact on Ethereum or other smart contract platforms.

For example, let's say that you have an address on the Cosmos Hub (the controller) with OSMO tokens that you want to stake on Osmosis (the host). With Interchain Accounts, you can create and control a new address on Osmosis without requiring a new private key. After sending your tokens to your Interchain Account using a regular IBC token transfer, you can send a wrapped `delegate` transaction over IBC, which will then be unwrapped and executed natively on Osmosis.

## The icamauth module
Blockchains implementing Interchain Accounts can decide which messages they allow a controller chain to execute via a whitelist. The **icamuath (interchain account message authentication) module** whitelists most of the message types available to the Cosmos Hub, allowing any account on a controller chain to interact with the Cosmos Hub as if owning a native account on the chain itself.
query message types that are allowed on a host chain:
```shell
gaiad q interchain-accounts host params
```

The following tutorial will demonstrate how to use Interchain Accounts through the [icamauth module](../../../x/icamauth).

## Setup preparation
We will run two Cosmos-SDK chains (controller chain: `test-0` and host chain: `test-1`) and a relayer to connect these two chains. We will create an account on chain `test-0` and call it `alice`, and register an Interchain Account (that we'll call `alice_ica`)  on chain `test-1` for `alice` on chain `test-0`. We will also create a standard account, `bob` on chain `test-1`.

Through these 3 accounts, we can test if:
- `alice` on chain `test-0` can control its `alice_ica` to transfer tokens to account `bob` on chain `test-1`.
- `alice` can control its `alice_ica` to transfer `alice_ica`'s token back to `alice` using a regular IBC token transfer.

### Prepare to run two chains
We've simplified the setup process via several shell scripts. If you'd like to learn more about what's happening under the hood we suggest you inspect the files more closely.

Set up the two chains, create the keys for `alice` and `bob`, and start running both chains:
```shell
source ./docs/modules/icamauth/init_chain_controller.sh
source ./docs/modules/icamauth/init_chain_host.sh
```

### Setting up a Hermes relayer
#### Build the Hermes binary
Install Rust:
```shell
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

Build the Hermes binary:
```shell
git clone https://github.com/informalsystems/ibc-rs.git
cd ibc-rs
git checkout v1.0.0
cargo build --release --no-default-features --bin hermes
# binary path: ./target/release/hermes
cp  ./target/release/hermes $HOME/.cargo/bin
```

#### Create the IBC connection
Run the following command in `gaia/docs/modules/icamauth` directory to create an IBC connection:
```shell
cd ./docs/modules/icamauth
source hermes_setup.sh
```

## Testing the Interchain Accounts functionality
First of all, you need to register an Interchain Account on `test-1` for `alice` by sending an `icamauth register` command signed by `alice` on the `test-0` chain:

Open a new terminal and add the following variables.
```shell
export HOME0=$HOME/test-0
export HOME1=$HOME/test-1
```

```shell
gaiad tx icamauth register --from alice --connection-id connection-0 --gas-prices 0.03stake --home test-0 --home $HOME0
```
query alice's ica:
```shell
gaiad query icamauth interchainaccounts connection-0  $(gaiad keys show alice -a --home $HOME0) --home $HOME0
```
To make things easier during the next steps, export the account addresses to environment variables:
```shell
export ALICE_ICA=$(gaiad query icamauth interchainaccounts connection-0  $(gaiad keys show alice -a --home $HOME0) --home $HOME0 -o json | jq -r '.interchain_account_address')
export ALICE=$(gaiad keys show alice -a --home $HOME0)
export BOB=$(gaiad keys show bob -a --home $HOME1)
```

Let's make sure `alice_ica` has some `stake`:
```shell
gaiad q bank balances $ALICE_ICA --home $HOME1
gaiad tx bank send $BOB $ALICE_ICA 1000stake --from bob --gas-prices 0.025stake --home $HOME1
gaiad q bank balances $ALICE_ICA --home $HOME1
```

### Exercises
We would like to invite you to try to perform the actions below yourself. If you're having issues, you can find the solutions at the bottom of this tutorial.

> NOTE:
> * `alice` = account on `test-0`
> * `alice_ica` = account on `test-1` owned by `alice` on `test-0`
> * `bob` = account on `test-1`

Q1: Let `alice` send `stake` to `bob` (hint: using ICA)

Q2: Let `bob` send `stake` back to `alice_ica` (hint: via the Bank module)

Q3: Let `alice` send `stake` to `bob` (hint: via a regular IBC token transfer)

Q4: Let `bob` send `ibc/stake` to `alice_ica` (hint: via the Bank module)

Q5: Let `alice_ica` send `ibc/stake` to `alice` (hint: via ICA & IBC-Transfer)

### Solutions
#### Q1: `alice_ica` sends tokens to `bob` 
Both `alice_ica` and `bob` are on chain `test-1`, however, we need `alice` from `test-0` to sign the transaction, because `alice` is the only account with access to `alice_ica` over `icamuath`.

Step 1: generate the transaction json: 
```shell
gaiad tx bank send $ALICE_ICA $BOB --from alice 100stake --generate-only | jq '.body.messages[0]' > ./send-raw.json

cat send-raw.json
```

This will generate and display a JSON file similar to this following file:
```shell
{
  "@type": "/cosmos.bank.v1beta1.MsgSend",
  "from_address": "cosmos1g2c3l9m7zpvwsa2k4yx007zsnx9gme9qyw89uccxf7gkus6ehylsaklv2y",
  "to_address": "cosmos1jl3p6e62ey4xad8c5x0vh4p26j5ml8ejxr936t",
  "amount": [
    {
      "denom": "stake",
      "amount": "100"
    }
  ]
}
```

Step 2: send the generated transaction and let `alice` sign it:
```shell
gaiad tx icamauth submit ./send-raw.json --connection-id connection-0 --from alice --gas-prices 0.025stake --home $HOME0
```

#### Q2: `bob` sends the tokens back to `alice_ica`
Note that this transaction is just a regular coin transfer using the Bank module because both accounts exist on `test-1` and you are interacting directly with that chain via the `--home` flag.

```shell
gaiad tx bank send $BOB $ALICE_ICA 100stake --home $HOME1
```

#### Q3: `alice` sends tokens to `bob` via IBC
Create a new IBC channel using Hermes:

```shell
 hermes -c rly-config.toml create channel --a-chain test-0 --a-connection connection-0 --port-a transfer --port-b transfer
```

Initiate the IBC token transfer:
```shell
gaiad tx ibc-transfer transfer transfer channel-1 $BOB 200stake --from alice --gas-prices 0.025stake --home $HOME0
```

IBC token transfers can take a while before they're confirmed. You can check the balance of `bob` on `test-1`:
```shell
gaiad q bank balances $BOB --home $HOME1
balances:
- amount: "200"
  denom: ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9
- amount: "99999999000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

Note how the `200stake` received has changed its denom to `ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9`. Tokens sending over IBC always are encoded with information about its origin in its denom.

#### Q4: Let `bob` send the `ibc/stake` it just received to `alice_ica`
Notice how this is just a regular token transfer using the Bank module:
```shell
gaiad tx bank send $BOB $AICA_ICA 200ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 --from bob --gas-prices 0.025stake --home $HOME1
```

#### Q5: `alice_ica` sends `100ibc/stake` to `alice`

we have already created the channel in the above [#Q3], we can just use this channel to send the token back from `alice_ica` to `alice`. 

Step 1: prepare the transaction JSON file:

```shell
gaiad tx ibc-transfer transfer transfer channel-1 $ALICE 100ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 --from $ALICE_ICA --generate-only | jq '.body.messages[0]' > send-raw.json

cat send-raw.json
```

This will generate and display the following JSON file:
```shell
{
  "@type": "/ibc.applications.transfer.v1.MsgTransfer",
  "source_port": "transfer",
  "source_channel": "channel-1",
  "token": {
    "denom": "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
    "amount": "100"
  },
  "sender": "cosmos13ys0vw7uhw5c70lrgzz6nw77f95k2pm42rt33areg33k0kltn2zsdjsfvu",
  "receiver": "cosmos1sjww7vhxhe5sfye44fex4fv9telmuakuahk9nh",
  "timeout_height": {
    "revision_number": "1",
    "revision_height": "4130"
  },
  "timeout_timestamp": "1641572037493534000"
}
```

Step 2: use Interchain Accounts to execute the IBC transfer in the JSON file:
```shell
gaiad tx icamauth submit send-raw.json --connection-id connection-0 --from alice --home $HOME0 --gas-prices 0.025stake
```

The long denom we saw will be changed from `ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9` back to `stake` when the token is back to a on chain `test-0`.

## References:
- [Hermes installation](https://hermes.informal.systems/installation.html)
- [Interchain Accounts tutorial](https://github.com/cosmos/interchain-accounts/blob/master/README.md)
