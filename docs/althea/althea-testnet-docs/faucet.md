# Althea testnet faucet

In order to use the faucet and receive testnet tokens you'll first need to setup a wallet

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

### Download Althea chain software

```
# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/althea-0.0.3-4-g30eddc7-linux-amd64
mv althea-0.0.3-4-g30eddc7-linux-amd64 althea

chmod +x althea
sudo mv althea /usr/bin/
```

### Init the config files

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1v4
```

### Generate your key

Be sure to back up the phrase you get! You’ll need it in a bit. If you don't back up the phrase here just follow the steps again to generate a new key.

Note 'myvalidatorkeyname' is just the name of your key here, you can pick anything you like, just remember it later.

You'll be prompted to create a password, I suggest you pick something short since you'll be typing it a lot

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1v4
althea keys add myvalidatorkeyname
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/althea-testnet1-v4-genesis.json
cp althea-testnet1-v4-genesis.json $HOME/.althea/config/genesis.json
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

### Request tokens from the faucet

First list all of your keys 

```
althea keys list
```

You'll see an output like this

```
- name: jkilpatr
  type: local
  address: cosmos1youraddresswillgohere
  pubkey: cosmospub1yourpublickleywillgohere
  mnemonic: ""
  threshold: 0
  pubkeys: []

```

Copy your address from the 'address' field and paste it into the command below


```
curl -X POST -d '{"address":"<address here>"}' https://faucet.althea.hub.hackatom.org
```

Once you execute this command you should see 10 testnet ALTG in your balance within a few blocks
