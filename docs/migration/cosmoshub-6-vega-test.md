# Cosmos Hub Vega  Upgrade Test Instruction

This document describes the  procedures to test cosmoshub vega upgrade locally.
This upgrade will bring the new release of Cosmos-SDK v0.43 RC and IBC 1.0 RC into gaia.

## Version
- presently running cosmoshub-4. Gaia version: v5.0.5
- going to upgrade to cosmoshub-5, Gaia version: v6.0.0-vega ???

## Chain Upgrade by Cosmovisor
Initial setup and configuration
```shell
git checkout release/v5.0.5
make install
# Never do unsafe-reset-all in production environment !!!
gaiad unsafe-reset-all
```

[comment]: <> (Configure the gaiad binary for testing)

[comment]: <> (```shell)

[comment]: <> (gaiad config chain-id test)

[comment]: <> (gaiad config keyring-backend test)

[comment]: <> (gaiad config broadcast-mode block)

[comment]: <> (```)

[comment]: <> (Init the chain)

[comment]: <> (```shell)

[comment]: <> (# Never do overwrite in production environment !!!)

[comment]: <> (gaiad init my-node --chain-id test --overwrite)

[comment]: <> (```)

Change the genesis file

We have prepared a genesis file which was obtained by `gaiad export` on cosmoshub-5 network at height 7368387. Uncompress this genesis file and use it as the genesis data to mock the comoshub-5 upgrade.

```shell
wget https://xxxx.json.gz
gunzip exported_genesis_v5.json.gz
cp exported_genesis_v5.json ~/.gaia/config/genesis.json
```
Set the minimum gas price to 0stake in `~/.gaia/config/app.toml`.
```shell
minimum-gas-prices = "0stake"
```

Get the private validator key

We have prepared a validator key (MNEMONIC: "net warfare noise fabric ring eager crumble pioneer assault segment trust bind inform warfare silk cement language kitten ginger stadium divide borrow tail great")
This private validator will be configured to own over 67% of voting power.

Reminder: please do not use this key for your cryptocurrency assets in production environment !!!
```shell
wget https://xxxx.json.gz
cp priv_validator_key.json ~/.gaia/config/priv_validator_key.json
```
Add this validator to genesis file 
We can add our created validator to the genesis file by replacing address and "pub_key" of one validator(name: Umbrella) which is already in the genesis file.

```shell
export $GENESIS = ~/.gaia/config/genesis.json
# change the chain-id to test
sed -i '' 's%"chain_id": "cosmoshub-4",%"chain_id": "test",%g' $GENESIS
# change the pub_key value
sed -i '' 's%z/Dg9WU/rlIB+LaQVMMHW/a7rvalfIcyz3VdOwfvguc=%i8Y6OsPLIbkFXbPAYv3iJJkHscPl4n2IGNw2Q2RO4F0=%g' $GENESIS
# change the address
sed -i '' 's%EBED694E6CE1224FB1E8A2DD8EE63A38568B1E2B%0C619FA4D49086F5341E04BDE3105D5AC0517B47%g' $GENESIS
# change the validator_address
cosmosvaloper1q9p5m5xemu0zln3sh02u5neh8g7zevcq45essg
sed -i '' 's%cosmosvalcons1a0kkjnnvuy3ylv0g5twcae368ptgk83tyalw6t%cosmosvaloper1q9p5m5xemu0zln3sh02u5neh8g7zevcq45essg%g' $EXPORTED_GENESIS

```

Config the validator to have over 67% voting power

[comment]: <> (Setup the Validator)

[comment]: <> (```shell)

[comment]: <> (# Create a key to hold your validator account)

[comment]: <> (gaiad keys add my-account)

[comment]: <> (# Add that key into the genesis.app_state.accounts array in the genesis file)

[comment]: <> (gaiad add-genesis-account $&#40;gaiad keys show my-account -a&#41; 3000000000stake)

[comment]: <> (# Creates your validator)

[comment]: <> (gaiad gentx my-account 1000000000stake --chain-id test)

[comment]: <> (# Add the generated bonding transaction to the genesis file)

[comment]: <> (gaiad collect-gentxs)

[comment]: <> (```)

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
















