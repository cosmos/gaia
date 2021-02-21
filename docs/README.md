<!--
parent:
  order: false
layout: home
-->

# Cosmos Hub Documentation

Welcome to the documentation of the **Cosmos Hub application: `gaia`**.

## Join Mainnet

- [Detailed node setup information](./docs/gaia-tutorials/join-mainnet.md)

**Instant Gratification Snippet**

```bash
git clone -b v4.0.4 https://github.com/cosmos/gaia
make install
gaiad init chooseanicehandle
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json
gaiad start --p2p.seeds bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656
```
If you'd like to save those seeds to your settings put them in ~/.gaia/config/config.toml in the p2p section under seeds in the same comma-separated list format.




## What is Gaia?

- [Intro to the `gaia` software](./gaia-tutorials/what-is-gaia.md)

## Join the Cosmos Hub Mainnet

- [Install the `gaia` application](./gaia-tutorials/installation.md)
- [Set up a full node and join the mainnet](./gaia-tutorials/join-mainnet.md)
- [Upgrade to a validator node](./validators/validator-setup.md)

## Join the Cosmos Hub Public Testnet

- [Join the testnet](./gaia-tutorials/join-testnet.md)

## Setup Your Own `gaia` Testnet

- [Setup your own `gaia` testnet](./gaia-tutorials/deploy-testnet.md)

## Additional Resources

- [Validator Resources](./validators/README.md): Contains documentation for `gaia` validators.
- [Delegator Resources](./delegators/README.md): Contains documentation for delegators.
- [Other Resources](./resources/README.md): Contains documentation on `gaiad`, genesis file, service providers, ledger wallets, ...
- [Cosmos Hub Archives](./resources/archives.md): State archives of past iteration of the Cosmos Hub.

# Contribute

See [this file](./DOCS_README.md) for details of the build process and
considerations when making changes.
