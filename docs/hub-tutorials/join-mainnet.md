# Join the Cosmos Hub Mainnet

The current Cosmos Hub mainnet`cosmoshub-4`is operating after its most recent [Vega upgrade `V6.0.0`](https://github.com/cosmos/gaia/releases/tag/v6.0.0). This guide includes instructions for joining the mainnet either as an archive/full node or a pruned node. For instructions to join as a validator, please also see the [validator instructions](https://hub.cosmos.network/main/validators/overview.html#).

### Overview

- [Explorers](#Explorers)
- [Getting Started](#Getting-Started)
- [General Configuration](#General-Configuration)
    - [Initialize Chain](#Initialize-Chain)
    - [Genesis File](#Genesis-File)
    - [Seeds & Peers](#Seeds-amp-Peers)
    - [Gas & Fees](#Gas-amp-Fees)
    - [Pruning of State](#Pruning-of-State)
    - [REST API](#REST-API)
    - [GRPC](#GRPC)
- [Choosing a Sync Mode](#Choosing-a-Sync-Mode)
    - [Sync from Scratch (Full/Archive Node)](#Sync-from-scratch)
    - [State Sync](#State-Sync)
    - [Quicksync](#Quicksync)
- [Snapshots](#Snapshots)
- [Hardware Requirements](#Hardware)
- [Releases](#Releases-amp-Upgrades)
- [Cosmovisor](#Cosmovisor)
- [Running via Background Process](#Running-via-Background-Process)
- [Exporting State](#Exporting-State)
- [Verify Mainnet](#Verify-Mainnet)

### Background

The current Cosmos Hub mainnet `cosmoshub-4` has been performing in place store migration upgrades as of the [Delta Upgrade](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-delta-upgrade.md) July 2021. The most recent upgrade was [Vega](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-vega-upgrade.md) December 2021. This type of upgrade preserves the same chain-id but is otherwise similar to a traditional hub upgrade, meaning historical state is lost after the upgrade takes place (still accessible of course by a v4.2.1 full node). Visit the [migration section](https://github.com/cosmos/gaia/tree/main/docs/migration) of the Hub's docs for more information on previous chain migrations.


## Explorers

There are many explorers for the Cosmos Hub. For reference while setting up a node, here are a few recommendations:

- [Mintscan](https://www.mintscan.io/cosmos)
- [Big Dipper](https://cosmos.bigdipper.live/)
- [Hubble](https://hubble.figment.io/cosmos/chains/cosmoshub-4)


## Getting Started

Before syncing a node, make sure the following prerequisites are completed:
- Choose the proper hardware/server configuration. See the [hardware guide](#hardware).
- Ensure Gaia is properly installed. See the [installation guide](https://hub.cosmos.network/main/getting-started/installation.html) for a walkthrough.
- Compile the Gaia binary. The version of Gaia will depend on which way the node is synced. See [releases](#releases) for more information on the appropriate version to build.
- Follow the [configuration guide](#General-Configuration) to intialize and prepare the node to sync with the network.


## General Configuration

Before choosing a method of syncing a node, make sure to walk through the basic setup and configuration. Operators will need to initialize gaiad, download the genesis file for `cosmoshub-4`, and set persistent peers and/or seeds for startup.

### Initialize Chain

Choose a custom moniker for the node and initialize. By default, the `init` command creates the `~/.gaia` directory with subfolders `config` and `data`. In the `/config` directory, the most important files for configuration are `app.toml` and `config.toml`.
```bash
gaiad init <custom-moniker>
```

> **Note**: Monikers can contain only ASCII characters. Using Unicode characters is not supported and renders the node unreachable.

The `moniker` can be edited in the `~/.gaia/config/config.toml` file:

```
# A custom human readable name for this node
moniker = "<custom_moniker>"
```

### Genesis File

Once the node is initialized, download the genesis file and move to the `/config` directory of the Gaia home directory.
```bash
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json
```

### Seeds & Peers

Upon startup the node will need to connect to peers. If there are specific nodes a node operator is interested in setting as seeds or as persistent peers, this can be configured in `$HOME/.gaia/config/config.toml`

```
# Comma separated list of seed nodes to connect to
seeds = "<seed node id 1>@<seed node address 1>:26656,<seed node id 2>@<seed node address 2>:26656"

# Comma separated list of nodes to keep persistent connections to
persistent_peers = "<node id 1>@<node address 1>:26656,<node id 2>@<node address 2>:26656"
```

Alternatively, node operators can also download the [Quicksync address book]( https://quicksync.io/addrbook.cosmos.json). Make sure this is moved to `~/.gaia/config/addrbook.json`.

### Gas & Fees

On Cosmos Hub mainnet, the accepted denom is `uatom`, where `1atom = 1.000.000uatom`

Transactions on the Cosmos Hub network need to include a transaction fee in order to be processed. This fee pays for the gas required to run the transaction. The formula is the following:

```
fees = ceil(gas * gasPrices)
```

The `gas` is dependent on the transaction. Different transaction require different amount of `gas`. The `gas` amount for a transaction is calculated as it is being processed, but there is a way to estimate it beforehand by using the `auto` value for the `gas` flag. Of course, this only gives an estimate. This estimate is adjustable with the flag `--gas-adjustment` (default `1.0`) to ensure enough `gas` is provided for the transaction.

The `gasPrice` is the price of each unit of `gas`. Each validator sets a `min-gas-price` value, and will only include transactions that have a `gasPrice` greater than their `min-gas-price`.

The transaction `fees` are the product of `gas` and `gasPrice`. The higher the `gasPrice`/`fees`, the higher the chance that your transaction will get included in a block.

**For mainnet, the recommended `gas-prices` is `0.0025uatom`.**

A full-node keeps unconfirmed transactions in its mempool. In order to protect it from spam, it is better to set a `minimum-gas-prices` that the transaction must meet in order to be accepted in your node's mempool. This parameter can be set in the following file `~/.gaia/config/app.toml`.

```
# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.25token1;0.0001token2).
minimum-gas-prices = "0.0025uatom"
```

The initial recommended `min-gas-prices` is `0.0025uatom`, but this can be changed later.

### Pruning of State

There are four strategies for pruning state. These strategies apply only to state and do not apply to block storage. A node operator may want to consider custom pruning if node storage is a concern or there is an interest in running an archive node.

To set pruning, adjust the `pruning` parameter in the `~/.gaia/config/app.toml` file.
The following pruning state settings are available:

1. `everything`: Prune all saved states other than the current state.
2. `nothing`: Save all states and delete nothing.
3. `default`: Save the last 100 states and the state of every 10,000th block.
4. `custom`: Specify pruning settings with the `pruning-keep-recent`, `pruning-keep-every`, and `pruning-interval` parameters.

By default, every node is in `default` mode which is the recommended setting for most environments.
If a node operator wants to change their node's pruning strategy then this **must** be done before the node is initialized.

In `~/.gaia/config/app.toml`
```
# default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
# nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
# everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
# custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'
pruning = "custom"

# These are applied if and only if the pruning strategy is custom.
pruning-keep-recent = "10"
pruning-keep-every = "1000"
pruning-interval = "10"
```

Passing a flag when starting `gaia` will always override settings in the `app.toml` file, if you would like to change your node to the `everything` mode then you can pass the `---pruning everything` flag when you call `gaiad start`.

> **Note**: When you are pruning state it will not be possible to query the heights that are not in the node's store.

### REST API

By default, the REST API is disabled. To enable the REST API, edit the `~/.gaia/config/app.toml` file, and set `enable` to `true` in the `[api]` section.

```
###############################################################################
###                           API Configuration                             ###
###############################################################################
[api]
# Enable defines if the API server should be enabled.
enable = true
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
# Address defines the API server to listen on.
address = "tcp://0.0.0.0:1317"
```

Optionally, you can activate swagger by setting `swagger` to `true` or change the port of the REST API in the parameter `address`.
After restarting your application, you can access the REST API on `<NODE IP>:1317`.

### GRPC

By default, gRPC is enabled on port `9090`. In the `~/.gaia/config/app.toml` file, you can make changes in the gRPC section. To disable the gRPC endpoint, set `enable` to `false`. To change the port, use the `address` parameter.

```
###############################################################################
###                           gRPC Configuration                            ###
###############################################################################
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```


## Choosing a Sync Mode

There are three main ways to sync a node on the Cosmos Hub. The way a node operator will want to sync is partially dependent on the type of node they will operate. This guide will focus on two types of common nodes; archive/full and pruned. For further information on syncing to run a validator node, see the section on [Validators](https://hub.cosmos.network/main/validators/overview.html).

If a node operator wishes to run a full/archive node, it is possible to start from scratch but will take a significant amount of time to catch up. Node operators not concerned with rebuilding original state from the beginning of `cosmoshub-4` can also leverage [Quicksync](#Quicksync)'s available archive history.

For operators interested in bootstrapping a pruned node, either [Quicksync](#Quicksync) or [State sync](#State-sync) would be sufficient.

Make sure to consult the [hardware](#hardware) section for guidance on the best configuration for the type of node operating.

### Sync from scratch

Syncing from scratch is possible by starting the chain from the beginning of `cosmoshub-4` and either manually upgrading the chain or setting up [Cosmovisor](#Cosmovisor) to upgrade automatically. Node operators syncing from scratch will need to navigate two upgrades, Delta and Vega. For more information on performing the manual upgrades, see [Releases & Upgrades](#Releases-amp=-Upgrades).

Start Gaia to begin syncing with the `skip-invariants` flag. For more information on this see [Verify Mainnet](#Verify-Mainnet).
```bash
gaiad start --x-crisis-skip-assert-invariants

```

The node will begin rebuilding state until it hits the first upgrade height at block `6910000`. If Cosmovisor is set up then there's nothing else to do besides wait, otherwise the node operator will need to perform the manual upgrade twice.

### State Sync

Enabling state sync can be one of the most efficient ways to bootstrap a node. On startup, the node will ask its peers for available snapshots. Once a snapshot has been accepted, it will begin requesting snapshot chunks from peers. Once the snapshot is verified, the node will begin to process blocks in realtime until it catches up with the head of the chain. For more information on state sync, see [Tendermint's documentation](https://docs.tendermint.com/master/spec/p2p/messages/state-sync.html).

To enable state, visit an explorer to get a recent block height and corresponding hash. A node operator can choose any height/hash in the current bonding period, but as the recommended snapshot period is `1000` blocks, it is advised to choose something close to`height - 1000`.

With the block height and hash selected, update the configuration in `~/.gaia/config/config.toml` to set `enable = true`, and populate the `trust_height` and `trust_hash`. Node operators can configure the rpc servers to a preferred provider, but there must be at least two entries. It is important that these are two rpc servers the node operator trusts to verify component parts of the chain state. While not reccommended, uniqueness is not currently enforced, so it is possible to duplicate the same server in the list and still sync successfully.

> **Note**: In the future, the RPC server requirement will be deprecated as state sync is [moved to the p2p layer in Tendermint 0.35](https://github.com/tendermint/tendermint/issues/6491).

```
#######################################################
###         State Sync Configuration Options        ###
#######################################################
[statesync]
# State sync rapidly bootstraps a new node by discovering, fetching, and restoring a state machine
# snapshot from peers instead of fetching and replaying historical blocks. Requires some peers in
# the network to take and serve state machine snapshots. State sync is not attempted if the node
# has any local state (LastBlockHeight > 0). The node will have a truncated block history,
# starting from the height of the snapshot.
enable = true

# RPC servers (comma-separated) for light client verification of the synced state machine and
# retrieval of state data for node bootstrapping. Also needs a trusted height and corresponding
# header hash obtained from a trusted source, and a period during which validators can be trusted.
#
# For Cosmos SDK-based chains, trust_period should usually be about 2/3 of the unbonding time (~2
# weeks) during which they can be financially punished (slashed) for misbehavior.
rpc_servers = "https://rpc-cosmoshub.keplr.app:443,https://rpc.cosmos.network:443"
trust_height = 8959784
trust_hash = "3D8F12EA302AEDA66E80939F7FC785206692F8B6EE6F727F1655F1AFB6A873A5"
trust_period = "168h0m0s"
```

Start Gaia to begin state sync. It may take take some time for the node to acquire a snapshot, but the command and output should look similar to the following:

```bash
$ gaiad start --x-crisis-skip-assert-invariants

...

> INF Discovered new snapshot format=1 hash="0x000..." height=8967000 module=statesync

...

> INF Fetching snapshot chunk chunk=4 format=1 height=8967000 module=statesync total=45
> INF Applied snapshot chunk to ABCI app chunk=0 format=1 height=8967000 module=statesync total=45
```

Once state sync successfully completes, the node will begin to process blocks normally. If state sync fails and the node operator encounters the following error:  `State sync failed err="state sync aborted"`, either try restarting `gaiad` or running `gaiad unsafe-reset-all` (make sure to backup any configuration and history before doing this).

### Quicksync
Quicksync.io offers several  daily snapshots of the Cosmos Hub with varying levels of pruning (`archive` 1.4TB, `default` 540GB, and `pruned` 265GB). For downloads and installation instructions, visit the [Cosmos Quicksync guide](https://quicksync.io/networks/cosmos.html).


## Snapshots
Saving and serving snapshots helps nodes rapidly join the network.

To contribute snapshots to the network, `snapshot-interval` needs to be configured in `~/.gaia/config/app.toml`. The Cosmos Hub recommends setting this value to `1000`. This value must match `pruning-keep-every` in `config.toml`.

In `app.toml`
```
###############################################################################
###                        State Sync Configuration                         ###
###############################################################################

# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]

# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 1000

# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 10
```

It is highly recommended that node operators use the same value for snapshot-interval in order to aid snapshot discovery. Discovery is easier when more nodes are serving the same snapshots.


## Hardware
Running a full archive node can be resource intensive as the full  current `cosmoshub-4` state is over `1.4TB`. For those who wish to run state sync or use quicksync, the following hardware configuration is recommended:

| Node Type     | RAM                   | Storage     |
| -----------   | --------------------- | ----------- |
| Validator     | 16GB (32 Recommended) | 500GB-2TB*  |
| Full/Archive  | 8GB (16 Recommended)  | 2TB         |
| Default       | 8GB (16 Recommended)  | 1TB         |
| Pruned        | 8 GB                  | 500GB       |

\* Storage size for validators will depend on level of pruning.


## Releases & Upgrades
**See all [Gaia Releases](https://github.com/cosmos/gaia/releases)**

The most up to date release of Gaia is [`V6.0.0`](https://github.com/cosmos/gaia/releases/tag/v6.0.0). For those that want to use state sync or quicksync to get their node up to speed, starting with the most recent version of Gaia is sufficient.

To sync an archive or full node from scratch, it is important to note that you must start with [`V4.2.1`](https://github.com/cosmos/gaia/releases/tag/v4.2.1) and proceed through two different upgrades Delta at block height `6910000` and Vega at block height `8695000`.

The process is summarized below but make sure to follow the manual upgrade instructions for each release:

**[Delta Instructions](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-delta-upgrade.md#Upgrade-will-take-place-July-12,-2021)**
Once `V4` reaches the upgrade block height, expect the chain to halt and to see the following message:
```bash
ERR UPGRADE "Gravity-DEX" NEEDED at height: 6910000: v5.0.0-4760cf1f1266accec7a107f440d46d9724c6fd08
```

Make sure to save a backup of `~/.gaia` in case rolling back is necessary.

Install Gaia [`V5.0.0`](https://github.com/cosmos/gaia/releases/tag/v5.0.0) and restart the daemon.


**[Vega Instructions](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-vega-upgrade.md)**

Once `V5` reaches the upgrade block height, the chain will halt and display the following message:
```bash
ERR UPGRADE "Vega" NEEDED at height: 8695000

```

Again, make sure to backup `~/.gaia`

Install Gaia [`V6.0.0`](https://github.com/cosmos/gaia/releases/tag/v6.0.0) and restart the daemon.


## Cosmovisor

Cosmovisor is a process manager developed to relieve node operators of having to manually intervene every time there is an upgrade. Cosmovisor monitors the governance module for upgrade proposals; it will take care of downloading the new binary, stopping the old one, switching to the new one, and restarting.

For more information on how to run a node via Cosmovisor, check out the [docs](https://github.com/cosmos/cosmos-sdk/blob/master/cosmovisor/README.md).


## Running via Background Process

To run the node in a background process with automatic restarts, it's recommended to use a service manager like `systemd`. To set this up run the following:

```bash
sudo tee /etc/systemd/system/<service name>.service > /dev/null <<EOF  
[Unit]
Description=Gaia Daemon
After=network-online.target

[Service]
User=$USER
ExecStart=$(which gaiad) start
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF
```

If using Cosmovisor then make sure to add the following:

```bash
Environment="DAEMON_HOME=$HOME/.gaia"
Environment="DAEMON_NAME=gaiad"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
```

After the `LimitNOFILE` line and replace `$(which gaiad)` with `$(which cosmovisor)`.

Run the following to setup the daemon:

```bash
sudo -S systemctl daemon-reload
sudo -S systemctl enable <service name>
```

Then start the process and confirm that it's running.

```bash
sudo -S systemctl start <service name>

sudo service <service name> status
```


## Exporting State

Gaia can dump the entire application state into a JSON file. This application state dump is useful for manual analysis and can also be used as the genesis file of a new network.

Export state with:

```bash
gaiad export > [filename].json
```

You can also export state from a particular height (at the end of processing the block of that height):

```bash
gaiad export --height [height] > [filename].json
```

If you plan to start a new network from the exported state, export with the `--for-zero-height` flag:

```bash
gaiad export --height [height] --for-zero-height > [filename].json
```


## Verify Mainnet

Help to prevent a catastrophe by running invariants on each block on your full
node. In essence, by running invariants you ensure that the state of mainnet is
the correct expected state. One vital invariant check is that no atoms are
being created or destroyed outside of expected protocol, however there are many
other invariant checks each unique to their respective module. Because invariant checks
are computationally expensive, they are not enabled by default. To run a node with
these checks start your node with the assert-invariants-blockly flag:

```bash
gaiad start --assert-invariants-blockly
```

If an invariant is broken on your node, your node will panic and prompt you to send
a transaction which will halt mainnet. For example the provided message may look like:

```bash
invariant broken:
    loose token invariance:
        pool.NotBondedTokens: 100
        sum of account tokens: 101
    CRITICAL please submit the following transaction:
        gaiad tx crisis invariant-broken staking supply

```
