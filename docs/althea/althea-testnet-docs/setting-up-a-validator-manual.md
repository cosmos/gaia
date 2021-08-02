# How to become a validator on the Althea testnet!

## What do I need?

A Linux server with any modern Linux distribution, 4cores, 8gb of ram and at least 20gb of SSD storage.

Althea chain can be run on Windows and Mac. Binaries are provided on the releases page. But validator instructions are not provided.

I also suggest an open notepad or other document to keep track of the keys you will be generating.

## Bootstrapping steps and commands

### Download Althea chain and the Gravity tools

```

mkdir althea-bin
cd althea-bin

# the althea chain binary itself

wget https://github.com/althea-net/althea-chain/releases/download/v0.2.3/althea-0.2.2-18-g73447b6-linux-amd64
mv althea-0.2.2-18-g73447b6-linux-amd64 althea

# Tools for the gravity bridge from the gravity repo

wget https://github.com/althea-net/althea-chain/releases/download/v0.2.3/gbt
chmod +x *
sudo mv * /usr/bin/

```

At specific points during the testnet you may be told to 'update your orchestrator' or 'update your althea binary'. In order to do that you can simply repeat the above instructions and then restart the affected software.

to check what version of the tools you have run `gbt --version` the current latest version is `gbt 0.5.6`

### Generate your key

Be sure to back up the phrase you get! You’ll need it in a bit. If you don't back up the phrase here just follow the steps again to generate a new key.

Note 'myvalidatorkeyname' is just the name of your key here, you can pick anything you like, just remember it later.

You'll be prompted to create a password, I suggest you pick something short since you'll be typing it a lot

```

cd $HOME
althea init mymoniker --chain-id althea-testnet2v3
althea keys add myvalidatorkeyname

```

### Copy the genesis file

```

wget https://github.com/althea-net/althea-chain/releases/download/v0.2.3/althea-testnet2v3-genesis.json
cp althea-testnet2v3-genesis.json $HOME/.althea/config/genesis.json

```

### Add seed node

Change the seed field in ~/.althea/config/config.toml to contain the following:

```

seeds = "6a9cd8d87ab9e49d7af91e09026cb3f40dec2f85@testnet2.althea.net:26656,3b8af242bf3bb82c203d1d8ef0949f8ca48767c8@althea-sentry-01.mahdisworld.net:26656,02c2b59771c3626f6744ab1fb1048ba967cb82cd@althea-sentry-02.mahdisworld.net:26656"

```

### Increasing the default open files limit

If we don't raise this value nodes will crash once the network grows large enough

```

sudo su -c "echo 'fs.file-max = 65536' >> /etc/sysctl.conf"
sysctl -p

sudo su -c "echo '* hard nofile 94000' >> /etc/security/limits.conf"
sudo su -c "echo '* soft nofile 94000' >> /etc/security/limits.conf"

sudo su -c "echo 'session required pam_limits.so' >> /etc/pam.d/common-session"
```

For this to take effect you'll need to (A) reboot (B) close and re-open all ssh sessions

To check if this has worked run

```
ulimit -n
```

If you see `1024` _then you need to reboot_

### Start your full node and wait for it to sync

Ask what the current blockheight is in the chat

```

althea start

```

### Request some funds be sent to your address

First find your address

```

althea keys list

```

Copy your address from the 'address' field and paste it into the command below remember to remove the `<>`

```

curl -vv -XPOST http://testnet2.althea.net/get_altg/<your address here without the brackets>

```

This will provide you 10 ALTG from the faucet storage.

### Send your validator setup transaction

```

althea tx staking create-validator \
 --amount=50000000000ualtg \
 --pubkey=$(althea tendermint show-validator) \
 --moniker="put your validator name here" \
 --chain-id=althea-testnet2v3 \
 --commission-rate="0.10" \
 --commission-max-rate="0.20" \
 --commission-max-change-rate="0.01" \
 --min-self-delegation="1" \
 --gas="auto" \
 --gas-adjustment=1.5 \
 --gas-prices="1ualtg" \
 --from=myvalidatorkeyname

```

### Confirm that you are validating

If you see one line in the response you are validating. If you don't see any output from this command you are not validating. Check that the last command ran successfully.

Be sure to replace 'my validator key name' with your actual key name. If you want to double check you can see all your keys with 'althea keys list'

```

althea query staking validator $(althea keys show myvalidatorkeyname --bech val --address)

```

### Setup Gravity bridge

You are now validating on the Althea blockchain. But as a validator you also need to run the Gravity bridge components or you will be slashed and removed from the validator set after about 16 hours.

### Register your delegate keys

Delegate keys allow the for the validator private keys to be kept in secure storage while the Orchestrator can use it's own delegated keys for Gravity functions. The delegate keys registration tool will generate Ethereum and Cosmos keys for you if you don't provide any. These will be saved in your local config for later use.

\*\*If you have set a minimum fee value in your `~/.althea/config/app.toml` modify the `--fees` parameter to match that value!

