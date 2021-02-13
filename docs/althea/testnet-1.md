# Althea Testnet 1

Althea Testnet #1 mainly focuses around the Gravity bridge integration. Our goal is to run a two week testnet covering.

- Althea chain parameter selection
- Gravity bridge slashing
- Gravity bridge Orchestrator stability
- IBC testing with B-Harvest and Agoric

This testnet will be launched with a four hour Zoom call with all participants online.

prospective validators can [sign up here](https://airtable.com/shr86l8MZB7nLvjkH)

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

I also suggest an open notepad or other document to keep track of the keys you will be generating.

## Warning

The genesis file and all binaries have been updated since yesterday, you should blow away your directory and start again if you did anything on the 12th.

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

Note 'myvalidatorkeyname' is just the name of your key here, you can pick anything you like, just remember it later.

You'll be prompted to create a password, I suggest you pick something short since you'll be typing it a lot

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1
althea keys add myvalidatorkeyname
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea-testnet1-genesis.json
cp althea-testnet1-genesis.json $HOME/.althea/config/genesis.json
```

### Add persistent peers

Change the p2p.persistent_peers field in ~/.althea/config/config.toml to contain the following:

```
persistent_peers = “05ded2f258ab158c5526eb53aa14d122367115a7@testnet1.althea.net:26656”
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
  --from=myvalidatorkeyname

```

To increase your ualtg (optional)

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

Once you save the phrase generated by this command you will have two phrases to keep track of. The one generated here is your 'delegate address' and the one you generated before is your 'validator address'.

```
RUST_LOG=INFO register-delegate-keys --validator-phrase="the phrase you saved earlier" --cosmos-rpc="http://localhost:1317" --fees=footoken
```

### Fund your delegate keys

Both your Ethereum delegate key and your Cosmos delegate key will need some tokens to pay gas. On the Althea chain side you where sent some 'footoken' along with your ALTG. We're essentially using footoken as a gas token for this testnet.

You should have received 100 Althea Governance Token in uALTG and the same amount of footoken. We're going to send half to the delegate address

To get the address for your validator key you can run the below, where 'myvalidatorkeyname' is whatever you named your key in the 'generate your key' step.

```
althea keys show myvalidatorkeyname
```

```
althea tx bank send myvalidatorkeyname <your delegate cosmos address> 50000000footoken --chain-id=althea-testnet1
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
0xB48095a68501bC157654d338ce86fdaEF4071B24
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
    --cosmos-phrase="your delegate key phrase" \
    --ethereum-key="your delegate ethereum private key" \
    --cosmos-legacy-rpc="http://localhost:1317" \
    --cosmos-grpc="http://localhost:9090" \
    --ethereum-rpc="http://localhost:8545" \
    --fees=footoken \
    --contract-address="0xB48095a68501bC157654d338ce86fdaEF4071B24"
```

```
bash start-orchestrator.sh
```

### Testing Gravity

Now that we've made it this far it's time to actually play around with the bridge

This first command will send some ERC20 tokens to an address of your choice on the Althea
chain. Notice that the Ethereum key is pre-filled. This address has both some test ETH and
a large balance of ERC20 tokens from the contracts listed here.

```
0xD7600ae27C99988A6CD360234062b540F88ECA43 - Bitcoin MAX (MAX)
0x7580bFE88Dd3d07947908FAE12d95872a260F2D8 - 2 Ethereum (E2H)
0xD50c0953a99325d01cca655E57070F1be4983b6b - Byecoin (BYE)
```

Note that the 'amount' field for this command is now in whole coins rather than wei like the previous testnets

```
RUST_LOG=info client eth-to-cosmos \
        --ethereum-key="0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7" \
        --ethereum-rpc="http://localhost:8545" \
        --contract-address="0xB48095a68501bC157654d338ce86fdaEF4071B24" \
        --erc20-address="any of the three values above" \
        --amount=1 \
        --cosmos-destination="any Cosmos address, I suggest your delegate Cosmos address"
```

You should see a message like this on your Orchestrator. The details of course will be different but it means that your Orchestrator has observed the event on Ethereum and sent the details into the Cosmos chain!

```
[2021-02-13T12:35:54Z INFO  orchestrator::ethereum_event_watcher] Oracle observed deposit with sender 0xBf660843528035a5A4921534E156a27e64B231fE, destination cosmos1xpfu40gseet70wfeazds773v05pjx3dwe7e03f, amount
999999984306749440, and event nonce 3
```

Once the event has been observed we can check our balance on the Cosmos side. We will see some peggy<ERC20 address> tokens in our balance. We have a good bit of code in flight right now so the module renaming from 'Peggy' to 'Gravity' has been put on hold until we're feature complete.

```
althea query bank balances <any cosmos address>
```

Now that we have some tokens on the Althea chain we can try sending them back to Ethereum. Remember to use the Cosmos phrase for the address you actually sent the tokens to.

```
RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the phrase containing the Gravity bridged tokens" \
        --cosmos-rpc="http://localhost:1317"  \
        --fees="must be the erc20 your sending back in the form peggy0xXXXXX" \
        --erc20-address="0xXXXXXXX" \
        --amount=.5 \
        --eth-destination="any eth address, try your delegate eth address"
```

It will take a moment or two for Etherescan to catch up, but once it has you'll see the new ERC20 token balance reflected at https://goerli.etherscan.io/

### Really testing Gravity

Now that we have the basics out of the way we can get into the fun testing, including hundreds of transactions across the bridge, upgrades, and slashing. Depending on how the average participant is doing we may or may not get to this during our chain start call.

- Send a 100 transaction batch
- Send 100 deposits to the Althea chain from Ethereum
- IBC bridge some tokens to another chain
- Exchange those bridged tokens on the Gravity DEX
- Have a governance vote to reduce the slashing period to 1 hr downtime, then have a volunteer get slashed
- Stretch goal, upgrade the testnet the following week for Gravity V2 features. This may end up not being practical depending on the amount of changes made.
