# Upgrade Gaia from v14.1.0 to v14.2.0

## This is a security upgrade. IT IS CONSENSUS BREAKING, so please apply the fix only on height 19460500.

### Release Details
* https://github.com/cosmos/gaia/releases/tag/v14.2.0
* Chain upgrade height : `19460500`. Exact upgrade time can be checked [here](https://www.mintscan.io/cosmos/block/19460500).
* Go version has been frozen at `1.20`. If you are going to build `gaiad` binary from source, make sure you are using the right GO version!

# Performing the co-ordinated upgrade

This co-ordinated upgrades requires validators to stop their validators at `halt-height`, switch their binary to `v14.2.0` and restart their nodes with the new version.

The exact sequence of steps depends on your configuration. Please take care to modify your configuration appropriately if your setup is not included in the instructions.

# Manual steps

## Step 1: Configure `halt-height` and restart the node.

This upgrade requires `gaiad` halting execution at a pre-selected `halt-height`. Failing to stop at `halt-height` may cause a consensus failure during chain execution at a later time.

There are two mutually exclusive options for this stage:

### Option 1: Set the halt height by modifying `app.toml`

* Stop the gaiad process.

* Edit the application configuration file at `~/.gaia/config/app.toml` so that `halt-height` reflects the upgrade plan:

```toml
# Note: Commitment of state will be attempted on the corresponding block.
halt-height = 19460500
```
* restart gaiad process

* Wait for the upgrade height and confirm that the node has halted

### Option 2: Restart the `gaiad` binary with command line flags

* Stop the gaiad process.

* Do not modify `app.toml`. Restart the `gaiad` process with the flag `--halt-height`:
```shell
gaiad start --halt-height 19460500
```

* Wait for the upgrade height and confirm that the node has halted

Upon reaching the `halt-height` you need to replace the `v14.1.0` gaiad binary with the new `gaiad v14.2.0` binary and remove the `halt-height` constraint.
Depending on your setup, you may need to set `halt-height = 0` in your `app.toml` before resuming operations.
```shell
   git clone https://github.com/cosmos/gaia.git
```

### Build and start the binary

```shell
cd $HOME/gaia
git pull
git fetch --tags
git checkout v14.2.0
make install

# verify install
gaiad version
# v14.2.0
```

```shell
gaiad start # starts the v14.2.0 node
```

## Cosmovisor steps

### Prerequisite: Alter systemd service configuration

Disable automatic restart of the node service. To do so please alter your `gaiad.service` file configuration and set appropriate lines to following values.

```
Restart=no 

Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=false"
```

After that you will need to run `sudo systemctl daemon-reload` to apply changes in the service configuration.

There is no need to restart the node yet; these changes will get applied during the node restart in the next step.

# Setup Cosmovisor
## Create the updated gaiad binary of v14.2.0

### Go to gaiad directory if present else clone the repository

```shell
   git clone https://github.com/cosmos/gaia.git
```

### Follow these steps if gaiad repo already present

```shell
   cd $HOME/.gaia
   git pull
   git fetch --tags
   git checkout v14.2.0
   make install
```

### Check the new gaiad version, verify the latest commit hash
```shell
   $ gaiad version --long
   name: gaiad
   server_name: gaiad
   version: 14.2.0
   commit: <commit-hash>
   ...
```

### Or check checksum of the binary if you decided to download it

```shell
$ shasum -a 256 gaiad-v14.2.0-linux-amd64
<checksum>  gaiad-v14.2.0-linux-amd64
```

## Copy the new gaiad (v14.2.0) binary to cosmovisor current directory
```shell
   cp $GOPATH/bin/gaiad ~/.gaiad/cosmovisor/current/bin
```

## Restore service file settings

If you are using a service file, restore the previous `Restart` settings in your service file: 
```
Restart=On-failure 
```
Reload the service control `sudo systemctl daemon-reload`.

# Revert `gaiad` configurations

Depending on which path you chose for Step 1, either:

* Reset `halt-height = 0` option in the `app.toml` or
* Remove it from start parameters of the gaiad binary and start node again