# Cosmos Hub 6 Vega  Upgrade Test Instruction

This document describes the  procedures to test an upgrade from cosmoshub-5.0.5 to cosmoshub-6.0.0-vega locally.
This upgrade will bring the new release of Cosmos-SDK v0.43 RC and IBC 1.0 RC into gaia.

## Version
- presently running cosmoshub-5.0.5
- going to upgrade to cosmoshub-6-vega.

## Chain Upgrade by Cosmovisor
Initial setup and configuration
```shell
git checkout release/v5.0.5
make install
# Never do unsafe-reset-all in production environment !!!
gaiad unsafe-reset-all
```
Configure the gaiad binary for testing
```shell
gaiad config chain-id test
gaiad config keyring-backend test
gaiad config broadcast-mode block
```
Init the chain
```shell
# Never do overwrite in production environment !!!
gaiad init my-node --chain-id test --overwrite
```

Setup the single Validator

```shell
# Create a key to hold your validator account
gaiad keys add my-account

# Add that key into the genesis.app_state.accounts array in the genesis file
gaiad add-genesis-account $(gaiad keys show my-account -a) 3000000000stake

# Creates your validator
gaiad gentx my-account 1000000000stake --chain-id test

# Add the generated bonding transaction to the genesis file
gaiad collect-gentxs
```

Change the genesis file

We have prepared a genesis file which was obtained by `gaiad export` on cosmoshub-5 network. Uncompress this genesis file and use it as the genesis data to mock the comoshub-5 upgrade.

```shell
wget https://xxxx.json.tar.bz2
tar xvzf xxxx.json.tar.bz2
mv xxxx.json ~/.gaia/config/genesis.json
```
Set the minimum gas price to 0stake in `~/.gaia/config/app.toml`.
```shell
minimum-gas-prices = "0stake"
```

Set the environment variables
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME=$HOME/.gaia
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Create the folder for the genesis binary and copy the gaiad binary (version 5.0.5)
```shell
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $DAEMON_HOME/cosmovisor/genesis/bin
```
Reduce the voting_period in `$HOME/.gaia/config/genesis.json` to 20 seconds (20s):
```shell
cat <<< $(jq '.app_state.gov.voting_params.voting_period = "20s"' $HOME/.gaia/config/genesis.json) > $HOME/.gaia/config/genesis.json
```

Build the new gaia binary
```shell
git checkout start-upgrade
make install
```

Create the folder for the upgrade binary and copy the upgrade gaia binary:
```shell
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/vega/bin
cp $(which gaiad) $DAEMON_HOME/cosmovisor/upgrades/vega/bin
```


Start cosmosvisor:
```shell
cosmovisor start
```
Open a new terminal window and submit an upgrade proposal along with a deposit and a vote. The below commands shows a proposal of upgrade at height 40.
```shell 
cosmovisor tx gov submit-proposal software-upgrade vega --title upgrade --description upgrade --upgrade-height 40 --from my-account --yes
cosmovisor tx gov deposit 1 10000000stake --from my-account --yes
cosmovisor tx gov vote 1 yes --from my-account --yes
```

## Upgrade result

The chain itself will continue to run after the upgrade height. But you can find info: `applying upgrade "vega" at height: 40` upon height 39.


## Reference

[cosmovisor quick start](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)

[changelog of cosmos-sdk v0.43.0](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/CHANGELOG.md#v0430---2021-08-10)

[cosmos/ibc-go v1.0.0](https://github.com/cosmos/ibc-go/tree/v1.0.0)
















