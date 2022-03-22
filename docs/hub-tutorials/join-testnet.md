<!--
order: 4
-->

# Join the Public Testnet

::: tip Current Testnet
See the [testnet repo](https://github.com/cosmos/testnets) for
information on the latest testnet, including the correct version
of Gaia to use and details about the genesis file.
:::

::: warning
**You need to [install gaia](./getting-started/installation.md) before you go further**
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

### Reset Data

First, remove the outdated files and reset the data.

```bash
rm $HOME/.gaia/config/addrbook.json $HOME/.gaia/config/genesis.json
gaiad unsafe-reset-all
```

Your node is now in a pristine state while keeping the original `priv_validator.json` and `config.toml`. If you had any sentry nodes or full nodes setup before,
your node will still try to connect to them, but may fail if they haven't also
been upgraded.

::: danger Warning
Make sure that every node has a unique `priv_validator.json`. Do not copy the `priv_validator.json` from an old node to multiple new nodes. Running two nodes with the same `priv_validator.json` will cause you to double sign.
:::

### Software Upgrade

Now it is time to upgrade the software:

```bash
git clone https://github.com/cosmos/gaia.git
cd gaia
git fetch --all && git checkout master
make install
```

::: tip
_NOTE_: If you have issues at this step, please check that you have the latest stable version of GO installed.
:::

Note we use `master` here since it contains the latest stable release.
See the [testnet repo](https://github.com/cosmos/testnets) for details on which version is needed for which testnet, and the [Gaia release page](https://github.com/cosmos/gaia/releases) for details on each release.

Your full node has been cleanly upgraded!