```

gbt init

gbt -a althea keys register-orchestrator-address --validator-phrase "the phrase you saved earlier" --fees=125000ualtg

```

#### Registering your delegate keys using a Ledger

**If you ran the above command skip this step**

This is an **optional** step for those who are using hardware security for their validator key. In order to register your keys you will use the `althea` cli instead of `gbt`. You will need to generate one ethereum key and one cosmos key yourself.

```

althea tx gravity set-orchestrator-address [validator key name] [orchestrator key name] [ethereum-address]

```

### Fund your delegate keys

Both your Ethereum delegate key and your Cosmos delegate key will need some tokens to pay gas. On the Althea chain side you where sent some 'footoken' along with your ALTG. We're essentially using footoken as a gas token for this testnet.

In a production network only relayers would need Ethereum to fund relaying, but for this testnet all validators run relayers by default, allowing us to more easily simulate a lively economy of many relayers.

You should have received fifty thousand Althea Governance Token in ALTG and the same amount of footoken. We're going to send half of the footoken to the delegate address

To get the address for your validator key you can run the below, where 'myvalidatorkeyname' is whatever you named your key in the 'generate your key' step.

```

althea keys show myvalidatorkeyname

```

```

althea tx bank send myvalidatorkeyname <your delegate cosmos address> 25000000000ufootoken --chain-id=althea-testnet2v3

```

With the Althea side funded, now we need some Goerli Eth

```

curl -vv -XPOST http://testnet2.althea.net/get_eth/<your address here without the brackets>

```

### Download and setup Geth on the Goerli testnet

We will be using Geth Ethereum light clients for this task. For production Gravity we suggest that you point your Orchestrator at a Geth light client and then configure your light client to peer with full nodes that you control. This provides higher reliability as light clients are very quick to start/stop and resync. Allowing you to for example rebuild an Ethereum full node without having days of Orchestrator downtime.

Geth full nodes do not serve light clients by default, light clients do not trust full nodes, but if there are no full nodes to request proofs from they can not operate. Therefore we are collecting the largest possible
list of Geth full nodes from our community that will serve light clients.

If you have more than 40gb of free storage, an SSD and extra memory/CPU power, please run a full node and share the node url. If you do not, please use the light client instructions

_Please only run one or the other of the below instructions, both will not work_

#### Light client instructions

```

wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.10.6-576681f2.tar.gz
tar -xvf geth-linux-amd64-1.10.6-576681f2.tar.gz
cd geth-linux-amd64-1.10.6-576681f2
wget https://github.com/althea-net/althea-chain/raw/main/docs/althea/configs/geth-light-config.toml
./geth --syncmode "light" --goerli --http --config geth-light-config.toml

```

#### Fullnode instructions

```

wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.10.6-576681f2.tar.gz
tar -xvf geth-linux-amd64-1.10.6-576681f2.tar.gz
cd geth-linux-amd64-1.10.6-576681f2
wget https://github.com/althea-net/althea-chain/raw/main/docs/althea/configs/geth-full-config.toml
./geth --goerli --http --config geth-full-config.toml

```

You'll see this url, please note your ip and share both this node url and your ip in chat to add to the light client nodes list

```
INFO [06-10|14:11:03.104] Started P2P networking self=enode://71b8bb569dad23b16822a249582501aef5ed51adf384f424a060aec4151b7b5c4d8a1503c7f3113ef69e24e1944640fc2b422764cf25dbf9db91f34e94bf4571@127.0.0.1:30303
```

Finally you'll need to wait for several hours until your node is synced, you can not continue with the instructions until your node is synced.

### Deployment of the Gravity contract

Once 66% of the validator set has registered their delegate Ethereum key it is possible to deploy the Gravity Ethereum contract. Once deployed the Gravity contract address on Görli will be posted here

Here is the contract address! Move forward!

```

0xFA2f45c5C8AcddFfbA0E5228bDf7E8B8f4fD2E84

```

### Start your Orchestrator

Now that the setup is complete you can start your Orchestrator. Use the Cosmos mnemonic generated in the 'register delegate keys' step and the Ethereum private key also generated in that step. You should setup your Orchestrator in systemd or elsewhere to keep it running and restart it when it crashes.

If your Orchestrator goes down for more than 16 hours during the testnet you will be slashed and booted from the active validator set.

Since you'll be running this a lot I suggest putting the command into a script, like so. The next version of the orchestrator will use a config file for these values and have encrypted key storage.

\*\*If you have set a minimum fee value in your `~/.althea/config/app.toml` modify the `--fees` parameter to match that value!

```

nano start-orchestrator.sh

```

```

#!/bin/bash
gbt -a althea orchestrator \
 --fees 125000ufootoken \
 --gravity-contract-address "0xFA2f45c5C8AcddFfbA0E5228bDf7E8B8f4fD2E84"

```

```

bash start-orchestrator.sh

```

```

```
