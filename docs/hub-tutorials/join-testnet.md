---
order: 3
title: Joining Testnet
---

# Join the Cosmos Hub Public Testnet

This tutorial will provide all necessary instructions for joining the current public testnet. If you're interested in more advanced configuration and synchronization options, see [Join Mainnet](./join-mainnet.md) for a detailed walkthrough.

* Current Version: v9-Lambda
* Chain ID: `theta-testnet-001`

## Background

The Cosmos Hub Public Testnet is currently running Gaia v8. Visit the [testnet explorer](https://explorer.theta-testnet.polypore.xyz/) to view all on-chain activity.

For those who just need instructions on performing the upgrade, see the [Upgrading Your Node](#upgrading-your-node) section.

### Version History

The table below shows all past and upcoming versions of the public testnet.

|  Release   | Upgrade Block Height |    Upgrade Date     |
|:----------:|:--------------------:|:-------------------:|
| v9.0.0-rc3 |      14,476,206      |     2023-02-08      |
| v8.0.0-rc3 |      14,175,595      |     2023-01-20      |
| v7.0.0-rc0 |      9,283,650       |     2022-03-17      |
|   v6.0.0   |       Genesis        | Launched 2022-03-10 |

See the [Gaia release page](https://github.com/cosmos/gaia/releases) for details on each release.

## How to Join

We offer three ways to set up a node in the testnet:

* Quickstart scripts
  * The [testnets](https://github.com/cosmos/testnets/tree/master/public#bash-script) repo has shell scripts to set up a node with a single command.
* Ansible playbooks
  * The [cosmos-ansible](https://github.com/hyphacoop/cosmos-ansible#-quick-start) repo has an inventory file to set up a node with a single command.
* Step-by-step instructions
  * The rest of this document provides a step-by-step walkthrough for setting up a testnet node.

We recommend running public testnet nodes on machines with at least 8 cores, 16GB of RAM, and 300GB of disk space.

## Sync Options

There are two ways to sync a testnet node, Fastsync and State Sync.

* [Fast Sync](https://docs.tendermint.com/v0.34/tendermint-core/fast-sync.html) syncs the chain from genesis by downloading blocks in parallel and then verifying them.
* [State Sync](https://docs.tendermint.com/v0.34/tendermint-core/state-sync.html) will look for snapshots from peers at a trusted height and then verifying a minimal set of snapshot chunks against the network.

State Sync is far faster and more efficient than Fast Sync, but Fast Sync offers higher data integrity and more robust history. For those who are concerned about storage and costs, State Sync can be the better option as it minimizes storage usage when rebuilding initial state.

## Step-by-Step Setup

The following set of instructions assumes you are logged in as root.
* You can run the relevant commands from a sudoer account.
* The `/root/` part in service file paths can be changed to `/home/<username>/`.

### Build Tools

Install build tools and Go.
```shell
sudo apt-get update
sudo apt-get install -y make gcc
wget https://go.dev/dl/go1.18.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Installation & Configuration

You will need to install and configure the Gaia binary using the script below. The Cosmos Hub Public Testnet is running Gaia [`v9.0.0-rc6`](https://github.com/cosmos/gaia/releases/tag/v9.0.0-rc6).

* For up-to-date endpoints like seeds and state sync RPC servers, visit the [testnets repository](https://github.com/cosmos/testnets/tree/master/public).

Build the gaiad binary and initialize the chain home folder.
```
cd $HOME
git clone https://github.com/cosmos/gaia
cd gaia
# To sync from genesis, comment out the next line.
git checkout v9.0.0-rc6
# To sync from genesis, uncomment the next line and skip the State Sync Setup section.
# git checkout v6.0.4
make install
export PATH=$PATH:$HOME/go/bin
gaiad init <custom_moniker>
```

Prepare the genesis file.
```
cd $HOME
wget https://github.com/cosmos/testnets/raw/master/public/genesis.json.gz
gzip -d genesis.json.gz
mv genesis.json $HOME/.gaia/config/genesis.json

# Set minimum gas price & peers
cd $HOME/.gaia/config
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025uatom"/' app.toml
sed -i 's/seeds = ""/seeds = "639d50339d7045436c756a042906b9a69970913f@seed-01.theta-testnet.polypore.xyz:26656,3e506472683ceb7ed75c1578d092c79785c27857@seed-02.theta-testnet.polypore.xyz:26656"/' config.toml
```

#### State Sync Setup (Recommended)

State sync requires you to configure a trust height and trust hash. These depend on the current block height, so they will vary depending on when you are joining the network.

* Visit a [testnet explorer](https://explorer.theta-testnet.polypore.xyz/) to find the block and hash for the current height - 1000.
* Set these parameters in the code snippet below: `<BLOCK_HEIGHT>` and `<BLOCK_HASH>`.


```
cd $HOME/.gaia/config
sed -i 's/enable = false/enable = true/' config.toml
sed -i 's/trust_height = 0/trust_height = <BLOCK_HEIGHT>/' config.toml
sed -i 's/trust_hash = ""/trust_hash = "<BLOCK_HASH>"/' config.toml
sed -i 's/rpc_servers = ""/rpc_servers = "http:\/\/state-sync-01.theta-testnet.polypore.xyz:26657,http:\/\/state-sync-02.theta-testnet.polypore.xyz:26657"/' config.toml
```

* For example, if the block explorer lists a current block height of 12,563,326, we could use a trust height of [12,562,000](https://explorer.theta-testnet.polypore.xyz/blocks/12562000) and the trust hash would be `6F958861E1FA409639C8F2DA899D09B9F50A66DBBD49CE021A2FF680FA8A9204`.

### Cosmovisor Setup (Optional)

Cosmovisor is a process manager that monitors the governance module for incoming chain upgrade proposals. When a proposal is approved, Cosmovisor can automatically download the new binary, stop the chain binary when it hits the upgrade height, switch to the new binary, and restart the daemon. Cosmovisor can be used with either Fast Sync or State Sync. 

The instructions below provide a simple way to sync via Cosmovisor. For more information on configuration, check out the Cosmos SDK's [Cosmovisor documentation](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor).

Cosmovisor requires the creation of the following directory structure:

```shell
.gaia
└── cosmovisor
    └── genesis
        └── bin
            └── gaiad
```

Install Cosmovisor and copy Gaia binary to genesis folder
```
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.3.0
mkdir -p ~/.gaia/cosmovisor/genesis/bin
cp ~/go/bin/gaiad ~/.gaia/cosmovisor/genesis/bin/
```

### Create Service File

* Cosmos Hub recommends running `gaiad` or `cosmovisor` with the `--x-crisis-skip-assert-invariants` flag. If checking for invariants, operators are likely to see `rounding error withdrawing rewards from validator`. These are expected. For more information see [Verify Mainnet](./join-mainnet.md#verify-mainnet).


Create one of the following service files.

If you are not using Cosmovisor: `/etc/systemd/system/gaiad.service`
```
[Unit]
Description=Gaia service
After=network-online.target

[Service]
User=root
ExecStart=/root/go/bin/gaiad start --x-crisis-skip-assert-invariants --home /root/.gaia
Restart=no
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

If you are using Cosmovisor: `/etc/systemd/system/cosmovisor.service`
```
[Unit]
Description=Cosmovisor service
After=network-online.target

[Service]
User=root
ExecStart=/root/go/bin/cosmovisor run start --x-crisis-skip-assert-invariants --home /root/.gaia
Restart=no
LimitNOFILE=4096
Environment='DAEMON_NAME=gaiad'
Environment='DAEMON_HOME=/root/.gaia'
Environment='DAEMON_ALLOW_DOWNLOAD_BINARIES=true'
Environment='DAEMON_RESTART_AFTER_UPGRADE=true'
Environment='DAEMON_LOG_BUFFER_SIZE=512'
Environment='UNSAFE_SKIP_BACKUP=true'

[Install]
WantedBy=multi-user.target
```

### Start the Service

Reload the systemd manager configuration.
```
systemctl daemon-reload
systemctl restart systemd-journald
```

If you are not using Cosmovisor:
```
systemctl enable gaiad.service
systemctl start gaiad.service
```

If you are using Cosmovisor:
```
systemctl enable cosmovisor.service
systemctl start cosmovisor.service
```

To follow the service log, run `journalctl -fu gaiad` or `journalctl -fu cosmovisor`.

* If you are using State Sync, the chain will start syncing once a snapshot is found and verified. Syncing to the current block height should take less than half an hour.
* If you are using Fast Sync, the chain will start syncing once the first block after genesis is found among the peers. **Syncing to the current block height will take several days**.

## Create a Validator (Optional)

If you want to create a validator in the testnet, request tokens through the [faucet Discord channel](https://discord.com/channels/669268347736686612/953697793476821092) and follow the [Running a validator](../validators/validator-setup.md) instructions provided for mainnet.

## Upgrading Your Node

Follow these instructions if you have a node that is already synced and wish to participate in a scheduled testnet software upgrade.

When the chain reaches the upgrade block height specified by a software upgrade proposal, the chain binary will halt and expect the new binary to be run (the system log will show `ERR UPGRADE "<Upgrade name>" NEEDED at height: XXXX` or something similar).

There are three ways you can update the binary:
1. Without Cosmovisor: You must build or download the new binary ahead of the upgrade. When the chain binary halts at the upgrade height:
  * Stop the gaiad service with `systemctl stop gaiad.service`.
  * Build or download the new binary, replacing the existing `~/go/bin` one.
  * Start the gaiad service with `systemctl start gaiad.service`.
2. With Cosmovisor: You must build or download the new binary and copy it to the appropriate folder ahead of the upgrade.
3. With Cosmovisor: Using the auto-download feature, assuming the proposal includes the binaries for your system architecture.

The instructions below are for option 2. For more information on auto-download with Cosmovisor, see the relevant [documentation](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor#auto-download) in the Cosmos SDK repo.

If the environment variable `DAEMON_ALLOW_DOWNLOAD_BINARIES` is set to `false`, Cosmovisor will look for the new binary in a folder that matches the name of the upgrade specified in the software upgrade proposal. For the `v9-Lambda` upgrade, the expected folder structure would look as follows:

```shell
.gaia
└── cosmovisor
    ├── current
    ├── genesis
    │   └── bin
    |       └── gaiad
    └── upgrades
        └── v9-lambda
            └── bin
                └── gaiad
```

> Note: for Cosmovisor v1.0.0, the upgrade name folder is not lowercased (use `cosmovisor/upgrades/v9-Lambda/bin` instead)

Prepare the upgrade directory
```
mkdir -p ~/.gaia/cosmovisor/upgrades/v8-rho/bin
```

Download and install the new binary version.
```
cd $HOME/gaia
git pull
git checkout v8.0.0
make install

# Copy the new binary to the v8-Rho upgrade directory
cp ~/go/bin/gaiad ~/.gaia/cosmovisor/upgrades/v9-lambda/bin/gaiad
```

When the upgrade height is reached, Cosmovisor will stop the gaiad binary, copy the new binary to the `current/bin` folder and restart. After a few minutes, the node should start syncing blocks using the new binary.
