---
title: Quick Start - Join Mainnet 
sidebar_position: 3
---
**Bootstrap a  `cosmoshub-4` mainnet node**

### Prerequisites

> **Note**: Make sure the [Gaia CLI is installed](./installation.md).

### Sync Options

To quickly get started, node operators can choose to sync via State Sync or by downloading a snapshot from Quicksync. State Sync works by replaying larger chunks of application state directly rather than replaying individual blocks or consensus rounds. Quicksync is a service provided courtesy of ChainLayer, and offers historical state of the chain available for download every 24 hours. For more advanced information on setting up a node, see the Sync Options section of the full [Joining Mainnet Tutorial](../hub-tutorials/join-mainnet.md)

#### State Sync 
To enable state sync, visit an [explorer](https://www.mintscan.io/cosmos/blocks) to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is 1000 blocks, it is advised to choose something close to current height - 1000. Set these parameters in the code snippet below `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`

For reference, the list of `rpc_servers` and `persistent` peers can be found in the [cosmos hub chain-registry repo](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json).

```bash
# Build gaiad binary and initialize chain
cd $HOME
git clone -b v18.0.0 https://github.com/cosmos/gaia --depth=1
cd gaiad
make install
gaiad init CUSTOM_MONIKER --chain-id cosmoshub-4

#Set minimum gas price & peers
sed -i'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025uatom"/' $HOME/.gaia/config/app.toml
sed -i'' 's/persistent_peers = ""/persistent_peers = '"\"$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/cosmoshub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')\""'/' $HOME/.gaia/config/config.toml

# Configure State sync
sed -i'' 's/enable = false/enable = true/' $HOME/.gaia/config/config.toml
sed -i'' 's/trust_height = 0/trust_height = <BLOCK_HEIGHT>/' $HOME/.gaia/config/config.toml
sed -i'' 's/trust_hash = ""/trust_hash = "<BLOCK_HASH>"/' $HOME/.gaia/config/config.toml
sed -i'' 's/rpc_servers = ""/rpc_servers = "https:\/\/cosmos-rpc.polkachu.com:443,https:\/\/rpc-cosmoshub-ia.cosmosia.notional.ventures:443,https:\/\/rpc.cosmos.network:443"/' $HOME/.gaia/config/config.toml

#Start Gaia
gaiad start --x-crisis-skip-assert-invariants
```

#### Quick Sync 

**Note**: Make sure to set the `--home` flag when initializing and starting `gaiad` if mounting quicksync data externally.

##### Create Gaia Home & Config

```bash
mkdir $HOME/.gaia/config -p
```

##### Start Quicksync Download

Node Operators can decide how much of historical state they want to preserve by choosing between `Pruned`, `Default`, and `Archive`. See the [Quicksync.io downloads](https://quicksync.io/networks/cosmos.html) for up-to-date snapshot sizes.

###### Default

```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl -L https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-default")|.url'`

echo $URL

cd $HOME/.gaia

aria2c -x5 $URL
```

###### Pruned

```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl -L https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-pruned")|.url'`

echo $URL

cd $HOME/.gaia

aria2c -x5 $URL
```

###### Archive

```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl -L https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-archive")|.url'`

echo $URL

cd $HOME/.gaia

aria2c -x5 $URL
```

**The download logs should look like the following**

```
01/11 07:48:17 [NOTICE] Downloading 1 item(s)
[#7cca5a 484MiB/271GiB(0%) CN:5 DL:108MiB ETA:42m41s]
```

**Completed Download Process:**

```
[#7cca5a 271GiB/271GiB(99%) CN:1 DL:77MiB]
01/11 08:32:19 [NOTICE] Download complete: /mnt/quicksync_01/cosmoshub-4-pruned.20220111.0310.tar.lz4

Download Results:
gid   |stat|avg speed  |path/URI
======+====+===========+=======================================================
7cca5a|OK  |   105MiB/s|/mnt/quicksync_01/cosmoshub-4-pruned.20220111.0310.tar.lz4

Status Legend:
(OK):download completed.
```

##### Unzip

```bash
lz4 -c -d `basename $URL` | tar xf -
```

##### Copy Address Book Quicksync

```bash
curl https://quicksync.io/addrbook.cosmos.json > $HOME/.gaia/config/addrbook.json
```

##### Start Gaia

```bash
gaiad start --x-crisis-skip-assert-invariants

```
