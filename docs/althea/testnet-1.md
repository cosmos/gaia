# Althea Testnet 1

Althea Testnet #1 mainly focuses around the Gravity bridge integration. Our goal is to run this testnet right up until the launch of the Althea chain

- Althea chain parameter selection
- Gravity bridge slashing
- Gravity bridge Orchestrator stability
- IBC testing with B-Harvest and Agoric

If you would like to join this testnet connect with us on the [Althea Discord](https://discordapp.com/invite/vw8twzR) where you can request testnet tokens to stake.

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

I also suggest an open notepad or other document to keep track of the keys you will be generating.

## Bootstrapping steps and commands

We’re going to have a centralized start testnet. Where Althea will launch a chain, send everyone else tokens, and then each participant will come in and ualtg to become a validator.
In order to further simplify bootstrapping for this testnet we will be using pre-built binaries I am placing into a github release. These include ARM binaries for those of you on ARM platforms. Note that you will need to be running a 64bit ARM machine with a 64 bit operating system to use these binaries. In order to download ARM binaries change the names in the wget links from ‘client’ to ‘client-arm’. Repeat for all binaries

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
althea init mymoniker --chain-id althea-testnet1v2
althea keys add myvalidatorkeyname
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea-testnet1-v2-genesis.json
cp althea-testnet1-v2-genesis.json $HOME/.althea/config/genesis.json
```

### Add persistent peers

Change the p2p.persistent_peers field in ~/.althea/config/config.toml to contain the following:

```
persistent_peers = "05ded2f258ab158c5526eb53aa14d122367115a7@testnet1.althea.net:26656"
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
  --chain-id=althea-testnet1v2 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --gas="auto" \
 --gas-adjustment=1.5 \
  --gas-prices="0.025ualtg" \
  --from=myvalidatorkeyname

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

After you do this you need to restart your validator hit ctrl-c and then run 'althea start' again

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
althea tx bank send myvalidatorkeyname <your delegate cosmos address> 50000000footoken --chain-id=althea-testnet1v2
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
./geth --syncmode "light" --goerli --http --cache 16
```

### Deployment of the Gravity contract

Once 66% of the validator set has registered their delegate Ethereum key it is possible to deploy the Gravity Ethereum contract. Once deployed the Gravity contract address on Görli will be posted here

Here is the contract address! Move forward!

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

### Appendix

#### Increase your stake

To increase your ualtg stake, if you have extra tokens lying around. The first command will show an output like this, you want to take the key starting with cosmosvaloper1 in the 'address' field.

```
- name: jkilpatr
  type: local
  address: cosmosvaloper1jpz0ahls2chajf78nkqczdwwuqcu97w6z3plt4
  pubkey: cosmosvaloperpub1addwnpepqvl0qgfqewmuqvyaskmr4pwkr5fwzuk8286umwrfnxqkgqceg6ksu359m5q
  mnemonic: ""
  threshold: 0
  pubkeys: []

```

```
althea keys show myvalidatorkeyname --bech val
althea tx staking delegate <the address from the above command> 99000000ualtg --from myvalidatorkeyname --chain-id althea-testnet1v2 --fees 50ualtg --broadcast-mode block
```

#### Unjail your validator

You can be jailed for several different reasons. As part of the Althea testnet we are testing slashing conditions for the Gravity bridge, so you will be slashed if the Orchestrator is not running properly, in addition to the usual Cosmos double sign and downtime slashing parameters. To unjail your validator run

```
althea tx slashing unjail --from myvalidatorkeyname --chain-id=althea-testnet1v2
```

### Upgrading from althea-testnet1 to altheatestnet1v2

Thank you very much for your patience and participation! During the initial launch of testnet1 we encountered several bugs that have now been patched. See our [blog post](https://blog.althea.net/althea-testnet-1-launched/) on the bugs you helped fix!

For now though lets talk about getting everyone back up and running

_You will keep your tokens and your validator in v2, as part of this guide you will unjail yourself and return to validating_

#### Update your binaries

We have a new version of everything as the fixes are quite expansive

```
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

#### Update your genesis file

This is the exported genesis file of the chain history we started on the 13th, we'll import it into our new updated chain keeping all balances and state

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.1/althea-testnet1-v2-genesis.json
cp althea-testnet1-v2-genesis.json $HOME/.althea/config/genesis.json
```

#### Start the chain

Unsafe reset all will reset the entire blockchain state in .althea allowing you to start althea-testnet1v2 using only the state from the genesis file

```
althea unsafe-reset-all
althea start
```

#### Restart your Orchestrator

No argument changes are required, just ctrl-c and start it again.

You may also want to check the status of your Geth node, no changes are required there.

#### Unjail yourself

This command will unjail you, completing the process of getting the chain back online!

_replace 'myvalidatorkeyname' with your validator keys name, if you don't remember run `althea keys list`_

```
althea tx slashing unjail --from myvalidatorkeyname --chain-id=althea-testnet1v2
```

#### Notes

The updated orchestrator addresses a lot of the community concerns and error messages, you should only see one error message, talking about potential bridge highjacking. It will go away in a few hours and is related to the fact that the chain stopped shortly after the time that I used for the genesis file. This is just a warning working as intended and can be safely ignored in this case.

### Stress test the Gravity bridge!

As part of our testnet we're going to try and put the maximum possible load on the bridge, Sunday February 21st at around 1pm west coast time. This isn't a hard time event, you can simply setup your client to send thousands of transactions and leave it alone for the rest of the day. We just want to get as many people as possible doing so.

This guide does not require that you be a validator on althea-testnet1v2, it is designed to run on any machine.

#### Download Althea chain and the Gravity tools

The client binary has been updated for this guide.

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

#### Generate your Althea testnet Cosmos address

If you are a validator you can skip this step as you already have an address ready. Or you can generate a new address there's not much real difference. Just remember to keep track of your seed phrases.

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1v2
althea keys add mytestingname
```

#### Request some ERC20 tokens on Cosmos

Paste your Cosmos key in the chat and request some ERC20 to test with (you may already have some if you've been testing as a validator)

#### Build a big batch

The goal of our test is to build 3 large transaction batches (over 100 transactions) and then have them
flow through the Gravity bridge all at once. When you request tokens I'll send you 100 of each type of ERC20 which you'll then use these commands to send back. They will take quite some time to run, that's intentional we want to have the largest number of different people sending transactions at once, not the most spam a single person can send at once

```
RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc:1317"  \
        --erc20-address="0xD7600ae27C99988A6CD360234062b540F88ECA43" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"

RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc:1317"  \
        --erc20-address="0x7580bFE88Dd3d07947908FAE12d95872a260F2D8" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"

RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc:1317"  \
        --erc20-address="0xD50c0953a99325d01cca655E57070F1be4983b6b" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"
```

#### Wait for it

the above commands have 'no-batch' set, which means the funds won't show up on Ethereum instead they are waiting for a relayer to deem the batch profitable enough to relay and request it. This is where our scheduled time comes in. I'll request batches for all three tokens at around 1pm West coast time.
