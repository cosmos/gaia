
# `provider-1` Chain Details

The provider chain functions as an analogue of the Cosmos Hub. Its governance parameters will provide short voting periods to accelerate the creation of consumer chains.

* **Chain-ID**: `provider-1`
* **denom**: `uatom`
* **Launch date**: 2023-01-23
* **Launch Gaia Version:** [`v9.0.0-rc1`](https://github.com/cosmos/gaia/releases/tag/v9.0.0-rc1)
* **Genesis File:**  [provider-1-genesis.json](provider-1-genesis.json), verify with `shasum -a 256 provider-1-genesis.json`
* **Genesis sha256sum**: `4cdd90af813df8655c9bdc9bd33cd18bf93dd1ab50ba3a76794947bef86dd487`

## Endpoints

Endpoints are exposed as subdomains for the sentry and snapshot nodes (described below) as follows:

* `https://rpc.<node-name>.rs-testnet.polypore.xyz:443`
* `https://rest.<node-name>.rs-testnet.polypore.xyz:443`
* `https://grpc.<node-name>.rs-testnet.polypore.xyz:443`
* `p2p.<node-name>.rs-testnet.polypore.xyz:26656`

Sentries:

1. `provider-sentry-01.rs-testnet.polypore.xyz`
2. `provider-sentry-02.rs-testnet.polypore.xyz`

Seed nodes:

1. `08ec17e86dac67b9da70deb20177655495a55407@provider-seed-01.rs-testnet.polypore.xyz:26656`
2. `4ea6e56300a2f37b90e58de5ee27d1c9065cf871@provider-seed-02.rs-testnet.polypore.xyz:26656`

The following state sync nodes serve snapshots every 1000 blocks:

1. `provider-state-sync-01.rs-testnet.polypore.xyz`
2. `provider-state-sync-02.rs-testnet.polypore.xyz`

## Block Explorer

* https://explorer.rs-testnet.polypore.xyz

## Consumer Chains IBC Data

Connections and channels will be posted here shortly after a consumer chain launches.

### Clients

* timeout-1: 07-tendermint-0
* timeout-2: 07-tendermint-1
* timeout-3: 07-tendermint-2

## Faucet

* Visit `faucet.rs-testnet.polypore.xyz` to request tokens and check your address balance.

## How to Join

The scripts provided in this repo will install Gaia and optionally set up a Cosmovisor service with the auto-download feature enabled on your machine.

You can choose to (not) use state sync. Your node will sync much faster if you use state sync, but it will not keep all the state locally.

### Bash Script

Run either one of the scripts provided in this repo to join the provider chain:
* `join-rs-provider.sh` will create a `gaiad` service.
* `join-rs-provider-cv.sh` will create a `cosmovisor` service.
* Both scripts must be run either as root or from a sudoer account.
* Both scripts will attempt to download the amd64 binary from the Gaia [releases](https://github.com/cosmos/gaia/releases) page. You can modify the `CHAIN_BINARY_URL` to match your target architecture if needed.

#### State Sync Option

* By default, the scripts will attempt to use state sync to catch up quickly to the current height. To turn off state sync, set `STATE_SYNC` to `false`.

## Creating a Validator

> If there are any consumer chains on the testnet, you must be running a node in those as well, or your validator will get jailed due to downtime depending on each consumer chain's `slashing` paramentes.

Once you have some tokens in your self-delegation account, you can submit the `create-validator` transaction.

1. Obtain the validator public key
```
gaiad tendermint show-validator
```

2. Submit the `create-validator` transaction.
```bash
gaiad tx staking create-validator \
--amount 1000000uprov \
--pubkey '<public key from the previous command>' \
--moniker <your moniker> \
--chain-id provider \
--commission-rate 0.10 \
--commission-max-rate 1.00 \
--commission-max-change-rate 0.1 \
--min-self-delegation 1000000 \
--gas auto \
--from <self-delegation-account>
```

You can verify the validator was created in the block explorer, or in the command line:
```
gaiad q staking validators -o json | jq '.validators[].description.moniker'
```