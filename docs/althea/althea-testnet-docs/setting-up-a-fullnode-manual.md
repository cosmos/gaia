# How to run a Althea testnet full node

A Althea chain full node is just like any other Cosmos chain full node and unlike the validator flow requires no external software

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

### Download Althea chain software

```
# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/althea-0.0.5-26-gcf92625-linux-amd64
mv althea-0.0.5-26-gcf92625-linux-amd64 althea

chmod +x althea
sudo mv althea /usr/bin/
```

### Init the config files

```
cd $HOME
althea init mymoniker --chain-id althea-testnet2v1
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/althea-testnet2-v1-genesis.json
cp althea-testnet2-v1-genesis.json $HOME/.althea/config/genesis.json
```

### Add seed node

Change the seed_nodes field in ~/.althea/config/config.toml to contain the following:

```

seed_nodes = "c6031b3c9f179a3884ef559f@testnet2.althea.net:26656"

```

### Start your full node and wait for it to sync

Ask what the current blockheight is in the chat

```
althea start
```
