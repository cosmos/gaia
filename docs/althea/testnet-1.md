# Althea Testnet 1

Althea Testnet #1 mainly focuses around the Gravity bridge integration. Our goal is to run a two week testnet covering.

- Althea chain parameter selection
- Gravity bridge slashing
- Gravity bridge relayer stability
- IBC testing with B-Harvest and Agoric

This testnet will be launched with a four hour Zoom call with all participants online.

prospective validators can [sign up here](https://airtable.com/shr86l8MZB7nLvjkH)

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

I also suggest an open notepad or other document to keep track of the keys you will be generating.

## Wait for it!

The rest of these steps are not ready to follow until the Feb 13th start of the testnet. There will be no node to sync with for example

## Bootstrapping steps and commands

We’re going to have a centralized start testnet. Where Althea will launch a chain, send everyone else tokens, and then each participant will come in and ualtg to become a validator.
In order to further simplify bootstrapping for this testnet we will be using pre-built binaries I am placing into a github release. These include ARM binaries for those of you on ARM platforms. Note that you will need to be running a 64bit ARM machine with a 64 bit operating system to use these binaries. In order to download ARM binaries change the names in the wget links from ‘client’ to ‘arm-client’. Repeat for all binaries

### Download Althea chain and the Gravity tools

```
mkdir althea-bin
cd althea-bin

# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea

# Tools for the gravity bridge from the gravity repo
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/client
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/orchestrator
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/register-delegate-keys
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/relayer
chmod +x *
sudo mv * /usr/bin/

```

At specific points during the testnet you may be told to 'update your orchestrator' or 'update your althea binary'. In order to do that you can simply repeat the above instructions and then restart the affected software.

### Generate your key

Be sure to back up the phrase you get! You’ll need it in a bit. If you don't back up the phrase here just follow the steps again to generate a new key.

Note 'validator' is just the name of your key here, you can pick anything you like, just remember it later.

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1
althea keys add validator
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea-testnet1-genesis.json
cp althea-testnet1-genesis.json $HOME/.althea/config/genesis.json
```

### Add persistent peers

Change the p2p.persistent_peers field in ~/.althea/config/config.toml to contain the following:

```
persistent_peers = “wait for it@testnet1.althea.net:26656”
```

### Start your full node and wait for it to sync

Ask what the current blockheight is in the chat

```
althea start
```

### Request some funds be sent to your address

Copy and paste your address into Zoom chat so that we can send you some tokens.

```
althea keys list
```

### Send your validator setup transaction

```
althea tx staking create-validator \
  --amount=100000000ualtg \
  --pubkey=$(althea tendermint show-validator) \
  --moniker="put your validator name here" \
  --chain-id=althea-testnet1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
 --gas-adjustment=1.5 \
  --gas-prices="0.025ualtg" \
  --from=validator

```

To increase your ualtg (ptional)

```
althea keys show validator1 --bech val
althea tx staking delegate $(althea tendermint show-validator) 99000000ualtg --from validator1 --chain-id althea-testnet1 --fees 50ualtg --broadcast-mode block
```

### Confirm that you are validating

If you see one line in the response you are validating. If you don't see any output from this command you are not validating. Check that the last command ran successfully.

```
althea query tendermint-validator-set | grep "$(althea tendermint show-address)"
```

### Setup Gravity bridge

You are now validating on the Althea blockchain. But as a validator you also need to run the Gravity bridge components or you will be slashed and removed from the validator set after about 16 hours.

### Edit your Validator node config to enable the RPC

In the app.toml edit the 'enable' for api to true. Here's an example of what it looks
like.

```
###############################################################################
###                           API Configuration                             ###
###############################################################################

[api]

# Enable defines if the API server should be enabled.
enable = true
```

```
nano $HOME/.althea/config/app.toml
```

### Register your delegate keys

Delegate keys allow the for the validator private keys to be kept in secure storage while the Orchestrator can use it's own delegated keys for Gravity functions. The delegate keys registration tool will generate Ethereum and Cosmos keys for you if you don't provide any. Please save them as you will need them later.

This call will be added to the Gravity cli before production to provide Ledger signing support.

```
RUST_LOG=INFO register-delegate-keys --validator-phrase=<the phrase you saved earlier> --cosmos-rpc="http://localhost:1317" --fees=footoken
```

### Fund your delegate keys

Both your Ethereum delegate key and your Cosmos delegate key will need some tokens to pay gas. On the Althea chain side you where sent some 'footoken' along with your ALTG. We're essentially using footoken as a gas token for this testnet.

You should have received 100 Althea Governance Token in uALTG and the same amount of footoken. We're going to send half to the delegate address

To get the address for your validator key you can run the below, where 'validator' is whatever you named your key in the 'generate your key' step.

```
althea keys show validator
```

```
althea tx bank send <your validator key> <your delegate cosmos address> 50000000footoken --chain-id=althea-testnet1
```

With the Althea side funded, now we need some Goerli Eth you can ask for some in chat or use [this faucet](https://goerli-faucet.slock.it/) for a small amount that should be more than sufficient for this testnet. Just paste in the Ethereum address that was generated in the previous step.

### Download and setup Geth on the Goerli testnet

We will be using Geth Ethereum light clients for this task. For production Gravity we suggest that you point your Orchestrator at a Geth light client and then configure your light client to peer with full nodes that you control. This provides higher reliability as light clients are very quick to start/stop and resync. Allowing you to for example rebuild an Ethereum full node without having days of Orchestrator downtime.

Please note that only Geth full nodes can serve Geth light clients, no other node type will do. Also you must configure a Geth full node to serve light client requests as they do not do so by default.

For the purposes of this testnet just follow the instructions below, even on the slowest node you should be synced inside of a few minutes.

```
wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.9.25-e7872729.tar.gz
tar -xvf geth-linux-amd64-1.9.25-e7872729.tar.gz
cd geth-linux-amd64-1.9.25-e7872729
./geth --syncmode "light" --goerli --http
```

### Deployment of the Gravity contract

Once 66% of the validator set has registered their delegate Ethereum key it is possible to deploy the Gravity Ethereum contract. Once deployed the Gravity contract address on Görli will be posted here

```
wait for it
```

### Start your Orchestrator

Now that the setup is complete you can start your Orchestrator. Use the Cosmos mnemonic generated in the 'register delegate keys' step and the Ethereum private key also generated in that step. You should setup your Orchestrator in systemd or elsewhere to keep it running and restart it when it crashes.

If your Orchestrator goes down for more than 16 hours during the testnet you will be slashed and booted from the active validator set.

Since you'll be running this a lot I suggest putting the command into a script, like so

```
nano start-orchestrator.sh
```

```
#!/bin/bash
RUST_LOG=INFO orchestrator \
    --cosmos-phrase="{{COSMOS_MNEMONIC}}" \
    --ethereum-key="{{ETH_PRIV_KEY}}" \
    --cosmos-legacy-rpc="http://localhost:1317" \
    --cosmos-grpc="http://localhost:9090" \
    --ethereum-rpc="http://localhost:8545" \
    --fees=footoken \
    --contract-address="wait for it"
```

```
bash start-orchestrator.sh
```

### Testing Peggy

Now that we've made it this far it's time to actually play around with the bridge
