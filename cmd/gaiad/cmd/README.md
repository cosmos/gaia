# Testnet command extensions

`gaiad` extends testnet commands with `unsafe-start-local-validator` command that should be used only for testing.

The command makes changes to the local mainnet node to make it suitable for local testing. The changes include modification of consensus and application states by removing old validator data and injecting the new one, and funding the addresses to be used in testing without affecting existing addresses.

The command is added as a sub-command of the `gaiad testnet` command.

## Building a local testnet binary

The gaia binary will contain the testnet extensions only if the `unsafe_start_local_validator` build tags is used.

```shell
make build BUILD_TAGS="unsafe_start_local_validator"
```

## CLI usage
Example of running the command:

```shell
./gaiad testnet unsafe-start-local-validator \
--validator-operator="cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r" \
--validator-pubkey="SLpHEfzQHuuNO9J1BB/hXyiH6c1NmpoIVQ2pMWmyctE=" \
--validator-privkey="AiayvI2px5CZVl/uOGmacfFjcIBoyk3Oa2JPBO6zEcdIukcR/NAe64070nUEH+FfKIfpzU2amghVDakxabJy0Q==" \
--accounts-to-fund="cosmos1ju6tlfclulxumtt2kglvnxduj5d93a64r5czge,cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl"  
```

## Local setup
```shell
gaiad init localnet-export
gaiad keys add test-key --keyring-backend test --keyring-dir ~/.gaia
# get --validator-operator
export VAL_ACC_ADDR=$(gaiad keys show test-key --home ~/.gaia --keyring-backend test --output json | jq .address -r)
gaiad keys parse $(gaiad keys parse $VAL_ACC_ADDR --output=json | jq .bytes -r) --output json | jq .
{
  "formats": [
    "cosmos1738fdeqcepf9mrpdwyrl9zvhlmf2jk4t2x3jwd",
    "cosmospub1738fdeqcepf9mrpdwyrl9zvhlmf2jk4t4qaphg",
    "cosmosvaloper1738fdeqcepf9mrpdwyrl9zvhlmf2jk4t0j98z7", # --> take this one
    "cosmosvaloperpub1738fdeqcepf9mrpdwyrl9zvhlmf2jk4ta3uyqm",
    "cosmosvalcons1738fdeqcepf9mrpdwyrl9zvhlmf2jk4tmpkmwl",
    "cosmosvalconspub1738fdeqcepf9mrpdwyrl9zvhlmf2jk4tv48v22"
  ]
}

# --validator-pubkey and --validator-privkey are in`$HOME/.gaia/config/priv_validator.json
cat $HOME/.gaia/config/priv_validator_key.json
{
  "address": "067CC9545EC0CD744C44D611E3E8857D69E9CAD4",
  "pub_key": {
    "type": "tendermint/PubKeyEd25519",
    "value": "0zorKvPxmVUSBRl49julqrBu69mu6U6+V4GxC8fvxcM="  # --validator-pubkey
  },
  "priv_key": {
    "type": "tendermint/PrivKeyEd25519",
    "value": "rfZBWExZtNzrLx+cy8lMPXQFowXO7AZc5FXBeOyvSdDTOisq8/GZVRIFGXj2O6WqsG7r2a7pTr5XgbELx+/Fww==" # --validator-privkey
  }
}
```

You will also need to change your chain-id in `genesis.json` to match the chain-id of the network you are using in testing (e.g. `cosmoshub-4`).
This is because `gaiad init` creates a `genesis.json` with a testing chain id.

## Example usecase

1. download a mainnet node snapshot
2. replace all validator key files (keyring data, `priv_validator_key.json`, values in `priv_validator_state.json` are reset to 0...)
3. run `gaiad testnet unsafe-start-local-validator` -> switches the validator set and starts the node

## Optional Cleanup for Log Readability

It's recommended to delete the contents of the data/cs.wal folder (from the mainnet node snapshot) before running the unsafe-start-local-validator command. This folder stores messages used for replaying, which are no longer needed since a new block will be created with the new validator setup. If not deleted, the logs may contain misleading errors related to the old state. While this deletion is not mandatory, it can help improve log readability and reduce confusion during testing.