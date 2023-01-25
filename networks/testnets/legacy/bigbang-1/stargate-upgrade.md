# The Stargate Upgrade

Stargate upgrade is to test the upgrade path from `cosmos-sdk@0.39.x` to `stargate` release. The current `bigbang-1` testnet 
is using `cosmos-sdk@0.39.1` and the upgrade will change the underlying SDK version to `cosmos-sdk@v0.40.0-rc1`.

## `stargate` - Software upgrade proposal
**Proposal ID:** 1

**Title:** Stargate Upgrade Proposal

**Description:** Stargate upgrade proposal is to upgrade the current Bigbang testnet to use stargate version of the software. 
Stargate compatible version is available on *bigbang* branch of the Akash Network repo (https://github.com/ovrclk/akash/tree/bigbang). 
This upgrade focuses on testing the upgrade path from 0.39 to stargate release candidate (i.e.cosmos-sdk@v0.40.0-rc1). In case of any upgrade issues, 
network will continue with old binary (akash@v0.8.1) using upgrade module's `--unsafe-skip-upgrades <height>` flag. Here height will be the height of 
the network before upgrade + 1.

**Upgrade Time:** 30th Oct 2020, 1500UTC

**Explorer Links:** https://bigbang.aneka.io/proposals/1 https://bigbang.bigdipper.live/proposals/1 https://look.ping.pub/#/governance/1?chain=bigbang-1

**Voting Period Start Time:** 26th Oct 2020, 1500UTC 

**Voting Period End Time:** 28th Oct 2020, 1500UTC 

## How to Vote for the proposal

```sh
akashctl tx gov vote 1 yes --from <your_key>
```
Note: Though there are different options for voting, it is advised to vote "yes" as this proposal is about testing the upgrade path itself.
 
## How to upgrade 
We will be using `cosmovisor` to perform an automatic software upgrade from `bigbang-1` to a stargate release candidate.

**Note**: Building the `bigbang` binary requires GO version 1.15+

#### Installing cosmovisor

```
mkdir -p $GOPATH/src/github.com/cosmos
cd $GOPATH/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk.git && cd cosmos-sdk/cosmovisor
make cosmovisor
cp cosmovisor $GOBIN/cosmovisor
```

#### Setting up directories

```
mkdir -p ~/.akashd/cosmovisor
mkdir -p ~/.akashd/cosmovisor/genesis/bin
mkdir -p ~/.akashd/cosmovisor/upgrades/stargate/bin
cp $GOBIN/akashd ~/.akashd/cosmovisor/genesis/bin
```

#### Building the stargate release

```
cd $GOPATH/src/github.com/ovrclk/akash
git fetch -a && git checkout bigbang
make all
```

This will create `akashd` binary built on stargate release. This binary has to be placed in the upgrades folder.
```
cp akashd ~/.akashd/cosmovisor/upgrades/stargate/bin
```

#### Setting up service file

**Note**: Using cosmovisor for automatic upgrade requires it to be set up as a service file.
Create a systemd file:
`sudo nano /lib/systemd/system/cosmovisor.service`
Copy-Paste in the following and update `<your_username>` and `<go_workspace>` as required:

```
[Unit]
Description=Cosmovisor daemon
After=network-online.target

[Service]
Environment="DAEMON_NAME=akashd"
Environment="DAEMON_HOME=/home/<your_username>/.akashd"
Environment="DAEMON_RESTART_AFTER_UPGRADE=on"
User=<your_username>
ExecStart=/home/<your_username>/<go_workspace>/bin/cosmovisor start
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

#### Enable the cosmovisor process
```
sudo systemctl daemon-reload
sudo systemctl enable cosmovisor.service
```

#### Stop the existing Akashd service file and start the Cosmovisor service.

```
sudo systemctl stop akashd.service
sudo systemctl start cosmovisor.service
```

You can see the logs using:
```
journalctl -u cosmovisor -f
```


## What if the upgrade fails?

If the scheduled upgrade fails, the network will resume with old software (akash@0.8.1). The validators would need to `skip` the upgrade in-order to 
make the old binary work. Following command will skip the upgrade and start the validator node with old binary. This requires akash@0.8.1 aka old binary

### General way to skip the upgrade (without cosmovisor/systemd)

```sh
akashd start --unsafe-skip-upgrades <height>
```
`height` here will be the next block height of the network (i.e., halt-height + 1).

Note: If you are using systemd (and not cosmovisor), You can edit the systemd file  and update the ExecBinary command as above. Once after restarting your .

### Upgrade failure contingency using cosmovisor

Due to ongoing development of Stargate release, it is possible the `bigbang` Stargate release candidate might have issues which might prevent the network from restarting. In this case the planned upgrade will have to be cancelled and the network will continue on the `v0.8.1` binary using the `--unsafe-skip-upgrades` flag.

**Note**: The following procedure has to be undertaken only if the planned upgrade fails. Please co-ordinate on the #bigbang-testnet channel on Discord to see if this procedure is necessary or not.

#### Stop the Cosmovisor service

```
sudo systemctl stop cosmovisor.service
sudo systemctl disable cosmovisor.service
```

#### Edit the original Akashd service file to include the flag
We cannot use cosmovisor for skipping upgrades. So, need to use original binary.

```
sudo nano /lib/systemd/system/akashd.service
```

Please co-ordinate on the #bigbang-testnet channel on Discord to see what `<upgrade-height>` will be.


```
[Unit]
Description=akash
After=network-online.target

[Service]
User=<your_username>
ExecStart=/home/<your_username>/<go_workspace>/bin/akashd start --unsafe-skip-upgrades <upgrade-height>
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

#### Start the Akashd service

```
sudo systemctl daemon-reload
sudo systemctl restart akashd.service
```
