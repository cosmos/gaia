# mAuth module

## Introduction to Interchain Accounts
**Interchain Accounts** (ICA) is a standard that allows an account on a *controller* chain to create and securely control an address on a different *host* chain using the Inter Blockchain Protocol (IBC). Transactions that are native to the host chain are wrapped inside an IBC packet and get sent from the Interchain Account on the controller chain, to be executed on the host chain. 

The benefit of ICA is that there is no need to create a custom IBC implementation for each of the unique transactions that a sovereign blockchain might have (trading on a DEX, executing a specific smart contract, etc). Instead, a **generic** implementation allows blockchains to speak to each other, much like contracts can interact on Ethereum or other smart contract platforms.

For example, let’s say that you have an address on the Cosmos Hub (the controller) with OSMO tokens that you wanted to stake on Osmosis (the host). With Interchain Accounts, you can create and control a new address on Osmosis, without requiring a new private key. After sending your tokens to your Interchain Account using a regular IBC token transfer, you can send a wrapped `delegate` transaction over IBC which will then be unwrapped and executed natively on Osmosis.

## The mAuth module
Blockchains implementing Interchain Accounts can decide which messages they allow a controller chain to execute via a whitelist. The **mAuth module** whitelists all message types available to the Cosmos Hub, allowing any account on a controller chain to interact with the Cosmos Hub as if owning a native account on the chain itself.

