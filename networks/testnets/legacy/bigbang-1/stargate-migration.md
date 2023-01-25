# Bigbang State Migration and Upgrade

We will be performing a state export of the current ```bigbang``` testnet and upgrade to **Stargate** phase of the testnet on Fri, 27 Nov 2020 1500UTC.
 
### Build the fixed binary:
**Note**: Building the fixed binary section can be skipped if a validator has participated in the previous migration attempt. Due to a minor bug in akash@v0.8.1, exporting state by passing the flag `--height` is not possible. It is recommended for all the validators to upgrade their nodes to [fixed](https://github.com/ovrclk/akash/tree/boz/mainnet/prevent-double-init) version of the binary prior to the upgrade.

```
cd $GOPATH/src/github.com/ovrclk/akash
git fetch -a && git checkout boz/mainnet/prevent-double-init
git pull origin boz/mainnet/prevent-double-init
make install
```
This will create the fixed version of `akashd` and place it automatically in your `$GOBIN`.

### Set the halt-time:
To ensure all the validators stop at the right time, it's recommended for all to edit the `app.toml` file with the right halt-time.
```
nano ~/.akashd/config/app.toml
```
Replace the default value of `halt-time` with `1606489200`. This is time in unix for Fri, 27 Nov 2020 1500UTC. Restart the node with the fixed binary with and `app.toml` modifications.
```
sudo systemctl restart akashd
```

### Migration steps

The node will stop automatically on Fri, 27 Nov 2020 1500UTC if the above steps are performed. Stop your current `akashd` process:
```
sudo systemctl stop akashd
```
State export of the current chain will be taken using this command:
**Note**: Please execute the follwing command only when the node stops on the pre-determined time. 
```
akashd export --for-zero-height --height <last-commit-height> > bigbang_genesis_export.json
```
Please co-ordinate on **#bigbang-testnet** channel to get the right `<last-commit-height>`.

The `export` command will generate the file `bigbang_genesis_export.json` with the state export. Verify the sha256 sum of the file
```
jq -S -c -M '' bigbang_genesis_export.json | shasum -a 256
```
Hash of the file: **a7cc66b7acde92b6a49354e9535a23a23d51afd15e76d0d2e473b4b77c27c4bb  -**

If you do not have the mnemonic of your key saved please export your private key using the following command:
```
akashctl keys export <key-name>
```
Save the contents of the output to a file and save it.

#### Build the bigbang binary
```
cd $GOPATH/src/github.com/ovrclk/akash
git fetch -a && git checkout bigbang
git pull origin bigbang
make install
```
Verify the version
```
akashd version --long
name: akashd
server_name: akashd
version: 0.7.9-rc14-12-g6a16ab5
commit: 6a16ab5e7d3fb7f850adbaa25079724328373c44
build_tags: netgo,ledger,mainnet
go: go version go1.15.5 linux/amd64
```

With the stargate release binary of `akashd`, migrate the genesis to `v0.40`.
```
akashd migrate v0.40 bigbang_genesis_export.json --chain-id bigbang-2 > new_v40_genesis.json
```

After migration it is important to add `ibc` module parameters to the genesis file. To add the ibc module to your genesis use the following command:
```
cat new_v40_genesis.json | jq '.app_state |= . + {"ibc":{"client_genesis":{"clients":[],"clients_consensus":[],"create_localhost":false},"connection_genesis":{"connections":[],"client_connection_paths":[]},"channel_genesis":{"channels":[],"acknowledgements":[],"commitments":[],"receipts":[],"send_sequences":[],"recv_sequences":[],"ack_sequences":[]}},"transfer":{"port_id":"transfer","denom_traces":[],"params":{"send_enabled":false,"receive_enabled":false}},"capability":{"index":"1","owners":[]}}' > genesis_with_ibc_state.json
```

Verify the sha256 sum of the new genesis file

```
jq -S -c -M '' genesis_with_ibc_state.json | shasum -a 256
```
Hash of the file: **3b6bd610a9b3e6ed142b078268084c8b8de8e26a0a76eb77a69294b04bddfd3e  -**

Replace the original genesis with the new genesis:

```
cp genesis_with_ibc_state.json ~/.akashd/config/genesis.json
```

Due to changes in SDK, `app.toml` will have to be updated/replaced to support gRPC & telemetry configurations. Sample app.toml is provided in the repo. To download it:

```
curl https://raw.githubusercontent.com/cosmos/testnets/master/bigbang-1/app.toml > $HOME/.akashd/config/app.toml
```

New genesis file will be pushed to this repo for validators after being generated.

#### Reset the state and restart:

**Note**: It is recommended to take a backup of the data before wiping the state in case of a rollback.
```
akashd unsafe-reset-all
```

Start `akashd` process with the new genesis file:
```
sudo systemctl start akashd
``` 

For the `stargate` release, client and daemon binaries have been merged into one entity. Validator account key has to be imported again to `akashd` now.
```
akashd keys add <key-name> --recover
```

Paste your mnemonic after the prompt to restore your key.

If you don't have the `seed`, you can export the private key using:
`akashctl keys export <keyname>`

Save the contents to a file and use that to import into `akashd` using the following command:
```
akashd keys import <key-name> <file-name>
```
