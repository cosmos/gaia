# How to run a Althea testnet full node

A Althea chain full node is just like any other Cosmos chain full node and unlike the validator flow requires no external software

## What do I need?

A Linux server with any modern Linux distribution, 2gb of ram and at least 20gb storage. Requirements are very minimal.

### Download Althea chain software

```
# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/althea-0.0.4-16-g6812f87-linux-amd64
mv althea-0.0.4-16-g6812f87-linux-amd64 althea

chmod +x althea
sudo mv althea /usr/bin/
```

### Init the config files

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1v5
```

### Copy the genesis file

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/althea-testnet1-v5-genesis.json
cp althea-testnet1-v5-genesis.json $HOME/.althea/config/genesis.json
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
