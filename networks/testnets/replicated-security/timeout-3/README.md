
# `timeout-3` Chain Details

> Status: **STARTED | waiting for `ccv_timeout_period`**

The timeout-3 chain will be launched with the purpose of testing the provider chain `ccv_timeout_period` timeout. The relayer used to first establish the connection between `timeout-3` and `provider-1` will be stopped shortly afterwards. The relayer will then be started again after the `ccv_timeout_period` has elapsed but before the `vsc_timeout_period` is reached.

- **Chain-ID**: `timeout-3`
- **Launch date**: 2023-01-23
* **GitHub Repo**: cosmos/interchain-security
* **Release**: [v1.0.0-rc3](https://github.com/cosmos/interchain-security/releases/tag/v1.0.0-rc3)
* **Genesis File:**  [timeout-3-genesis.json](timeout-3-genesis.json), verify with `shasum -a 256 timeout-3-genesis.json`
* **Genesis sha256sum**: `60779cd350ee2cc184b58bbfe84831b25ae1a10e75d438bb096b9b3cba1dce6e`

Seed node:

1. `08ec17e86dac67b9da70deb20177655495a55407@timeout-3-seed.rs-testnet.polypore.xyz:26656`

## How to Join

The scripts provided in this repo will install `interchain-security-cd` and setup a system service on your machine.

### Bash Script

Run the script provided in this repo to join the `timeout-3` chain:
* `join-rs-timeout-3.sh` will create a `timeout-3` service.
* It can be run either as root or from a sudoer account.
* It will attempt to build the binary from the interchain-security [repo](https://github.com/cosmos/interchain-security).