In the following tutorial, we will demonstrate how to use Interchain Accounts through the [mAuth module](https://github.com/cosmos/gaia/tree/ica-acct-auth/x/mauth) inside the Gaia binary.

## Setup preparation
We will run two Cosmos-SDK chains (controller chain: `test-1` and host chain: `test-2`), and a relayer to connect these two chains. We will create an account on chain `test-1` and call it `a`, and register an Interchain Account (that we'll call `aica`) for `a` on chain `test-2`. We will create a normal account `b` on chain `test-2` as well. 

Through these 3 account, we can test if:
- `a` can control its `aica` to transfer tokens to account `b` on chain `test-2`.
- `a` can control its `aica` to tranfer `aica`'s token back to `a` using a regular IBC token transfer.

### Prepare to run two chains
We've simplified the setup process via several shell scripts. If you'd like to learn more about what's happening under the hood we suggest you inspect the files more closely.

Set up the two chains, create the keys for `a` and `b`, and start running both chains:
```shell
sh init_test_1.sh
sh init_test_2.sh
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
git checkout v0.9.0
cargo build --release --no-default-features --bin hermes
# binary path: ./target/release/hermes
cp  ./target/release/hermes $HOME/.cargo/bin
```

#### Create the IBC connection and start Hermes
```shell
source hermes/rly-variables.sh
sh hermes/rly-restore-keys.sh
sh hermes/rly-create-conn.sh
sh hermes/rly-start.sh
```

Now, you are running two Gaia (Cosmos Hub) blockchains in the background, as well as a Hermes relayer. Open a new terminal and config the Gaia home directories for the two chains.
```shell
sh gaia-home-config.sh
```

You can use the `--home` flag to let Gaia know which folder it should use for its configuration and data. This way, you can interact with multiple blockchains at once, using the same binary. Check your two Gaia home configs:
```shell
gaiad config --home test-1
{
	"chain-id": "test-1",
	"keyring-backend": "test",
	"output": "text",
	"node": "http://localhost:16657",
	"broadcast-mode": "sync"
}
```

```shell
 gaiad config --home test-2
{
	"chain-id": "test-2",
	"keyring-backend": "test",
	"output": "text",
	"node": "http://localhost:26657",
	"broadcast-mode": "sync"
}
```

## Trying out the Interchain Accounts functionality
Before you can get started you'll need to register an Interchain Account on `test-2` by sending an `icamsgauth register` command signed by `a` on the `test-1` chain:

```shell
gaiad tx icamsgauth register --from a --connection-id connection-0 --counterparty-connection-id connection-0 --gas 150000 --home test-1
```

To make things easier during the next steps, export the account addresses to environment variables:
```shell
export ICA_ADDR=$(gaiad query icamsgauth interchainaccounts $(gaiad keys show a -a --home test-1) connection-0 connection-0 --home test-1 -o json | jq -r '.interchain_account_address')
export A=$(gaiad keys show a -a --home test-1)
export B=$(gaiad keys show b -a --home test-2)
```

During the setup, `b` was given `100000000000stake` in one of the scripts, let's make sure `aica` also has some `stake`:
```
gaiad q bank balances $A --home test-2
# b's balance is 100000000000stake, send 1000stake to aica
gaiad tx bank send $B $AICA 1000stake --from b --home test-2
```

### Exercises
We would like to invite you to try to perform the actions below yourself. If you're having issues, you can find the solutions at the bottom of this tutorial.

> NOTE:
> * `a` = account on `test-1`
> * `aica` = account on `test-2` controlled by `a` on `test-1`
> * `b` = account on `test-2`

Q1: Let `a'` send `stake` to `b` (hint: using ICA)

Q2: Let `b` send `stake` back to `aica` (hint: via the Bank module)

Q3: Let `a` send `stake` to `b` (hint: via a regular IBC token transfer)

Q4: Let `b` send `ibc/stake` to `aica` (hint: via the Bank module)

Q5: Let `aica` send `ibc/stake` to `a` (hint: via ICA & IBC-Transfer)

<p align="center">
<img src="./ica.jpg" width="300" margin-left="auto"/>
</p>
<figcaption align = "center"><b>Fig.1 ICA exercise questions </b></figcaption>

### Solutions
#### Q1: `aica` sends tokens to `b` 
Both `aica` and `b` are on chain `test-2`, however, we need `a` from `test-1` to sign the transaction, because `a` is the only account with access to `aica` over ICA.

Step 1: generate the transaction json: 
```shell
gaiad tx bank send $AICA $B --chain-id test-2 --from a --gas 90000 --home test-1 --generate-only | jq '.body.messages[0]' > ./send-raw.json

cat send-raw.json
```

This will generate and display the following JSON file:
```shell
{
  "@type": "/cosmos.bank.v1beta1.MsgSend",
  "from_address": "cosmos13ys0vw7uhw5c70lrgzz6nw77f95k2pm42rt33areg33k0kltn2zsdjsfvu",
  "to_address": "cosmos13f43j8hyf63mhz4pkqkcnncpu7fvuxnmh9yxdw",
  "amount": [
    {
      "denom": "stake",
      "amount": "100"
    }
  ]
}
```

Step 2: send the generated transaction and let `a` sign it:
```shell
gaiad tx icamsgauth submit-tx ./send-raw.json --connection-id connection-0 --counterparty-connection-id connection-0 --from a --chain-id test-1 --gas 150000 --home test-1
```

#### Q2: `b` sends the tokens back to `aica`
Note that this transaction is just a regular coin transfer using the Bank module because both accounts exist on `test-2` and you are interacting directly with that chain via the `--home` flag.

```shell
gaiad tx bank send $B $AICA 100stake --home test-2
```

#### Q3: `a` sends tokens to `b` via IBC
Create a new IBC channel using Hermes:

```shell
hermes -c hermes/rly-config.toml create channel --port-a transfer --port-b transfer test-1 test-2
```

Kill the Hermes process and restart:
```shell
hermes -c hermes/rly-config.toml start
```

Initiate the IBC token transfer:
```shell
gaiad tx ibc-transfer transfer transfer channel-1 $B 200stake --from a --home test-1
```

IBC token transfers can take a while before they're confirmed. You can check the balance of `b` on `test-2`:
```shell
gaiad q bank balances $B --home test-2
balances:
- amount: "200"
  denom: ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9
- amount: "99999999000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

Note how the `200stake` received has changed it's denom to `ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9`. This is because a token that's sent over IBC always contains information about it's origin in its denom.

#### Q4: Let `b` send the `ibc/stake` it just received to `aica`
Notice how this is just a regular token transfer using the Bank module:
```shell
gaiad tx bank send $B $AICA 200ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9 --from b --home test-2
```

#### Q5: `aica` sends `100ibc/stake` to `a`

Create the channel for ibc transfer:
```shell
hermes create channel --port-a transfer --port-b transfer test-1 test-2
```

Step 1: prepare the transaction JSON file:

```shell
gaiad tx ibc-transfer transfer transfer channel-1 $A 100ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9 --from $AICA --home test-2 --generate-only | jq '.body.messages[0]' > send-raw.json
cat send-raw.json
```

This will generate and display the following JSON file:
```shell
{
  "@type": "/ibc.applications.transfer.v1.MsgTransfer",
  "source_port": "transfer",
  "source_channel": "channel-1",
  "token": {
    "denom": "ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9",
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
gaiad tx icamsgauth submit-tx send-raw.json --connection-id connection-1 --counterparty-connection-id connection-1 --from a --home test-1 --gas 150000
```

The long denom we saw will be changed from `ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9` back to `stake` when the token is back to a on chain `test-1`.

You might need to kill the Hermes process and restart to receive the tokens.

## References:
- [Hermes installation](https://hermes.informal.systems/installation.html)
- [Interchain Accounts tutorial](https://github.com/cosmos/interchain-accounts/blob/master/README.md)
