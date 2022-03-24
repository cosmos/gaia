<!--
order: 4
-->

# Join the Public Testnet

| Current Upgrade | Chain Id       | Upgrade Block Height | Upgrade Date     |
| --------------- | -------------- | -------------------- | ---------------- |
| Theta           | `theta-testnet-001` | TBD   | March 17 2021 |


## Background
The current Cosmos Hub Testnet is running to prepare for the [Theta Upgrade](https://interchain-io.medium.com/preparing-for-the-cosmos-hub-v7-theta-upgrade-2fc41ce34787). Visit the [testnet explorer](https://explorer.theta-testnet.polypore.xyz/) to view all on chain activity.

For those who just need instructions on performing the upgrade, see the [Upgrade](#upgrading) section.

## Releases
If syncing before the Theta update, checkout [`v6.0.0`](https://github.com/cosmos/gaia/tree/v6.0.0). Until a release is cut for the upgrade, feel free to track the [`theta-prepare` branch](https://github.com/cosmos/gaia/tree/theta-prepare).

## Prerequisites

**Hardware**

It's recommended that public testnet nodes are running on machines with at least `16GB` of RAM.

**Make sure Go & Gaia are [properly installed](../getting-started/installation.md). The most recent Gaia version for the Theta Testnet is [`v6.0.0`](https://github.com/cosmos/gaia/tree/v6.0.0).**


This tutorial will provide all necessary instructions for joining the current public testnet. If you're interested in more advanced configuration and synchronization options, see [Join Mainnet](./join-mainnet.md) for a detailed walkthrough.

## Sync Options
There are two ways to sync a testnet node, Blocksync and State Sync. [Blocksync](https://docs.tendermint.com/v0.35/tendermint-core/block-sync/) syncs the chain from genesis by downloading blocks in paralell and then verifying them. [State Sync](https://docs.tendermint.com/master/tendermint-core/state-sync/#) will look for snapshots from peers at a trusted height and then verifying a minimal set of snapshot chunks against the network.

State Sync is far faster and more efficient than Blocksync, but Blocksync offers higher data integrity and more robust history. For those who are concerned about storage and costs, State Sync can be the better option as it minimizes storage usage when rebuilding initial state.

### Configuration & Setup

To get started, you'll need to install and configure the Gaia binary using the script below. **For Blocksync, it is important to checkout Gaia `release/v6.0.0`. For State Sync checkout the most recent [testnet release](https://github.com/cosmos/gaia/tree/v6.0.0) until the upgrade is performed**

This example is using the Theta testnet genesis. For up to date values like `seeds`, visit the [testnet repository](https://github.com/cosmos/testnets).

> **Note**: Cosmos Hub recommends running `gaiad` or `cosmovisor` with the `--x-crisis-skip-assert-invariants` flag. If checking for invariants, operators are likely to see `rounding error withdrawing rewards from validator`. These are expected. For more information see [Verify Mainnet](./join-mainnet.md#verify-mainnet)

```
# Build gaiad binary and initialize chain
cd $HOME
git clone -b release/<release_version> https://github.com/cosmos/gaia
cd gaia
make install
gaiad init <custom_moniker>

# Prepare genesis file
wget https://github.com/hyphacoop/testnets/raw/add-theta-testnet/v7-theta/public-testnet/genesis.json.gz
gzip -d genesis.json.gz
mv genesis.json $HOME/.gaia/config/genesis.json

# Set minimum gas price & peers
cd $HOME/.gaia/config
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.001uatom"/' app.toml
sed -i 's/persistent_peers = ""/persistent_peers = "<persistent_peer_node_id_1@persistent_peer_address_1:p2p_port>,<persistent_peer_node_id_2@persistent_peer_address_2:p2p_port>"/' config.toml
```

### State Sync

::: warning
State Sync requires Gaia version [`v6.0.0`](https://github.com/cosmos/gaia/tree/v6.0.0) until the upgrade is performed.
:::

**Check out the [quickstart script](https://github.com/cosmos/testnets/tree/master/v7-theta/public-testnet#quickstart-on-a-fresh-machine-eg-on-digital-ocean-droplet) to bootstrap a Theta testnet node and configure as needed**

There will need to be additional configuration to enable State Sync on the testnet. State Sync requires setting an initial list of `persistent_peers` to fetch snapshots from. This will change and eventually move to the p2p layer when the Cosmos Hub upgrades to [Tendermint `v0.35`](https://github.com/tendermint/tendermint/issues/6491). For the sake of simplicity, this step is already done in the [Configuration & Setup](#configuration-amp=-setup) section.

Visit a [testnet explorer](https://explorer.theta-testnet.polypore.xyz/) to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is 1000 blocks, it is advised to choose something close to current height - 1000. Set these parameters in the code snippet below `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`

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
    └── v7-Theta
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

It is possible to enable autodownload for the new binary, but for the purpose of this tutorial, the setup instructions will include how to do this manually. For more information on autodownload with Cosmovisor, see the full docs on [setting up Cosmosvisor](https://github.com/cosmos/cosmos-sdk/blob/master/cosmovisor/README.md).

The following script installs, configures and starts Cosmovisor:

```
# Install Cosmovisor
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor

> NOTE: If you ran a full node on a previous testnet, please skip to [Upgrading From Previous Testnet](#upgrading-from-previous-testnet).

To start a new node, the mainnet instructions apply:

- [Join the mainnet](./join-mainnet.md)
- [Deploy a validator](../validators/validator-setup.md)

The only difference is the SDK version and genesis file. See the [testnet repo](https://github.com/cosmos/testnets) for information on testnets, including the correct version of the Cosmos-SDK to use and details about the genesis file.

## Upgrading Your Node

These instructions are for full nodes that have ran on previous versions of and would like to upgrade to the latest testnet.

When the chain reaches the upgrade block height, the chain will halt and you will have to download the new binary and move it to the correct folder. For the `Theta` upgrade, this would look like:
```
# Prepare Theta upgrade directory
mkdir -p ~/.gaia/cosmovisor/upgrades/Theta/bin

# Download and install the new binary version.
cd $HOME/gaia
git pull
git checkout <upgrade-release>
make install

# Move the new binary to the Theta upgrade directory
cp $GOPATH/bin/gaiad ~/.gaia/cosmovisor/upgrades/Theta/bin
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

### Blocksync
Blocksync will require navigating the Theta upgrade either via [Cosmovisor](#using-cosmovisor) or manually.

Manually updating `gaiad` will require stopping the chain and installing the new binary once it halts at the expected block height (some time on March 17, TBA).

Logs will show `ERR UPGRADE "Theta" NEEDED at height: XXXX`. Stop `gaiad` and run the following:

```
cd $HOME/gaia
git checkout <theta release candidate>
make install
```

::: tip
_NOTE_: If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

Note we use `master` here since it contains the latest stable release.
See the [testnet repo](https://github.com/cosmos/testnets) for details on which version is needed for which testnet, and the [Gaia release page](https://github.com/cosmos/gaia/releases) for details on each release.

Once the new binary is installed, restart the Gaia daemon. Logs will show `INF applying upgrade "Theta" at height: XXXXX`. After a few minutes, the node will start syncing blocks.
