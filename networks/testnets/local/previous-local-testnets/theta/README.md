# v7-Theta Local Testnet

These instructions will help you simulate the `v7-Theta` upgrade on a single validator node testnet as follows:

- Start with gaia version: `v6.0.4`
- After the upgrade: gaia branch `theta-prepare`

We will use a [modified genesis file](genesis.json.gz) during this upgrade. This modified genesis file is similar to the one we are running on the public testnet, and has been modified in part to replace an existing validator (Coinbase Custody) with a new validator account that we control. The account's [mnemonic](mnemonic.txt) and [private validator key](priv_validator_key.json) are provided in this repo.  
For a full list of modifications to the genesis file, please [see below](#genesis-modifications).

If you are interested in running v7-Theta without going through the upgrade, you can checkout gaia branch `release/v7.0.0` in the [Build gaia](#build-gaia) section and follow the rest of the instructions up until the node is running and producing blocks.

## Run a local testnet

### Requirements

Follow the [installation instructions](https://hub.cosmos.network/main/getting-started/installation.html) to understand build requirements. You'll need to install Go 1.17.

```
sudo apt update
sudo apt upgrade
sudo apt install git build-essential

curl -OL https://golang.org/dl/go1.17.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.17.4.linux-amd64.tar.gz
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
git checkout release/v6.0.4
make install
```

### Configure the chain

First initialize your chain.

```
export CHAIN_ID=theta-localnet
export NODE_MONIKER=my-theta-local-validator # whatever you like
export BINARY=gaiad
export NODE_HOME=$HOME/.gaia

$BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME
$BINARY config broadcast-mode block --home $NODE_HOME
$BINARY init $NODE_MONIKER --home $NODE_HOME --chain-id=$CHAIN_ID
```

Then replace the genesis file with our modified genesis file.

```
wget https://github.com/hyphacoop/testnets/raw/add-theta-testnet/v7-theta/local-testnet/genesis.json.gz
gunzip genesis.json.gz 
mv genesis.json $NODE_HOME/config/genesis.json
```

Make sure you have the correct genesis file.

```
shasum -a 256 $NODE_HOME/config/genesis.json
ecfe4f227f5565d2d53f2ad6635b8ca425a76877c85dd2bff035a272777d2b5f
````
Also replace the validator key.

```
wget https://github.com/hyphacoop/testnets/raw/add-theta-testnet/v7-theta/local-testnet/priv_validator_key.json
mv priv_validator_key.json $NODE_HOME/config/priv_validator_key.json
```

Now add your user account. This account has over 75% tokens bonded to your validator.

```
export USER_MNEMONIC="junk appear guide guess bar reject vendor illegal script sting shock afraid detect ginger other theory relief dress develop core pull across hen float"
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
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0
```

Setup the Cosmovisor directory structure. There are two methods to use Cosmovisor:

1. **Manual:** Node runners can manually build the old and new binary and put them into the `cosmovisor` folder (as shown below). Cosmovisor will then switch to the new binary upon upgrade height.

2. **Auto-download:** Allowing Cosmovisor to [auto-download](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#auto-download) the new binary at the upgrade height automatically.

**Cosmovisor directory structure**

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

### Manually prepare the upgrade binary (if you do not have auto-download enabled on Cosmovisor)

Build the upgrade binary.
```
cd $HOME/gaia
git checkout theta-prepare
git pull
go get github.com/cosmos/gaia/v7/app
make install
```

Copy over the v7-Theta binary into the correct directory.
```
mkdir -p $NODE_HOME/cosmovisor/upgrades/v7-Theta/bin
cp $(which gaiad) $NODE_HOME/cosmovisor/upgrades/v7-Theta/bin
export BINARY=$NODE_HOME/cosmovisor/upgrades/v7-Theta/bin/gaiad
```

## Submit and vote on a software upgrade proposal

You can submit a software upgrade proposal without specifiying a binary, but this only works for those nodes who are manually preparing the upgrade binary.

```
cosmovisor tx gov submit-proposal software-upgrade v7-Theta \
--title v7-Theta \
--deposit 100uatom \
--upgrade-height 9035600 \
--upgrade-info "upgrade to v7-Theta" \
--description "upgrade to v7-Theta" \
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
$BINARY tx gov vote 61 yes \
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
$BINARY query gov proposal 61 --home $NODE_HOME
```

After `PROPOSAL_STATUS_PASSED`, wait until the upgrade height is reached Cosmovisor will now auto-download the new binary specific to your platform and apply the upgrade.

Please note, the upgrade info in method II does not contain the download link of the binary for GOOS=darwin GOARCH=arm64 (for Mac M1 users). Please use method I to upgrade.

## Genesis Modifications

Full list of modifications are as follows:

* Swapping chain id to theta-localnet
* Increasing balance of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 175000000000000 uatom
* Increasing supply of uatom by 175000000000000
* Increasing balance of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 1000 theta
* Increasing supply of theta by 1000
* Creating new coin theta valued at 1000
* Increasing balance of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 1000 rho
* Increasing supply of rho by 1000
* Creating new coin rho valued at 1000
* Increasing balance of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 1000 lambda
* Increasing supply of lambda by 1000
* Creating new coin lambda valued at 1000
* Increasing balance of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 1000 epsilon
* Increasing supply of epsilon by 1000
* Creating new coin epsilon valued at 1000
* Increasing balance of cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh by 550000000000000 uatom
* Increasing supply of uatom by 550000000000000
* Increasing delegator stake of cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9 by 550000000000000
* Increasing validator stake of cosmosvaloper1wvvhhfm387xvfnqshmdaunnpujjrdxznxqep2k by 550000000000000
* Increasing validator power of D5AB5E458FD9F9964EF50A80451B6F3922E6A4AA by 550000000
* Swapping min governance deposit amount to 1uatom
* Swapping tally parameter quorum to 0.000000000000000001
* Swapping tally parameter threshold to 0.000000000000000001
* Swapping governance voting period to 60s
* Swapping staking unbonding_time to 1s

Please note that you will need to set `fast-sync` to false in your `config.toml` file and wait for approximately 10mins for a single node testnet to start. This is due to an [issue](https://github.com/osmosis-labs/osmosis/issues/735) with state export based testnets that can't get to consensus without multiple peered nodes.
