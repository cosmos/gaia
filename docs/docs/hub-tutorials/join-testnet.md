---
title: Joining Testnet
order: 3
---

Visit the [testnets repo](https://github.com/cosmos/testnets) for the most up-to-date information on the currently available public testnets:

* Interchain Security (ICS) Testnet: [`provider`](https://github.com/cosmos/testnets/blob/master/interchain-security/provider/README.md)
* Release Testnet: [`theta-testnet-001`](https://github.com/cosmos/testnets/blob/master/release/README.md)

## How to Join

You can set up a testnet node with a single command using one of the options below:

* Run a shell script from the testnets repo
  * [ICS Testnet](https://github.com/cosmos/testnets/tree/master/interchain-security/provider#bash-script)
  * [Release testnet](https://github.com/cosmos/testnets/blob/master/release/README.md#bash-script)
* Run an Ansible playbook from the [cosmos-ansible](https://github.com/hyphacoop/cosmos-ansible) repo
  * [ICS Testnet](https://github.com/hyphacoop/cosmos-ansible/blob/main/examples/README.md#provider-chain)
  * [Release Testnet](https://github.com/hyphacoop/cosmos-ansible/blob/main/examples/README.md#join-the-cosmos-hub-release-testnet)

## Create a Validator (Optional)

If you want to create a validator in either testnet, request tokens through the [faucet Discord channel](https://discord.com/channels/669268347736686612/953697793476821092) and follow the [this guide](https://github.com/cosmos/testnets/blob/master/interchain-security/VALIDATOR_JOINING_GUIDE.md#creating-a-validator-on-the-provider-chain). If you are creating a validator in the Release Testnet, you can disregard the instructions about joining live consumer chains.

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

If the environment variable `DAEMON_ALLOW_DOWNLOAD_BINARIES` is set to `false`, Cosmovisor will look for the new binary in a folder that matches the name of the upgrade specified in the software upgrade proposal.

### Cosmovisor Upgrade Example

Using the `v17` upgrade as an example, the expected folder structure would look as follows:

```shell
.gaia
└── cosmovisor
    ├── current
    ├── genesis
    │   └── bin
    |       └── gaiad
    └── upgrades
        └── v17
            └── bin
                └── gaiad
```

Prepare the upgrade directory

```shell
mkdir -p ~/.gaia/cosmovisor/upgrades/v17/bin
```

Download and install the new binary version.

```shell
cd $HOME/gaia
git pull
git checkout v17.0.0-rc0
make install

# Copy the new binary to the v17 upgrade directory
cp ~/go/bin/gaiad ~/.gaia/cosmovisor/upgrades/v17/bin/gaiad
```

When the upgrade height is reached, Cosmovisor will stop the gaiad binary, update the symlink from `current` to the relevant upgrade folder, and restart. After a few minutes, the node should start syncing blocks using the new binary.