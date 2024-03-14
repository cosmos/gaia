# Gaia v15.1.0  Release Notes 

***This release contain the changes introduced by the v14.2.0 emergency release and should be used for the v15 upgrade (instead of ~~v15.0.0~~).***

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v15.1.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v14.2.0...v15.1.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v15.x/UPGRADING.md) when migrating from `v14.2.x` to `v15.1.x`.

## 🚀 Highlights

<!-- Add any highlights of this release --> 

As this release replaces the [v15.0.0](https://github.com/cosmos/gaia/releases/tag/v15.0.0) release, please check out the release notes for all the highlights. 

In addition, this release bumps Packet Forward Middleware to `v7.1.3-0.20240228213828-cce7f56d000b`, which contains the same fix as the one introduced by the [v14.2.0](https://github.com/cosmos/gaia/releases/tag/v14.2.0) emergency release. It also fixes a series of escrow accounts by minting and transfering the missing assets to reach parity with counterparty chain supply. 

## 🔨 Build from source

You must use Golang `v1.21` if building from source.

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v15.1.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.