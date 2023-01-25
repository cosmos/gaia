# Rho Developer Testnet

Integrators such as exchanges and wallets who want to test against Rho endpoints early may do so using the Rho devnet.

- The devnet consists of two chains with a relayer connecting them over IBC.
- Each chain has two nodes: a validator and a sentry.
- The `cosmos-ansible` repo provides inventory files to [join this devnet](https://github.com/hyphacoop/cosmos-ansible/tree/main/examples#join-the-rho-devnet) using Ansible plays.

**NOTE:** This devnet is restarted regularly with a fresh state. If you would like us to add you to our genesis accounts each time, please make a PR with a cosmos address and stating that you would like to be automatically included in future devnets.

## Chain 1

* **Chain ID**: `rho-chain-1`
* **Launch date:** 2023-01-17
* **Gaia version:** `v8.0.0-rc3`
* **Genesis file:** [rho-chain-1-genesis.json.gz](rho-chain-1-genesis.json.gz)
* **Persistent peer:** `42c0aa9a5e922ae446fa8537f076854a6c2c5053@sentry.chain-1.rho-devnet.polypore.xyz:26656`

### Endpoints

* `http://rpc.sentry.chain-1.rho-devnet.polypore.xyz:26657`
* `http://rest.sentry.chain-1.rho-devnet.polypore.xyz:1317`
* `http://grpc.sentry.chain-1.rho-devnet.polypore.xyz:9090`
* `tcp://p2p.sentry.chain-1.rho-devnet.polypore.xyz:26656`

### REST Faucet API

* Request endpoint
  ```
  http://sentry.chain-1.rho-devnet.polypore.xyz:8000/request?chain=rho-chain-1&address=<cosmos address>
  ```
* Balance endpoint
  ```
  http://sentry.chain-1.rho-devnet.polypore.xyz:8000/balance?chain=rho-chain-1&address=<cosmos address>
  ```

## Chain 2

- **Chain ID**: `rho-chain-2`
- **Launch date:** 2023-01-17
- **Gaia version:** `v8.0.0-rc3`
- **Genesis file:** [rho-chain-2-genesis.json.gz](rho-chain-2-genesis.json.gz)
* **Persistent peer:** `c45efb232d01df3b2399ef9b4b801fd41f6a89fc@sentry.chain-2.rho-devnet.polypore.xyz:26656`

### Endpoints

* `http://rpc.sentry.chain-2.rho-devnet.polypore.xyz:26657`
* `http://rest.sentry.chain-2.rho-devnet.polypore.xyz:1317`
* `http://grpc.sentry.chain-2.rho-devnet.polypore.xyz:9090`
* `tcp://p2p.sentry.chain-2.rho-devnet.polypore.xyz:26656`

### REST Faucet API

* Request endpoint
  ```
  http://sentry.chain-2.rho-devnet.polypore.xyz:8000/request?chain=rho-chain-2&address=<cosmos address>
  ```
* Balance endpoint
  ```
  http://sentry.chain-2.rho-devnet.polypore.xyz:8000/balance?chain=rho-chain-2&address=<cosmos address>
  ```

