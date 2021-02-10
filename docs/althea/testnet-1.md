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

## Wait for it!

The rest of these steps are not ready to follow until the Feb 13th start of the testnet. There will be no node to sync with for example

## Bootstrapping steps and commands

We’re going to have a centralized start testnet. Where Althea will launch a chain, send everyone else tokens, and then each participant will come in and stake to become a validator.
In order to further simplify bootstrapping for this testnet we will be using pre-built binaries I am placing into a github release. These include ARM binaries for those of you on ARM platforms. Note that you will need to be running a 64bit ARM machine with a 64 bit operating system to use these binaries. In order to download ARM binaries change the names in the wget links from ‘client’ to ‘arm-client’. Repeat for all binaries

### Download the Gravity tools

```
mkdir gravity-tools
cd gravity-tools
# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/Testnet1/peggy

# Tools for the gravity bridge from the gravity repo
wget https://github.com/althea-net/gravity/releases/download/Testnet2/client
wget https://github.com/althea-net/gravity/releases/download/Testnet2/orchestrator
wget https://github.com/althea-net/gravity/releases/download/Testnet2/register-eth-key
wget https://github.com/althea-net/gravity/releases/download/Testnet2/relayer
chmod +x *
sudo mv * /usr/bin/

```

You may need to repeat this process to update your orchestrator throughout the testnet

### Generate your keys

Be sure to back up the phrase you get! You’ll need it in a bit

```
cd $HOME
peggy init mymoniker --chain-id althea-testnet1
peggy keys add validator
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/Testnet1/althea-testnet1-genesis.json
cp althea-testnet1-genesis.json $HOME/.althea/config/genesis.json
```

### Add persistent peers

Change the p2p.persistent_peers field in ~/.peggy/config/config.toml to contain the following:

```
persistent_peers = “737f401b6ed982bdd95568fd2232394a9c754a6a@testnet1.altheamesh.com:26657”
```

### Start your full node and wait for it to sync

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
  --amount=1500000stake \
  --pubkey=$(althea tendermint show-validator) \
  --moniker="put your validator name here" \
  --chain-id=althea-testnet1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
 --gas-adjustment=1.5 \
  --gas-prices="0.025stake" \
  --from=validator

```

Or if you need to change your stake. This is optional you only need to run the first command

```
althea tx staking create-validator \
  --amount=1500000stake \
  --pubkey=$(althea tendermint show-validator) \
  --moniker="put your validator name here" \
  --chain-id=peggy-testnet1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
 --gas-adjustment=1.5 \
  --gas-prices="0.025stake" \
  --from=validator

```

Finally to increase your stake (also optional)

```
althea keys show validator1 --bech val
althea tx staking delegate $(althea tendermint show-validator) 99000000stake --from validator1 --chain-id althea-testnet1 --fees 50stake --broadcast-mode block
```

### Confirm that you are validating

If you see one line in the response you are validating. If you don't see any output from this command you are not validating. Check that the last command ran successfully.

```
peggy query tendermint-validator-set | grep "$(peggy tendermint show-validator)"
```

### Setup Gravity bridge

You are now validating on the Althea blockchain. But as a validator you also need to run the Gravity bridge components or you will be slashed and removed from the validator set after about 16 hours.

### Edit your Validator node config to enable the RPC

Go to the line for api configuration and set enable=true, then restart your validator.

```
nano $HOME/.althea/config/app.toml
```

### Register your delegate keys

Save the keys that this generates as you will need them later

```
register-delegate-keys <todo update these args>
```

### Download
