<!--
order: 4
-->

# Join the Public Testnet

| Current Upgrade | Chain Id       | Upgrade Block Height | Upgrade Date     |
| --------------- | -------------- | -------------------- | ---------------- |
| Vega            | `vega-testnet` | `7453750`            | November 12 2021 |


## Background
The Cosmos Hub Testnet is currently running after it's most recent [Vega Upgrade](https://interchain-io.medium.com/cosmos-hub-vega-upgrade-testnet-details-e9c5d69a59c). Visit the [testnet explorer](https://vega-explorer.hypha.coop/) to view all on chain activity.

For those who just need instructions on performing the upgrade, see the [Upgrade](#upgrading) section.

## Prerequisites

**Hardware**

It's recommended that public testnet nodes are running on machines with at least `16GB` of RAM.

**Make sure Go & Gaia are [properly installed](../getting-started/installation.md). The most recent Gaia version for the Vega Testnet is [`v6.0.0-rc3`](https://github.com/cosmos/gaia/tree/v6.0.0-rc3)**


This tutorial will provide all necessary instructions for joining the current public testnet. If you're interested in more advanced configuration and synchronization options, see [Join Mainnet](./join-mainnet.md) for a detailed walkthrough.

## Sync Options
There are two ways to sync a testnet node, Blocksync and State Sync. [Blocksync](https://docs.tendermint.com/v0.35/tendermint-core/block-sync/) syncs the chain from genesis by downloading blocks in paralell and then verifying them. [State Sync](https://docs.tendermint.com/master/tendermint-core/state-sync/#) will look for snapshots from peers at a trusted height and then verifying a minimal set of snapshot chunks against the network.

State Sync is far faster and more efficient than Blocksync, but Blocksync offers higher data integrity and more robust history. For those who are concerned about storage and costs, State Sync can be the better option as it minimizes storage usage when rebuilding initial state.

### Configuration & Setup

To get started, you'll need to install and configure the Gaia binary using the script below. **For Blocksync, it is important to checkout Gaia `release/v5.0.5`. For State Sync checkout `release/v6.0.0-rc3`**

This example is using the Vega testnet genesis. For up to date values like `persistent_peers`, visit the [testnet repository](https://github.com/cosmos/testnets).

> **Note**: Cosmos Hub recommends running `gaiad` or `cosmovisor` with the `--x-crisis-skip-assert-invariants` flag. If checking for invariants, operators are likely to see `rounding error withdrawing rewards from validator`. These are expected. For more information see [Verify Mainnet](./join-mainnet.md#verify-mainnet)

```
# Build gaiad binary and initialize chain
cd $HOME
git clone -b release/<release_version> https://github.com/cosmos/gaia
cd gaia
make install
gaiad init <custom_moniker>

# Prepare genesis file
wget https://github.com/cosmos/vega-test/raw/master/public-testnet/modified_genesis_public_testnet/genesis.json.gz
gzip -d genesis.json.gz
mv genesis.json $HOME/.gaia/config/genesis.json

# Set minimum gas price & peers
cd $HOME/.gaia/config
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.001uatom"/' app.toml
sed -i 's/persistent_peers = ""/persistent_peers = "<persistent_peer_node_id_1@persistent_peer_address_1:p2p_port>,<persistent_peer_node_id_2@persistent_peer_address_2:p2p_port>"/' config.toml
```

### State Sync

::: warning
State Sync requires Gaia version [`v6.0.0-rc3`](https://github.com/cosmos/gaia/tree/release/v6.0.0-rc3).
:::

There will need to be additional configuration to enable State Sync on the testnet. State Sync requires setting an initial list of `persistent_peers` to fetch snapshots from. This will change and eventually move to the p2p layer when the Cosmos Hub upgrades to [Tendermint `v0.35`](https://github.com/tendermint/tendermint/issues/6491). For the sake of simplicity, this step is already done in the [Configuration & Setup](#configuration-amp=-setup) section.

Visit a [testnet explorer](https://vega-explorer.hypha.coop/) to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is 1000 blocks, it is advised to choose something close to current height - 1000. Set these parameters in the code snippet below `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`

For up to date values like `rpc_servers`, visit the current [testnet repository](https://github.com/cosmos/testnets).

```
cd $HOME/.gaia/config
sed -i 's/enable = false/enable = true/' config.toml
sed -i 's/trust_height = 0/trust_height = <BLOCK_HEIGHT>/' config.toml
sed -i 's/trust_hash = ""/trust_hash = "<BLOCK_HASH>"/' config.toml
sed -i 's/rpc_servers = ""/rpc_servers = "<rpc_address_1>:26657,<rpc_address_2>:26657"/' config.toml
```

Now run `gaiad start --x-crisis-skip-assert-invariants` or if using [Cosmovisor](#using-cosmovisor),  `cosmovisor start --x-crisis-skip-assert-invariants`. Once a snapshot is found and verified, the chain will start syncing via regular consensus within minutes.

### Using Cosmovisor

Cosmovisor is a process manager that monitors the governance module for incoming chain upgrade proposals. When a proposal is approved, Cosmovisor can automatically download the new binary, stop the chain when it hits the upgrade height, switch to the new binary, and restart the daemon. This tutorial will provide instructions for the most efficient way to sync via Cosmovisor. For more information on configuration, check out the Cosmos SDK's [Cosmovisor repository documentation](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#auto-download).

Cosmovisor can be used when syncing with Blocksync or State Sync. Make sure to follow the Cosmovisor setup below, and then run `cosmovisor start` in place of `gaiad start`.

Cosmovisor requires the creation the following directory structure:
```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad
└── upgrades
    └── Vega
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

It is possible to enable autodownload for the new binary, but for the purpose of this tutorial, the setup instructions will include how to do this manually. For more information on autodownload with Cosmovisor, see the Vega Testnet respository's [documentation on Cosmosvisor](https://github.com/cosmos/vega-test/blob/master/local-testnet/README.md#Cosmovisor).

The following script installs, configures and starts Cosmovisor:

```
# Install Cosmovisor
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor

# Set environment variables
echo "export DAEMON_NAME=gaiad" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.gaia" >> ~/.profile
source ~/.profile

mkdir -p ~/.gaia/cosmovisor/upgrades
mkdir -p ~/.gaia/cosmovisor/genesis/bin/
cp $(which gaiad) ~/.gaia/cosmovisor/genesis/bin/

# Verify cosmovisor and gaiad versions are the same.
cosmovisor version

# Start Cosmovisor
cosmovisor start --x-crisis-skip-assert-invariants
```

#### Upgrading

Cosmovisor will continually poll the `$DAEMON_HOME/data/upgrade-info.json` for new upgrade instructions. When an upgrade is ready, node operators can download the new binary and place it under `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin` where `<name>` is the URI-encoded name of the upgrade as specified in the upgrade module plan.

When the chain reaches block height `7,453,750`, the chain will halt and you will have to download the new binary and move it to the correct folder. For the `Vega` upgrade, this would look like:
```
# Prepare Vega upgrade directory
mkdir -p ~/.gaia/cosmovisor/upgrades/Vega/bin

# Download and install the new binary version.
cd $HOME/gaia
git pull
git checkout v6.0.0-rc3
make install

# Move the new binary to the Vega upgrade directory
cp $GOPATH/bin/gaiad ~/.gaia/cosmovisor/upgrades/Vega/bin
```

If Cosmovisor is already running, there's nothing left to do, otherwise run `cosmovisor start` to start the daemon.

### Blocksync
Blocksync will require nagivating the Vega upgrade either via [Cosmovisor](#using-cosmovisor) or manually.

Manually updating `gaiad` will require stopping the chain and installing the new binary once it halts at block height `7,453,750`.

Logs will show `ERR UPGRADE "Vega" NEEDED at height: 7453750`. Stop `gaiad` and run the following:

```
cd $HOME/gaia
git checkout release/v6.0.0-rc3
make install

# Verify the correct installation
gaiad -version
```

Once the new binary is installed, restart the Gaia daemon. Logs will show `INF applying upgrade "Vega" at height: 7453750`. After a few minutes, the node will start syncing blocks.
