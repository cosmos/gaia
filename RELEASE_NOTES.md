# Gaia v21.0.0 Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v21.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v20.0.0....v21.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v21.x/UPGRADING.md) when migrating from `v20.x` to `v21.x`.

## 🚀 Highlights

<!-- Add any highlights of this release -->

This release bumps Interchain Security (ICS) to [v6.3.0](https://github.com/cosmos/interchain-security/releases/tag/v6.3.0) which brings the following improvements:

- Enable consumer chains to use the memo field of the IBC transfer packets to tag ICS rewards with the consumer ID. As a result, consumer chains can send ICS rewards in any denomination and on any IBC channel. For example, a consumer chain could send USDC as ICS rewards. 

- Enable each consumer chain to permissionlessly add up to three denoms that will be accepted as ICS rewards.

The release also distributes all the unaccounted known denoms from  the consumer rewards pool via migration.

## 🔨 Build from source

❗***You must use Golang v1.22 if building from source.***

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v21.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.