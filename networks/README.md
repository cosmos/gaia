# Cosmos Hub Mainnet

## Overview

The current Gaia Version of the Cosmos Hub mainnet is [`v6.0.4`](https://github.com/cosmos/gaia/releases/tag/v6.0.4). To bootstrap a mainnet node, it is possible to sync from `v6.0.4` via Quicksync or via [State Sync](https://hub.cosmos.network/main/hub-tutorials/join-mainnet.html#state-sync).

For a full set of instructions on boostrapping a mainnet node, see the Hub's [**Join the Cosmos Hub Mainnet**](https://hub.cosmos.network/main/hub-tutorials/join-mainnet.html) documentation.

However to build a node from scratch a node operator will need to first run [v4.2.6](https://github.com/cosmos/gaia/releases/tag/v4.2.6) until the node panics at block height [6910000](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-delta-upgrade.md#Upgrade-will-take-place-July-12,-2021). The node should stop running after the panic, if it does not stop automatically, wait for 5-10 minutes and then kill it manually. Then install the latest version of gaia ([v5.0.2](https://github.com/cosmos/gaia/releases/tag/v5.0.2)) and then begin running the binary agian with the optional flag `--x-crisis-skip-assert-invariants`. This will begin syncing the node since the last upgrade until it is at the current height.

## Scheduled Upgrade 🗓️ 

The `v7-Theta` upgrade was proposed through an [on-chain software upgrade proposal](https://www.mintscan.io/cosmos/proposals/65) and is open to voting between 2022-03-25 and 2022-04-08. The upgrade is proposed to take place at block height **10,085,397** which should occur approximately at **April 12th, 2022 at 12:00:00 UTC.** The chain id will remain `cosmoshub-4`. 

You can find `v7.0.0` tagged code and binaries in the [gaia repository](https://github.com/cosmos/gaia/releases/tag/v7.0.0). The json for the on-chain proposal is [archived in this repository](upgrades/v7-Theta-proposal.json) for future reference. 

We recommend that you participate in the [Theta testnet](https://github.com/cosmos/testnets/tree/master/v7-theta/public-testnet) to familiarize yourself with the upgrade process. From experience on the Theta testnet, we expect this upgrade to be much faster than the Vega upgrade due to a simpler migration.

Upgrades can be a memory intensive process. Please see the guide for the [current hardware recommendations](https://hub.cosmos.network/main/hub-tutorials/join-mainnet.html#hardware). We recommend that you are prepared to upgrade on a machine with sufficient physical memory. Some node operators also choose add 8-16 GB of swap to avoid out of memory issues.

You can check out the Cosmos Hub documentation for a step-by-step tutorial on how to [upgrade your node](https://hub.cosmos.network/main/hub-tutorials/upgrade-node.html). Version specific upgrade notes using Cosmovisor are provided below.

### Upgrading using Cosmovisor

At the proposed halt height, you should expect to see a message like:

```
ERR UPGRADE "v7-Theta" NEEDED at height: 10085397
```

We recommend that you use [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#auto-download) for managing your gaia upgrades. You can do this in one of two ways:

#### 1. Autodownload

If you're using Cosmovisor's **autodownload** feature, please set the environment variable `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`

**IMPORTANT:** In case you're using auto-download, on Gaia v6.0.0 or v6.0.3 Cosmosvisor won't auto-download the binary unfortunately. v6.0.4 will work fine. Please refer to [this issue](https://github.com/cosmos/gaia/issues/1342) for details.

#### 2. Manual

If you're **manually preparing your binary**, please download v7.0.0 and move the binary to the v7-Theta upgrade directory in your cosmovisor directory

```
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad
└── upgrades
    └── v7-Theta
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

## Quickstart

**Preresquisites**
- `make` & `gcc`
- `Go 1.16+`

> **Note**: Make sure to have all prerequisites installed. See the [installation docs](https://hub.cosmos.network/main/getting-started/installation.html) for clarification and a detailed set of instructions.

**Quicksync**

Quicksync.io offers several daily snapshots of the Cosmos Hub with varying levels of pruning (archive 1.4TB, default 540GB, and pruned 265GB). For downloads and installation instructions, visit the [Cosmos Quicksync guide](https://quicksync.io/networks/cosmos.html).

**State Sync**

To enable state sync, visit an [explorer](https://www.mintscan.io/cosmos/blocks) to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is 1000 blocks, it is advised to choose something close to current height - 1000. Set these parameters in the code snippet below `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`

For reference, the list of `rpc_servers` and `persistent` peers can be found in the [cosmos hub chain-registry repo](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json).

```bash
# Build gaiad binary and initialize chain
cd $HOME
git clone -b v6.0.4 https://github.com/cosmos/gaia
cd gaiad
make install
gaiad init <custom moniker>

# Prepare genesis file for cosmoshub-4
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json $HOME/.gaia/config/genesis.json

#Set minimum gas price & peers
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.001uatom"/' app.toml
sed -i 's/persistent_peers = ""/persistent_peers = "6e08b23315a9f0e1b23c7ed847934f7d6f848c8b@165.232.156.86:26656,ee27245d88c632a556cf72cc7f3587380c09b469@45.79.249.253:26656,538ebe0086f0f5e9ca922dae0462cc87e22f0a50@34.122.34.67:26656,d3209b9f88eec64f10555a11ecbf797bb0fa29f4@34.125.169.233:26656,bdc2c3d410ca7731411b7e46a252012323fbbf37@34.83.209.166:26656,585794737e6b318957088e645e17c0669f3b11fc@54.160.123.34:26656,5b4ed476e01c49b23851258d867cc0cfc0c10e58@206.189.4.227:26656"/' config.toml

# Configure State sync
cd $HOME/.gaia/config
sed -i 's/enable = false/enable = true/' config.toml
sed -i 's/trust_height = 0/trust_height = <BLOCK_HEIGHT>/' config.toml
sed -i 's/trust_hash = ""/trust_hash = "<BLOCK_HASH>"/' config.toml
sed -i 's/rpc_servers = ""/rpc_servers = "https:\/\/rpc.cosmos.network:443,https:\/\/rpc.cosmos.network:443"/' config.toml

#Start Gaia
gaiad start --x-crisis-skip-assert-invariants
```

**Sync from Scratch**

```bash
# Build gaiad binary and initialize chain
git clone -b v4.2.1 https://github.com/cosmos/gaia
cd gaia
make install
gaiad init <custom moniker>

# Prepare genesis file for cosmoshub-4
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json


#Set minimum gas price & peers
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.001uatom"/' app.toml
sed -i 's/persistent_peers = ""/persistent_peers = "6e08b23315a9f0e1b23c7ed847934f7d6f848c8b@165.232.156.86:26656,ee27245d88c632a556cf72cc7f3587380c09b469@45.79.249.253:26656,538ebe0086f0f5e9ca922dae0462cc87e22f0a50@34.122.34.67:26656,d3209b9f88eec64f10555a11ecbf797bb0fa29f4@34.125.169.233:26656,bdc2c3d410ca7731411b7e46a252012323fbbf37@34.83.209.166:26656,585794737e6b318957088e645e17c0669f3b11fc@54.160.123.34:26656,5b4ed476e01c49b23851258d867cc0cfc0c10e58@206.189.4.227:26656"/' config.toml

gaiad start --x-crisis-skip-assert-invariants
```
Now wait until the chain reaches block height 6910000. It will panic and log the following:
```
ERR UPGRADE "Gravity-DEX" NEEDED at height: 6910000: v5.0.0-4760cf1f1266accec7a107f440d46d9724c6fd08

panic: UPGRADE "Gravity-DEX" NEEDED at height: 6910000: v5.0.0-4760cf1f1266accec7a107f440d46d9724c6fd08
```

It's now time to perform the manual Delta upgrade:
```bash
git checkout -b v5.0.2
make install
gaiad start --x-crisis-skip-assert-invariants
```

Once `V5` reaches the upgrade block height, the chain will halt and display the following message:
```
ERR UPGRADE "Vega" NEEDED at height: 8695000

```

This will indicate it is time to perform the Vega upgrade. Similar with the previous upgrade, checkout `V6`, compile the new binary and restart `gaiad`

```bash
git checkout -b v6.0.4
make install
gaiad start --x-crisis-skip-assert-invariants
```

> _NOTE_:  If the node is unable to connect to any of the seeds listed here, find additional seeds and peers in [this document](https://hackmd.io/@KFEZk8oMTz6vBlwADz0M4A/BkKEUOsZu#) maintained by community members, and at [Atlas](https://atlas.cosmos.network/nodes), which is automatically generated by crawling the network. Additionally, node operators can just copy [Quicksync's addressbook](https://quicksync.io/addrbook.cosmos.json) and move it to `$HOME/.gaia/config/addrbook.json`


## Upgrade to Validator Node

You now have an active full node. What's the next step? You can upgrade your full node to become a Cosmos Validator. The top 125 validators have the ability to propose new blocks to the Cosmos Hub. Continue onto [the Validator Setup](../validators/validator-setup.md).
