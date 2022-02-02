<!--
order: 3
-->

# Quick Start

**Bootstrap a  `cosmoshub-4` mainnet node**

### Prerequisites
> **Note**: Make sure the [Gaia CLI is installed](./installation.md).

### Sync Options
To quickly get started, node operators can choose to sync via State Sync or by downloading a snapshot from Quicksync. State Sync works by replaying larger chunks of application state directly rather than replaying individual blocks or consensus rounds. Quicksync is a service provided courtesy of ChainLayer, and offers historical state of the chain available for download every 24 hours. For more advanced information on setting up a node, see the Sync Options section of the full [Joining Mainnet Tutorial](../hub-tutorials/joining-mainnet.md)

<!-- #sync options -->
::::::: tabs :options="{ useUrlFragment: false }"

:::::: tab "State Sync"

To enable state sync, visit an [explorer](https://www.mintscan.io/cosmos/blocks) to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is 1000 blocks, it is advised to choose something close to current height - 1000. Set these parameters in the code snippet below `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`

For reference, the list of `rpc_servers` and `persistent` peers can be found in the [cosmos hub chain-registry repo](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json).

```bash
# Build gaiad binary and initialize chain
cd $HOME
git clone -b v6.0.0 https://github.com/cosmos/gaia
cd gaiad
make install
gaiad init <custom moniker>

# Prepare genesis file for cosmoshub-4
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json $HOME/.gaia/config/genesis.json

#Set minimum gas price & peers
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025uatom"/' app.toml
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
::::::

:::::: tab Quicksync

> **Note**: Make sure to set the `--home` flag when initializing and starting `gaiad` if mounting quicksync data externally.

#### Create Gaia Home & Config
```bash
mkdir $HOME/.gaia/config -p
```

#### Start Quicksync Download
<!-- #quicksync options -->
Node Operators can decide how much of historical state they want to preserve by choosing between `Pruned`, `Default`, and `Archive`. See the [Quicksync.io downloads](https://quicksync.io/networks/cosmos.html) for up to date snapshot sizes.

:::: tabs :options="{ useUrlFragment: false }"

::: tab Default
```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-default")|.url'`

aria2c -x5 $URL
```
:::

::: tab Pruned
```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-pruned")|.url'`

aria2c -x5 $URL
```
:::

::: tab Archive
```bash=
sudo apt-get install wget liblz4-tool aria2 jq -y

export URL=`curl https://quicksync.io/cosmos.json|jq -r '.[] |select(.file=="cosmoshub-4-archive")|.url'`

aria2c -x5 $URL
```
:::

::::

<!-- #end -->

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

#### Unzip
```bash
lz4 -c -d `basename $URL` | tar xf -
```


#### Copy Address Book Quicksync
```bash
curl https://quicksync.io/addrbook.cosmos.json > $HOME/.gaia/config/addrbook.json
```


#### Start Gaia
```bash
gaiad start --x-crisis-skip-assert-invariants

```
::::::

:::::::

<!-- #end -->
