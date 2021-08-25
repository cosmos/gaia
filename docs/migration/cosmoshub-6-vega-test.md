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


Change the genesis file

We have prepared a genesis file which was obtained by `gaiad export` on cosmoshub-5 network at height 7368387. Uncompress this genesis file and use it as the genesis data to mock the comoshub-5 upgrade.

```shell
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
cp priv_validator_key.json ~/.gaia/config/priv_validator_key.json
```
Add this validator to genesis file 
We can add our created validator to the genesis file by replacing address and "pub_key" of one validator(name: Umbrella) which is already in the genesis file.

```shell
export $GENESIS=~/.gaia/config/genesis.json
# change the chain-id to test
sed -i '' 's%"chain_id": "cosmoshub-4",%"chain_id": "test",%g' $GENESIS
# change the consensus pub_key value 
sed -i '' 's%z/Dg9WU/rlIB+LaQVMMHW/a7rvalfIcyz3VdOwfvguc=%/BcD6ZbLvQY29Tx6QJckzHkqvZu/4MfsO12h4a6bSh0%g' $GENESIS
# change the concensus key address 
sed -i '' 's%EBED694E6CE1224FB1E8A2DD8EE63A38568B1E2B%94BFF2A5382CA04897142B4C5B8605B3532F5F7E%g' $GENESIS
# change the validator_address
sed -i '' 's%cosmosvalcons1a0kkjnnvuy3ylv0g5twcae368ptgk83tyalw6t%cosmosvalcons1jjll9ffc9jsy39c59dx9hps9kdfj7hm7w63d0c%g' $GENESIS
# change user account
sed -i '' 's%cosmos1qf7rj85uflxlq2pth5wgst2y9k95ky5zehqeue%cosmos1padpexf2eg9txfkluc5kela53z7g9l333examx%g' $GENESIS
sed -i '' 's%Ah+xi5KEr3N4e8QfYyfP4dRI3MwdCg7mhlMHWjUlspUu%AqJQlj3TwtUkaXpAY+9sahzFCFnbNlbQ36tVqFDAEsLg%g' $GENESIS
```

Config the validator to have over 67% voting power
```shell
# fix the delegation amount to be over 67%
sed -i '' 's%"618372700353.000000000000000000"%"10000000000000618372700353.000000000000000000"%g' $GENESIS

# fix power of the validator
sed -i '' 's%"power": "618372"%"power": "60618372"%g' $GENESIS

# fix last_total_power
sed -i '' 's%"194616038"%"6194616038"%g' $GENESIS
# fix total supply of uatom
sed -i '' 's%277834757180509%1000277834757180509%g' $GENESIS

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
















