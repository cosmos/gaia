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

### Generate your keys

Be sure to back up the phrase you get! You’ll need it in a bit. If you don't back up the phrase here just follow the steps again to generate a new key.

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1
althea keys add validator
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea-testnet1-genesis.jsonn
cp althea-testnet1-genesis.json $HOME/.althea/config/genesis.json
```

### Add persistent peers

Change the p2p.persistent_peers field in ~/.althea/config/config.toml to contain the following:

```
persistent_peers = “<this value won't be available until Feb 13th>@testnet1.altheamesh.com:26657”
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

Finally to increase your stake (also optional)

```
althea keys show validator1 --bech val
althea tx staking delegate $(althea tendermint show-validator) 99000000stake --from validator1 --chain-id althea-testnet1 --fees 50stake --broadcast-mode block
```

### Confirm that you are validating

If you see one line in the response you are validating. If you don't see any output from this command you are not validating. Check that the last command ran successfully.

```
althea query tendermint-validator-set | grep "$(althea tendermint show-validator)"
```

### Setup Gravity bridge

You are now validating on the Althea blockchain. But as a validator you also need to run the Gravity bridge components or you will be slashed and removed from the validator set after about 16 hours.

### Edit your Validator node config to enable the RPC

Go to the line for api configuration and set enable=true, then restart your validator.

```
nano $HOME/.althea/config/app.toml
```

### Register your delegate keys

Delegate keys allow the for the validator private keys to be kept in secure storage while the Orchestrator can use it's own delegated keys for Gravity functions. The delegate keys registration tool will generate Ethereum and Cosmos keys for you if you don't provide any. Please save them as you will need them later.

```
register-delegate-keys --validator-phrase=<the phrase you saved earlier> --cosmos-rpc="http://localhost:1317"
```

### Download and setup Geth on the Groli testnet

We will be using Geth Ethereum light clients for this task. For production Gravity we suggest that you point your Orchestrator at a Geth light client and then configure your light client to peer with full nodes that you control. This provides higher reliability as light clients are very quick to start/stop and resync. Allowing you to for example rebuild an Ethereum full node without having days of Orchestrator downtime.

Please note that only Geth full nodes can serve Geth light clients, no other node type will do. Also you must configure a Geth full node to serve light client requests as they do not do so by default.

```
wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.9.25-e7872729.tar.gz
tar -xvf geth-linux-amd64-1.9.25-e7872729.tar.gz
cd geth-linux-amd64-1.9.25-e7872729
./geth --syncmode "light" --groli --http
```

### Deployment of the Gravity contract

Once 66% of the validator set has registered their delegate Ethereum key it is possible to deploy the Gravity Ethereum contract. Once deployed the Gravity contract address on Görli will be posted here

```
0xXXXXXXXXXXXXXXXXXXXXX
```

### Start your Orchestrator

Now that the setup is complete you can start your Orchestrator. Use the Cosmos mnemonic generated in the 'register delegate keys' step and the Ethereum private key also generated in that step. You should setup your Orchestrator in systemd or elsewhere to keep it running and restart it when it crashes. You will also need to send some Görli to the Orchestrator address (it will print the address corresponding to your private key on startup).

```
RUST_LOG=INFO orchestrator \
    --cosmos-phrase="{{COSMOS_MNEMONIC}}" \
    --ethereum-key="{{ETH_PRIV_KEY}}" \
    --cosmos-legacy-rpc="http://localhost:1317" \
    --cosmos-grpc="http://localhost:9090" \
    --ethereum-rpc="http://localhost:8545" \
    --fees=footoken \
    --contract-address="0xXXXXXXXXXXXXXXXXX"
```
