# v8-Rho Local Testnet

These instructions will help you simulate the `v8-Rho` upgrade on a single validator node testnet as follows:

- Start with gaia version: `v7.1.0`
- After the upgrade: gaia release `v8.0.0-rc3`

We will use a modified genesis file during this upgrade. This modified genesis file is similar to the one we are running on the public testnet, and has been modified in part to replace an existing validator (Coinbase Custody) with a new validator account that we control. The account's mnemonic, validator key, and node key are provided in this repo.  
For a full list of modifications to the genesis file, please [see below](#genesis-modifications).

If you are interested in running v8-Rho without going through the upgrade, you can download one of the binaries in the Gaia [releases](https://github.com/cosmos/gaia/releases) page follow the rest of the instructions up until the node is running and producing blocks.

* **Chain ID**: `local-testnet`
* **Gaia version:** `v7.1.0`
* **Modified genesis file:** [tinkered-genesis_2023-01-06T20:22:50.460851274Z_v7.1.0_13562370.json.gz](https://files.polypore.xyz/genesis/mainnet-genesis-tinkered/tinkered-genesis_2023-01-06T20%3A22%3A50.460851274Z_v7.1.0_13562370.json.gz)
* **Original genesis file:** [mainnet-genesis_2023-01-06T20:22:50.460851274Z_v7.1.0_13562370.json.gz](https://files.polypore.xyz/genesis/mainnet-genesis-export/mainnet-genesis_2023-01-06T20%3A22%3A50.460851274Z_v7.1.0_13562370.json.gz)
* **Validator key:** [priv_validator_key](priv_validator_key.json)
* **Node key:** [node_key](node_key.json)
* **Validator mnemonic:** [mnemonic.txt](mnemonic.txt)

## Set up with Ansible Playbook

Use the example inventory file from the [cosmos-ansible](https://github.com/hyphacoop/cosmos-ansible) repo to set up a local testnet node:

```
git clone https://github.com/hyphacoop/cosmos-ansible.git
cd cosmos-ansible
ansible-playbook node.yml -i examples/inventory-local-genesis.yml -e 'target=SERVER_IP_OR_DOMAIN'
```

For additional information, visit the [examples page](https://github.com/hyphacoop/cosmos-ansible/tree/main/examples#start-a-local-testnet-using-a-modified-genesis-file).

### Upgrade proposal requirements

Log into the target machine and switch to the `gaia` user with `su gaia`.

```
export NODE_HOME=$HOME/.gaia
export USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
export USER_KEY_NAME=my-validator-account
echo $USER_MNEMONIC | gaiad --home $NODE_HOME keys add $USER_KEY_NAME --recover --keyring-backend=test
```

## Manual setup

### Requirements

Follow the [installation instructions](https://hub.cosmos.network/main/getting-started/installation.html) to understand build requirements. You'll need to install Go 1.19.

```
sudo apt update
sudo apt upgrade
sudo apt install git build-essential

curl -OL https://golang.org/dl/go1.19.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.14.4.linux-amd64.tar.gz
```

### Modify your paths
```
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> ~/.profile
source ~/.profile
```

### Build gaia 

```
cd $HOME
git clone https://github.com/cosmos/gaia.git
cd gaia
git checkout v7.1.0
make install
```

### Configure the chain

First initialize your chain.

```
export CHAIN_ID=local-testnet
export NODE_MONIKER=my-local-validator # whatever you like
export BINARY=gaiad
export NODE_HOME=$HOME/.gaia

$BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME
$BINARY config broadcast-mode block --home $NODE_HOME
$BINARY init $NODE_MONIKER --home $NODE_HOME --chain-id=$CHAIN_ID
```

Then replace the genesis file with our modified genesis file.

```
wget https://files.polypore.xyz/genesis/mainnet-genesis-tinkered/tinkered-genesis_2023-01-06T20%3A22%3A50.460851274Z_v7.1.0_13562370.json.gz
gunzip tinkered-genesis_2023-01-06T20:22:50.460851274Z_v7.1.0_13562370.json.gz
mv tinkered-genesis_2023-01-06T20:22:50.460851274Z_v7.1.0_13562370.json $NODE_HOME/config/genesis.json
```

Make sure you have the correct genesis file.

```
shasum -a 256 $NODE_HOME/config/genesis.json
0aafe5e71241c79d2beb542314de7c7416a77a452231ee1a38378b3c0e04dc38
````

Replace the validator and node keys.

```
wget https://raw.githubusercontent.com/cosmos/testnets/master/local/priv_validator_key.json
mv priv_validator_key.json $NODE_HOME/config/priv_validator_key.json
wget https://raw.githubusercontent.com/cosmos/testnets/master/local/node_key.json
mv node_key.json $NODE_HOME/config/node_key.json
```

Now add your user account. This account has over 75% tokens bonded to your validator.

```
export USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
export USER_KEY_NAME=my-validator-account
echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add $USER_KEY_NAME --recover --keyring-backend=test
```

Set minimum gas prices.

```
sed -i -e 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025uatom"/g' $NODE_HOME/config/app.toml
```

Set block sync to be false. This allow us to achieve liveness without additional peers. See this [issue](https://github.com/osmosis-labs/osmosis/issues/735) for details.

```
sed -i -e '/fast_sync =/ s/= .*/= false/' $NODE_HOME/config/config.toml
```

### Cosmovisor

First download Cosmovisor.

```
export GO111MODULE=on
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.3.0
```

Setup the Cosmovisor directory structure. There are two methods to use Cosmovisor:

1. **Manual:** Node runners can manually build the old and new binary and put them into the `cosmovisor` folder (as shown below). Cosmovisor will then switch to the new binary upon upgrade height. **If you are using Cosmovisor `v1.0.0`, the version name is not lowercased (use `upgrades/v8-Rho/bin/gaiad`)**.

2. **Auto-download:** Allowing Cosmovisor to [auto-download](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#auto-download) the new binary at the upgrade height automatically.

**Cosmovisor directory structure**

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad
└── upgrades
    └── v8-rho
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

For both methods, you should first start by creating the genesis directory as well as copying over the starting binary.

```
mkdir -p $NODE_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $NODE_HOME/cosmovisor/genesis/bin
export BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
```

We recommend running Cosmovisor as a systemd service. Here's how to create the service:

```
touch /etc/systemd/system/$NODE_MONIKER.service

echo "[Unit]"                               >> /etc/systemd/system/$NODE_MONIKER.service
echo "Description=cosmovisor-$NODE_MONIKER" >> /etc/systemd/system/$NODE_MONIKER.service
echo "After=network-online.target"          >> /etc/systemd/system/$NODE_MONIKER.service
echo ""                                     >> /etc/systemd/system/$NODE_MONIKER.service
echo "[Service]"                            >> /etc/systemd/system/$NODE_MONIKER.service
echo "User=root"                        >> /etc/systemd/system/$NODE_MONIKER.service
echo "ExecStart=/root/go/bin/cosmovisor start --x-crisis-skip-assert-invariants" >> /etc/systemd/system/$NODE_MONIKER.service
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
```

Set the following environment variables for the Cosmovisor service:

```
export DAEMON_NAME=gaiad
export DAEMON_HOME=$NODE_HOME
```

Before running the service, we recommend reloading the systemctl daemon and restarting the journald service.

```
sudo systemctl daemon-reload
sudo systemctl restart systemd-journald
```

### Run your node

You are now ready to start your node like this:

```
sudo systemctl enable $NODE_MONIKER.service
sudo systemctl start $NODE_MONIKER.service
```

And view the logs like this:

```
sudo journalctl -fu $NODE_MONIKER.service
```

**Please make sure your node is running and producing blocks before you proceed further!** It can take up to 10 minutes for your node to start up. Once it's producing blocks you'll start seeing log messages like the following:

```
INF committed state app_hash=99D509C03FDDFEACAD90608008942C0B4C801151BDC1B8998EEC69A1772B22DF height=9060257 module=state num_txs=0
```

## Manually prepare the upgrade binary (if you do not have auto-download enabled on Cosmovisor)

Build the upgrade binary.
```
cd $HOME/gaia
git checkout v8.0.0-rc3
git pull
make install
```

Copy over the v8-Rho binary into the correct directory. **If you are using Cosmovisor `v1.0.0`, the version name is not lowercased (use `upgrades/v8-Rho/bin/gaiad`)**.
```
mkdir -p $NODE_HOME/cosmovisor/upgrades/v8-rho/bin
cp $(which gaiad) $NODE_HOME/cosmovisor/upgrades/v8-rho/bin
export BINARY=$NODE_HOME/cosmovisor/upgrades/v8-rho/bin/gaiad
```

## Submit and vote on a software upgrade proposal

You can submit a software upgrade proposal without specifiying a binary, but this only works for those nodes who are manually preparing the upgrade binary.

```
cosmovisor tx gov submit-proposal software-upgrade v8-Rho \
--title v8-Rho \
--deposit 100uatom \
--upgrade-height TBD \
--upgrade-info "upgrade to v8-Rho" \
--description "upgrade to v8-Rho" \
--gas auto \
--fees 400uatom \
--from $USER_KEY_NAME \
--keyring-backend test \
--chain-id $CHAIN_ID \
--home $NODE_HOME \
--node tcp://localhost:26657 \
--yes
```

Vote on it.

```
gaiad tx gov vote 95 yes \
--from $USER_KEY_NAME \
--keyring-backend test \
--chain-id $CHAIN_ID \
--home $NODE_HOME \
--gas auto \
--fees 400uatom \
--node tcp://localhost:26657 \
--yes
```

After the voting period ends, you should be able to query the proposal to see if it has passed. Like this:

```
$gaiad query gov proposal 95 --home $NODE_HOME
```

After `PROPOSAL_STATUS_PASSED`, wait until the upgrade height is reached Cosmovisor will now auto-download the new binary specific to your platform and apply the upgrade.

Please note, the upgrade info in method II does not contain the download link of the binary for GOOS=darwin GOARCH=arm64 (for Mac M1 users). Please use method I to upgrade.

## Genesis Modifications

Full list of modifications are as follows:

* Swapping chain id to local-testnet
* Increasing balance of cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl by 175000000000000 uatom
* Increasing supply of uatom by 175000000000000
* Increasing balance of cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl by 550000000000000 uatom
* Increasing supply of uatom by 550000000000000
* Increasing delegator stake of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 550000000000000
* Increasing validator stake of cosmosvaloper1r5v5srda7xfth3hn2s26txvrcrntldju7lnwmv by 550000000000000
* Increasing validator power of 973C48DF8B3356C45E44494723A6E0D45DEB8131 by 550000000
* Swapping min governance deposit amount to 1uatom
* Swapping tally parameter quorum to 0.000000000000000001
* Swapping tally parameter threshold to 0.000000000000000001
* Swapping governance voting period to 60s
* Swapping staking unbonding_time to 1s

Please note that you will need to set `fast-sync` to false in your `config.toml` file and wait for approximately 10mins for a single node testnet to start. This is due to an [issue](https://github.com/osmosis-labs/osmosis/issues/735) with state export based testnets that can't get to consensus without multiple peered nodes.
