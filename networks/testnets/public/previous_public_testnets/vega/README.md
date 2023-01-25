# Cosmos-Hub Vega Upgrade Public Testnet Instructions

We are running a public testnet for the Vega upgrade. All Cosmos Hub validators are invited to participate in the public testnet. In addition, we are providing a few public endpoints for testing integrations.

We are also available on the `#validators-verified` channel in the [Cosmos Developers discord](https://discord.gg/cosmosnetwork) for support.

## Joining as a validator
You can continue using the same account and validator key as the one you're using for `cosmoshub-4`. If you'd like to **request testnet ATOMs from the faucet** for delegation, please make an issue with your delegator account information.

**Using Keplr:** Use [this jsfiddle](https://jsfiddle.net/qkmecjz2/) to add the `vega-testnet` chain to your Keplr browser extension.

## Schedule 🗓️ 


| Date                       | Testnet plan                |
| -------------------------- | --------------------------- |
| November 8 2021  | ✅ Launch testnet chain with Gaia v5 (previous version)  |
| November 11 2021 | ✅ Submit software upgrade proposal            |
| November 12 2021  | ✅ Voting ends                 |
| November 12 2021    | ✅ Vega upgrade (Gaia v6-rc3) is live on the testnet |

**The Cosmos Hub Testnet for the Vega upgrade successfully passed the upgrade proposal at 12 Nov 2021, 5:09:48pm UTC which resulted in a network upgrade at height 7,453,750 which was at 12 Nov 2021, 20:25:32 UTC.**

## Configuring your full node 🎛️

There are two ways recommended ways sync to the current block height

**1. Syncing without snapshots:**

This is the conventional method. It can take a long time (10-15 hours) because you sync block by block. To follow this method: launch your node with `gaia v5.0.x` and start syncing with the network. Your node will halt at block height `7,453,750` as per the [testnet upgrade plan](https://vega-explorer.hypha.coop/proposals/54). At this point you should upgrade your binary to the Vega release candidate, i.e., `gaia v6.0.0-rc3`. 

**2. Using state-sync to sync with snapshots:** 

This method can allow you to sync very quickly (~10 mins) with snapshots, but your database will be missing historical state. To follow this method: launch your node with `gaia v6.0.0-rc3` and enable state-sync in your `config.toml` file. For detailed instructions see the section below on [using state-sync](#using-state-sync).

### Public testnet Chain-ID

`vega-testnet`

### Genesis file

We're using a modified exported genesis file from `cosmoshub-4` where we take control over the Coinbase custody, Binance, and Certus One validator accounts to be able to continue producing blocks. You can either use the [replace_ref.sh](replace_ref.sh) script to modify the [genesis file here](../exported_unmodified_genesis.json.gz) or you can download our prepared, modified genesis file from [here](modified_genesis_public_testnet/genesis.json.gz)

The `sha256sum` for the modified genesis file is `89d1cb03d1dbe4eb803319f36f119651457de85246e185d6588a88e9ae83f386`.

### Peers and endpoints

| Node              | Node ID                                    | Public IP      | Ports                                                 |
| ----------------- | ------------------------------------------ | -------------- | ----------------------------------------------------- |
| HYPHA "Coinbase"    | `99b04a4efd48846f654da25532c85bd1fa6a6a39` | `198.50.215.1` | p2p: `46656`, rpc: `46657`, api: `4317`, grpc: `4090` |
| HYPHA "Certus-one"  | `1edc806e29bfb380dc0298ce4fded8e3e8554e2a` | `198.50.215.1` | p2p: `36656`, rpc: `36657`, api: `3327`, grpc: `3080` |
| Interchain "Binance" Sentry | `66a9e52e207c8257b791ff714d29100813e2fa00` | `143.244.151.9` | p2p: `26656 `, rpc: `26657 ` , api: `1317 `, grpc: `9090` |

### Minimum gas
Please use `minimum-gas-prices = 0.001uatom` in your `app.toml`

### Invariant checks

Please run with the `--x-crisis-skip-assert-invariants` flag. If you do check for invariants, you may see `rounding error withdrawing rewards from validator`. These are expected.

## Doing the upgrade 

To use Cosmovisor to manage your upgrade, please follow the [Cosmovisor instructions in the README for the local testnet](../local-testnet/README.md#Cosmovisor).

Make sure your machine is resourced with 16GB while performing the upgrade. The upgrade process is memory intensive.

**Note about auto-downloads:** If validators would like to enable the auto-download option (which we don't recommend), and they are currently running an application using Cosmos SDK v0.42, they will need to use Cosmovisor v0.1. Later versions of Cosmovisor do not support Cosmos SDK v0.42 or earlier if the auto-download option is enabled. Please note that with v0.1 you could face node hanging issues with your API server enabled as explained in this [issue](https://github.com/cosmos/cosmos-sdk/issues/9875). If your node is running on darwin, arm64 CPU architecture, please do not use cosmovisor auto-download because the upgrade proposal will not provide the binary download link for this architecture. Please prepare the new binary yourself from this [v6.0.0-rc3 tag](https://github.com/cosmos/gaia/tree/v6.0.0-rc3).

## Using state-sync

We're serving snapshots every 1000 blocks from the following three nodes. Their p2p listen addresses are (`<node-id>@<public-ip>:<port>`):

* `5303f0b47c98727cd7b19965c73b39ce115d3958@134.122.35.247:26656`
* `9e1e3ce30f22083f04ea157e287d338cf20482cf@165.22.235.50:26656`
* `b7feb9619bef083e3a3e86925824f023c252745b@143.198.41.219:26656`

Add these to your persistent_peers list to help your nodes discover snapshots quickly. To enable snapshot discovery, you'll need to configure the `[statesync]` section of your `config.toml` file. You'll need to set `enable = true`, set a `trust_height`, a corresponding `trust_hash` (easily found on a block explorer), and at least two trusted RPC servers that your node will use to cross-check hashes. A reccommended trusted block height is "current height - 1000." Note than in the future, the RPC server requirement will be deprecated as state sync is [moved to the p2p layer in Tendermint 0.35](https://github.com/tendermint/tendermint/issues/6491).

### Serving your own snapshots

If you'd like to contribute your own snapshots, please configure your `snapshot-interval` to a value greater than 0 in your `app.toml` file. It is highly recommended that node operators use the same value for `snapshot-interval` in order to aid snapshot discovery. Discovery is easier when more nodes are serving the same snapshots.

**Recommended configuration**
```
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 1000

# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

## Example setup

This example script shows how you might configure a full node for a testnet, setting up cosmovisor, and starting a cosmovisor service using systemd. To use it:

* copy below script into a file like `run_all.sh`
* edit the configuration variables at the top for the testnet you're running. You may only need to edit the `NODE_MONIKER`
* make it executable `chmod +x run_all.sh`
* run like `./run_all.sh`
* note that I run using `bash -i`, i.e., bash in interactive mode. This is so that the `source ~/.bashrc` command sets env vars in the current shell and not a separate shell

**Using state-sync**

You can configure this script to use state sync to catch up directly to the tip of the chain, withouth having to start with a previous version of gaiad. We're currently serving snapshots every 1000 blocks via three peers that have been added to the `PERSISTENT_PEERS` list in the configuration.

You may need to edit the `TRUST_HEIGHT` and the corresponding `TRUST_HASH` to be closer to the current tip.

To disable state sync, just set `STATE_SYNC` to `false`.

**Not using state-sync**

If you're not using state-sync, you'll have to set the gaiad version to the previous release, sync to the upgrade height, upgrade the binary, and then sync to the tip. Expect this process to take a long time.

⚠️ *This script is provided only as a guideline. It is meant to be adapted. Please do not attempt to run the script without understanding what it's doing.* ⚠️

```bash=
#!/bin/bash -i

##### CONFIGURATION ###

export GAIA_BRANCH=release/v6.0.0-rc3
export GENESIS_ZIPPED_URL=https://github.com/cosmos/vega-test/raw/master/public-testnet/modified_genesis_public_testnet/genesis.json.gz
export NODE_HOME=~/.gaia
export CHAIN_ID=vega-testnet
export NODE_MONIKER=my-state-synced-node # only really need to change this one
export BINARY=gaiad
export PERSISTENT_PEERS="5303f0b47c98727cd7b19965c73b39ce115d3958@134.122.35.247:26656,9e1e3ce30f22083f04ea157e287d338cf20482cf@165.22.235.50:26656,b7feb9619bef083e3a3e86925824f023c252745b@143.198.41.219:26656"

##### OPTIONAL STATE SYNC CONFIGURATION ###

export STATE_SYNC=true
export TRUST_HEIGHT=7834500
export TRUST_HASH="B3A6A0158A1DF235BD49B6FC3670EA621460219D4DC145FA30E754B7AD0DC537"
export SYNC_RPC="134.122.35.247:26657,165.22.235.50:26657"

# you shouldn't need to edit anything below this

echo "Updating apt-get..."
sudo apt-get update

echo "Getting essentials..."
sudo apt-get install git build-essential

echo "Installing go..."
wget -q -O - https://git.io/vQhTU | bash -s - --version 1.15

echo "Sourcing bashrc to get go in our path..."
source /root/.bashrc

echo "Getting gaia..."
git clone https://github.com/cosmos/gaia.git

echo "cd into gaia..."
cd gaia

echo "checkout gaia branch..."
git checkout $GAIA_BRANCH

echo "building gaia..."
make install
echo "***********************"
echo "INSTALLED GAIAD VERSION"
gaiad version
echo "***********************"

cd ..
echo "getting genesis file"
wget $GENESIS_ZIPPED_URL
gunzip genesis.json.gz 

echo "configuring chain..."
$BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME
$BINARY config broadcast-mode block --home $NODE_HOME
$BINARY init $NODE_MONIKER --home $NODE_HOME --chain-id=$CHAIN_ID

if $STATE_SYNC; then
    echo "enabling state sync..."
    sed -i -e '/enable =/ s/= .*/= true/' $NODE_HOME/config/config.toml
    sed -i -e "/trust_height =/ s/= .*/= $TRUST_HEIGHT/" $NODE_HOME/config/config.toml
    sed -i -e "/trust_hash =/ s/= .*/= \"$TRUST_HASH\"/" $NODE_HOME/config/config.toml
    sed -i -e "/rpc_servers =/ s/= .*/= \"$SYNC_RPC\"/" $NODE_HOME/config/config.toml
else
    echo "disabling state sync..."
fi

echo "copying over genesis file..."
cp genesis.json $NODE_HOME/config/genesis.json

echo "setup cosmovisor dirs..."
mkdir -p $NODE_HOME/cosmovisor/genesis/bin

echo "copy binary over..."
cp $(which gaiad) $NODE_HOME/cosmovisor/genesis/bin

echo "re-export binary"
export BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad

echo "install cosmovisor"
export GO111MODULE=on
go get github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor

echo "setup systemctl"
touch /etc/systemd/system/$NODE_MONIKER.service

echo "[Unit]"                               >> /etc/systemd/system/$NODE_MONIKER.service
echo "Description=cosmovisor-$NODE_MONIKER" >> /etc/systemd/system/$NODE_MONIKER.service
echo "After=network-online.target"          >> /etc/systemd/system/$NODE_MONIKER.service
echo ""                                     >> /etc/systemd/system/$NODE_MONIKER.service
echo "[Service]"                            >> /etc/systemd/system/$NODE_MONIKER.service
echo "User=root"                        >> /etc/systemd/system/$NODE_MONIKER.service
echo "ExecStart=/root/go/bin/cosmovisor start --x-crisis-skip-assert-invariants --home \$DAEMON_HOME --p2p.persistent_peers $PERSISTENT_PEERS" >> /etc/systemd/system/$NODE_MONIKER.service
echo "Restart=always"                       >> /etc/systemd/system/$NODE_MONIKER.service
echo "RestartSec=3"                         >> /etc/systemd/system/$NODE_MONIKER.service
echo "LimitNOFILE=4096"                     >> /etc/systemd/system/$NODE_MONIKER.service
echo "Environment='DAEMON_NAME=gaiad'"      >> /etc/systemd/system/$NODE_MONIKER.service
echo "Environment='DAEMON_HOME=$NODE_HOME'" >> /etc/systemd/system/$NODE_MONIKER.service
echo "Environment='DAEMON_ALLOW_DOWNLOAD_BINARIES=true'" >> /etc/systemd/system/$NODE_MONIKER.service
echo "Environment='DAEMON_RESTART_AFTER_UPGRADE=true'" >> /etc/systemd/system/$NODE_MONIKER.service
echo "Environment='DAEMON_LOG_BUFFER_SIZE=512'" >> /etc/systemd/system/$NODE_MONIKER.service
echo ""                                     >> /etc/systemd/system/$NODE_MONIKER.service
echo "[Install]"                            >> /etc/systemd/system/$NODE_MONIKER.service
echo "WantedBy=multi-user.target"           >> /etc/systemd/system/$NODE_MONIKER.service

echo "reload systemd..."
sudo systemctl daemon-reload

echo "starting the daemon..."
sudo systemctl start $NODE_MONIKER.service

echo "***********************"
echo "find logs like this:"
echo "sudo journalctl -fu $NODE_MONIKER.service"
echo "***********************"
```
