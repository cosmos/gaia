# Testnet command extensions

`gaiad` extends testnet commands with `unsafe-start-local-validator` command that should be used only for testing.

The command makes changes to the local mainnet node to make it suitable for local testing. The changes include modification of consensus and application states by removing old validator data and injecting the new one, and funding the addresses to be used in testing without affecting existing addresses.

The command is added as a sub-command of the `gaiad testnet` command.

## Building a local testnet binary

The gaia binary will cointain the testnet extensions only if the `unsafe_start_local_validator` build tags is used.

```shell
make build BUILD_TAGS="-tag unsafe_start_local_validator
```

## CLI usage
Example of running the command:

```shell
./gaiad testnet unsafe-start-local-validator  
--validator-operator="cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r"  
--validator-pukey="SLpHEfzQHuuNO9J1BB/hXyiH6c1NmpoIVQ2pMWmyctE=" 
--validator-privkey="AiayvI2px5CZVl/uOGmacfFjcIBoyk3Oa2JPBO6zEcdIukcR/NAe64070nUEH+FfKIfpzU2amghVDakxabJy0Q=="  
--accounts-to-fund="cosmos1ju6tlfclulxumtt2kglvnxduj5d93a64r5czge,cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl"  
[other gaiad start flags]
```

## Example usecase

1. download a mainnet node snapshot
2. replace all validator key files (keyring data, `priv_validator_key.json`, values in `priv_validator_state.json` are reset to 0...)
3. run `gaiad testnet unsafe-start-local-validator` -> switches the validator set and starts the node

