<!--
order: 5
-->

# Upgrade Your Node

This document describes the upgrade procedure of a `gaiad` full-node to a new version.

## Cosmovisor

The CosmosSDK provides a convenient process manager that wraps around the `gaiad` binary and can automatically swap in new binaries upon a successful governance upgrade proposal. Cosmovisor is entirely optional but recommended. More information can be found in [cosmos.network docs](https://docs.cosmos.network/master/run-node/cosmovisor.html) and [cosmos-sdk/cosmovisor/readme](https://github.com/cosmos/cosmos-sdk/blob/master/cosmovisor/README.md).

### Setup

To get started with Cosmovisor first download it

```bash
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor
```

Set up the environment variables

```bash
echo "# Setup Cosmovisor" >> ~/.profile
echo "export DAEMON_NAME=gaiad" >> ~/.profile
echo "export DAEMON_HOME=$HOME/.gaia" >> ~/.profile
source ~/.profile
```

Create the appropriate directories

```bash
mkdir -p ~/.gaia/cosmovisor/upgrades
mkdir -p ~/.gaia/cosmovisor/genesis/bin/
cp $(which gaiad) ~/.gaia/cosmovisor/genesis/bin/

# verify the setup. 
# It should return the same version as gaiad
cosmovisor version
```

Now `gaiad` can start by running

```bash
cosmovisor start
```

### Preparing an Upgrade

Cosmovisor will continually poll  the `$DAEMON_HOME/data/upgrade-info.json` for new upgrade instructions. When an upgrade is ready, node operators can download the new binary and place it under `$DAEMON_HOME/cosmovisor/upgrades/<name>/bin` where `<name>` is the URI-encoded name of the upgrade as specified in the upgrade module plan.

It is possible to have Cosmovisor automatically download the new binary. To do this set the following environment variable.

```bash
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
```

## Manual Software Upgrade

First, stop your instance of `gaiad`. Next, upgrade the software:

```bash
cd gaia
git fetch --all && git checkout <new_version>
make install
```

::: tip
_NOTE_: If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

See the [testnet repo](https://github.com/cosmos/testnets) for details on which version is needed for which public testnet, and the [Gaia release page](https://github.com/cosmos/Gaia/releases) for details on each release.

Your full node has been cleanly upgraded! If there are no breaking changes then you can simply restart the node by running:

```bash
gaiad start
```

## Upgrade Genesis File

:::warning
If the new version you are upgrading to has breaking changes, you will have to restart your chain. If it is not breaking, you can skip to [Restart](#restart)
:::

To upgrade the genesis file, you can either fetch it from a trusted source or export it locally.

### Fetching from a Trusted Source

If you are joining the mainnet, fetch the genesis from the [mainnet repo](https://github.com/cosmos/launch). If you are joining a public testnet, fetch the genesis from the appropriate testnet in the [testnet repo](https://github.com/cosmos/testnets). Otherwise, fetch it from your trusted source.

Save the new genesis as `new_genesis.json`. Then replace the old `genesis.json` with `new_genesis.json`

```bash
cd $HOME/.gaia/config
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

Then, go to the [reset data](#reset-data) section.

### Exporting State to a New Genesis Locally

If you were running a node in the previous version of the network and want to build your new genesis locally from a state of this previous network, use the following command:

```bash
cd $HOME/.gaia/config
gaiad export --for-zero-height --height=<export-height> > new_genesis.json
```

The command above take a state at a certain height `<export-height>` and turns it into a new genesis file that can be used to start a new network.

Then, replace the old `genesis.json` with `new_genesis.json`.

```bash
cp -f genesis.json new_genesis.json
mv new_genesis.json genesis.json
```

At this point, you might want to run a script to update the exported genesis into a genesis that is compatible with your new version. For example, the attributes of a the `Account` type changed, a script should query encoded account from the account store, unmarshall them, update their type, re-marshal and re-store them. You can find an example of such script [here](https://github.com/cosmos/cosmos-sdk/blob/02c6c9fafd58da88550ab4d7d494724a477c8a68/contrib/migrate/v0.33.x-to-v0.34.0.py).

## Reset Data

:::warning
If the version <new_version> you are upgrading to is not breaking from the previous one, you should not reset the data. If it is not breaking, you can skip to [Restart](#restart)
:::

::: warning
If you are running a **validator node** on the mainnet, always be careful when doing `gaiad unsafe-reset-all`. You should never use this command if you are not switching `chain-id`.
:::

::: danger IMPORTANT
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to get slashed due to double sign !
:::

First, remove the outdated files and reset the data. **If you are running a validator node, make sure you understand what you are doing before resetting**.

```bash
gaiad unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before, your node will still try to connect to them, but may fail if they haven't also been upgraded.
