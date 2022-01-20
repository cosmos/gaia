# Mauth module

## Backgroud introduction
Mauth module allows all the message types from cosmos-sdk.

[insert introduction here]

In the following tutorial, we will practice how to use interchain account through [mauth module](https://github.com/cosmos/gaia/tree/ica-acct-auth/x/mauth) of gaia.
## Setup preparation
We will run two cosmos-sdk chains (control chain: `test-1` and host chain: `test-2`), and a relayer to connect these two chains. We will create an account on chain `test-1` and call it `a`, and register an interchain account(ica) for `a` on chain `test-2`, we will create a normal account `b` on chain `test-2` as well. Through these 3 account, we can test:
 - `a` can control `a`'s `ica` to transfer tokens to an account on chain `test-2`.
- `a` can control `a`'s `ica` to tranfer `ica`'s token to `a` (this also involves ibc transfer)

### Prepare to run two chains
set up the two chains:
```shell
sh init_test_1.sh
sh init_test_2.sh
```

### Set up hermes relayer
#### Build hermes binary
Install rust:
```shell
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

Build hermes binary:
```shell
git clone https://github.com/informalsystems/ibc-rs.git
cd ibc-rs
git checkout v0.9.0
cargo build --release --no-default-features --bin hermes
# binary path: ./target/release/hermes
cp  ./target/release/hermes $HOME/.cargo/bin
```
#### Create interchain connection and start hermes
```shell
source hermes/rly-variables.sh
sh hermes/rly-restore-keys.sh
sh hermes/rly-create-conn.sh
sh hermes/rly-start.sh
```
Now, you are running two cosmos-sdk chain in the background, and a hermes relayer. Open a new terminal and config gaia home for the two chains. Then you are ready to solve the [interchain account exercise](##Interchain account exercise).
```shell
sh gaia-home-config.sh
```
Check your two gaia home config:
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

## Interchain account exercise
### Questions
> NOTE:
> * `a` = account on `test-1`
> * `aica` = account on `test-2` controlled by `a` on `test-1`
> * `b` = account on `test-2`

Q1: `a'` sends `stake` to `b` (hint: via ICA)

Q2: `b` sends `stake` back to `aica` (hint: via Bank)

Q3: `a` sends `stake` to `b` (hint: via IBC-Transfer)

Q4: `b` sends `ibc/stake` to `aica` (hint: via Bank)

Q5: `aica` sends `ibc/stake` to `a` (hint: via ICA & IBC-Transfer)

<p align="center">
<img src="./ica.jpg" width="300" margin-left="auto"/>
</p>
<figcaption align = "center"><b>Fig.1 ICA exercise questions </b></figcaption>

### Register the interchain accounts
Register an interchain account on chain test-2 from `a` on chain test-1 and transfer some stakes to a's ica from b on chain test-2:
```shell
gaiad tx icamsgauth register --from a --connection-id connection-0 --counterparty-connection-id connection-0 --gas 150000 --home test-1

gaiad q bank balances $(gaiad keys show b -a --home test-2) --home test-2
# b's balance 100000000000stake, send 1000stake to aica
gaiad tx bank send $(gaiad keys show b -a --home test-2) $AICA 1000stake --from b --home test-2
```
get the a's ica (aica) address:
```shell
gaiad q icamsgauth interchainaccounts $(gaiad keys show a -a --home test-1) connection-0 connection-0 --home test-1
interchain_account_address: cosmos13ys0vw7uhw5c70lrgzz6nw77f95k2pm42rt33areg33k0kltn2zsdjsfvu
```
export account names:
```shell
export ICA_ADDR=$(gaiad query icamsgauth interchainaccounts $(gaiad keys show a -a --home test-1) connection-0 connection-0 --home test-1 -o json | jq -r '.interchain_account_address')
export A=$(gaiad keys show a -a --home test-1)
export B=$(gaiad keys show b -a --home test-2)
```


### Solutions
#### Q1: `aica` sends tokens to `b` 
Both `aica` and `b` are on chain test-2, however, we need `a` from `test-1` to sign the transaction. 

step 1: generate the transaction json: 
```shell
gaiad tx bank send $AICA $B --chain-id test-2 --from a --gas 90000 --home test-1 --generate-only | jq '.body.messages[0]' > ./send-raw.json

cat send-raw.json
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
step 2: send the generated transaction with `a` singing it:
```shell
gaiad tx icamsgauth submit-tx ./send-raw.json --connection-id connection-0 --counterparty-connection-id connection-0 --from a --chain-id test-1 --gas 150000 --home test-1
```


#### Q2: `b` sends the tokens back to `aica`

```shell
gaiad tx bank send $B $AICA 100stake --home test-2
```

#### Q3: `a` sends tokens to `b` via IBC
Create a hermes channel:

```shell
hermes -c hermes/rly-config.toml create channel --port-a transfer --port-b transfer test-1 test-2
```

kill hermes and restart
```shell
hermes -c hermes/rly-config.toml start
```

```shell
gaiad tx ibc-transfer transfer transfer channel-1 $B 200stake --from a --home test-1
```

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

#### Q4: `b` sends the ibc/stake to `aica`
```shell
gaiad tx bank send $B $AICA 200ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9 --from b --home test-2
```
#### Q5: `aica` sends 100ibc/stake to `a`
step 1: prepare the transaction json file
```shell
gaiad tx ibc-transfer transfer transfer channel-1 $A 100ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9 --from $AICA --home test-2 --generate-only | jq '.body.messages[0]' > send-raw.json 

 cat send-raw.json
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
step 2: use interchain account to execute the ibc transfer in the json file:
```shell
gaiad tx icamsgauth submit-tx send-raw.json --connection-id connection-1 --counterparty-connection-id connection-1 --from a --home test-1 --gas 150000
```
The ibc/stake will changed from the denom of `ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9` to `stake` when the token is back to a on chain test-1.
 You might need to kill hermes and restart to receive the tokens.

## References:
- [hermes installation](https://hermes.informal.systems/installation.html)
- [interchain account tutorial](https://github.com/cosmos/interchain-accounts/blob/master/README.md)
