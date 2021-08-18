# Cosmos Hub 6 Vega  Upgrade Test Instructions

This document describes the  procedures of how to test an upgrade from cosmoshub-5.0.5 to cosmoshub-6.0.0-vega locally.

This upgrade will bring the new release of Cosmos-SDK v0.43 RC and IBC 1.0 RC into gaia. Cosmo
After this upgrade, gaia will unlock two new modules, `x/feegrant` and `x/authz`. We will test this two modules functions in this upgrade tests!

# Version
- presently running cosmoshub-5.0.5
- going to upgrade to cosmoshub-6-vega.

# Chain Upgrade by Cosmovisor
Initial setup and configuration
```shell
git checkout release/v5.0.5
make install
# Never do unsafe-reset-all in production environment !!!
gaiad unsafe-reset-all
```
Configure the gaiad binary for testing:
```shell
gaiad config chain-id test
gaiad config keyring-backend test
gaiad config broadcast-mode block
```
Init the chain
```shell
# Never do unsafe-reset-all in production environment !!!

gaiad init my-node --overwrite
```

Setup the Validator

```shell
# Create a key to hold your validator account
gaiad keys add my-account

# Add that key into the genesis.app_state.accounts array in the genesis file
gaiad add-genesis-account $(gaiad keys show my-account -a) 3000000000stake,1000000000validatortoken

# Creates your validator
gaiad gentx my-account 1000000000stake --chain-id test

# Add the generated bonding transaction to the genesis file
gaiad collect-gentxs
```
Set the minimum gas price to 0stake in ~/.gaia/config/app.toml: ??? maybe not needed

wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json

```shell
minimum-gas-prices = "0stake"
```
Set the environment variables:
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME=$HOME/.gaia
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Create the folder for the genesis binary and copy the gaiad binary:
```shell
mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $DAEMON_HOME/cosmovisor/genesis/bin
```
Change voting_period in $HOME/.gaia/config/genesis.json to a reduced time of 20 seconds (20s):
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
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/test1/bin
cp $(which gaiad) $DAEMON_HOME/cosmovisor/upgrades/vega/bin
```


Start cosmosvisor:
```shell
cosmovisor start
```
Open a new terminal window and submit an upgrade proposal along with a deposit and a vote:
```shell 
cosmovisor tx gov submit-proposal software-upgrade vega --title upgrade --description upgrade --upgrade-height 40 --from my-account --yes
cosmovisor tx gov deposit 1 10000000stake --from my-account --yes
cosmovisor tx gov vote 1 yes --from my-account --yes
```

Result:




