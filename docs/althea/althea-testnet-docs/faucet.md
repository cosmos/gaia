# Althea testnet faucet

In order to use the faucet and receive testnet tokens you'll first need to setup a wallet

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

### Download Althea chain software

```
# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.3/althea-0.2.2-18-g73447b6-linux-amd64
mv althea-0.2.2-18-g73447b6-linux-amd64 althea

chmod +x althea
sudo mv althea /usr/bin/
```

### Init the config files

```
cd $HOME
althea init mymoniker --chain-id althea-testnet2v3
```

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

seeds = "6a9cd8d87ab9e49d7af91e09026cb3f40dec2f85@testnet2.althea.net:26656"

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

Copy your address from the 'address' field and paste it into the command below remember to remove the `<>`

```
curl -vv -XPOST http://testnet2.althea.net/get_altg/<your address here without the brackets>
```

Once you execute this command you should see 10 testnet ALTG in your balance within a few blocks

This faucet also provides Gorli ETH like so

```
curl -vv -XPOST http://testnet2.althea.net/get_eth/<your eth address here without the brackets>
```
