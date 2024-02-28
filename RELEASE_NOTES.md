# Gaia v15.0.0  Release Notes 

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/gaia/blob/v15.0.0/CHANGELOG.md) for a list of relevant changes or [compare all changes](https://github.com/cosmos/gaia/compare/v14.1.0...v15.0.0) from last release.

<!-- Add the following line for major releases -->
Refer to the [upgrading guide](https://github.com/cosmos/gaia/blob/release/v15.x/UPGRADING.md) when migrating from `v14.x` to `v15.x`.

## 🚀 Highlights

<!-- Add any highlights of this release --> 

The focus of this release is the upgrade of Cosmos SDK to v0.47 -- this release uses [v0.47.10-ics-lsm](https://github.com/cosmos/cosmos-sdk/tree/v0.47.10-ics-lsm), a special Cosmos SDK branch with support for both ICS and LSM. Consequently, it also upgrades the following dependencies:

- IBC to [v7.3.1](https://github.com/cosmos/ibc-go/releases/tag/v7.3.1)
- CometBFT to [v0.37.4](https://github.com/cometbft/cometbft/releases/tag/v0.37.4)
- Interchain Security to [v3.3.3-lsm](https://github.com/cosmos/interchain-security/releases/tag/v3.3.3-lsm)
- Packet Forward Middleware to [v7.1.2](https://github.com/cosmos/ibc-apps/releases/tag/middleware%2Fpacket-forward-middleware%2Fv7.1.2)

In addition, this release migrates the following state:

- Sets the min commission rate staking parameter to `5%` and updates the commission rate for all validators accordingly (according to [governance proposal 826](https://www.mintscan.io/cosmos/proposals/826)). 
- Migrates the vesting funds from _cosmos145hytrc49m0hn6fphp8d5h4xspwkawcuzmx498_
 to the community pool (according to [governance proposal 860](https://www.mintscan.io/cosmos/proposals/860)).

Also, this release adds support for metaprotocols using the transaction extension options. See the [docs of the x/metaprotocols module](https://github.com/cosmos/gaia/tree/release/v15.x/x/metaprotocols) for more details.  

Finally, this releases fixes a series of issues found during the [Oak Security audit of SDK 0.47](https://github.com/oak-security/audit-reports/blob/master/Cosmos%20SDK/2024-01-23%20Audit%20Report%20-%20Cosmos%20SDK%20v1.0.pdf). As a result, this release introduces the following API changes:

 - Reject `MsgVote` messages from accounts with less than 1 ATOM staked.
 - A new `MinDepositRatio` param is added (with a default value of `0.01`) and now deposits are required to be at least `MinDepositRatio*MinDeposit` to be accepted.


## 🔨 Build from source

You must use Golang `v1.21` if building from source.

```bash
git clone https://github.com/cosmos/gaia
cd gaia && git checkout v15.0.0
make install
```

## ⚡️ Download binaries

Binaries for linux, darwin, and windows are available below.