---
title: Cosmos Hub 4, Rho Upgrade
order: 2
---
<!-- markdown-link-check-disable -->
# Cosmos Hub 4, v8-Rho Upgrade, Instructions

This document describes the steps for validator and full node operators for the successful execution of the [v8-Rho Upgrade](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md#v8-rho-upgrade-expected-q1-2023), which contains the following main new features/improvement:

- [ibc-go](https://github.com/cosmos/ibc-go) to [v3.4.0](https://github.com/cosmos/ibc-go/blob/v3.4.0/CHANGELOG.md) to fix a vulnerability in ICA. See [v3.4.0 CHANGELOG.md](https://github.com/cosmos/ibc-go/releases/tag/v3.4.0) and [v3.2.1 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v3.2.1) for details.
- [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.12). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.
- [tendermint](https://github.com/tendermint/tendermint) to [0.34.24](https://github.com/tendermint/tendermint/tree/v0.34.24). See [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/v0.34.24/CHANGELOG.md) for details.
- [liquidity](https://github.com/Gravity-Devs/liquidity) to [v1.5.3](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.3).
- [packet-forwarding-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v3.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v3.1.1).
- [globalfee](https://github.com/cosmos/gaia/tree/main/x/globalfee) module. See [globalfee docs](https://github.com/cosmos/gaia/blob/main/docs/modules/globalfee.md) for more details.
- [#1845](https://github.com/cosmos/gaia/pull/1845) Add bech32-convert command to gaiad.
- [Add new fee decorator](https://github.com/cosmos/gaia/pull/1961) to change `MaxBypassMinFeeMsgGasUsage` so importers of x/globalfee can change `MaxGas`.
- [#1870](https://github.com/cosmos/gaia/issues/1870) Fix bank denom metadata in migration. See [#1892](https://github.com/cosmos/gaia/pull/1892) for more details.
- [#1976](https://github.com/cosmos/gaia/pull/1976) Fix Quicksilver ICA exploit in migration. See [the bug fix forum post](https://forum.cosmos.network/t/upcoming-interchain-accounts-bugfix-release/8911) for more details.
- [E2E tests](https://github.com/cosmos/gaia/tree/main/tests/e2e). The tests cover transactions/queries tests of different modules, including Bank, Distribution, Encode, Evidence, FeeGrant, Global Fee, Gov, IBC, packet forwarding middleware, Slashing, Staking, and Vesting module.
- [#1941](https://github.com/cosmos/gaia/pull/1941) Fix packet forward configuration for e2e tests.
- Use gaiad to swap out [Ignite](https://github.com/ignite/cli) in [liveness tests](https://github.com/cosmos/gaia/blob/main/.github/workflows/test.yml).


TOC:

- [Cosmos Hub 4, v8-Rho Upgrade, Instructions](#cosmos-hub-4-v8-rho-upgrade-instructions)
  - [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
  - [Upgrade will take place Feb 16, 203](#upgrade-will-take-place-feb-16-2023)
  - [Chain-id will remain the same](#chain-id-will-remain-the-same)
  - [Preparing for the upgrade](#preparing-for-the-upgrade)
    - [System requirement](#system-requirement)
    - [Backups](#backups)
    - [Testing](#testing)
    - [Current runtime, cosmoshub-4 (pre-v8-Rho upgrade) is running Gaia v7.0.x](#current-runtime-cosmoshub-4-pre-v7-theta-upgrade-is-running-gaia-v60x)
    - [Target runtime, cosmoshub-4 (post-v8-Rho upgrade) will run Gaia v8.0.0](#target-runtime-cosmoshub-4-post-v8-rho-upgrade-will-run-gaia-v800)
  - [v8-Rho upgrade steps](#v8-Rho-upgrade-steps)
    - [Method I: Manual Upgrade](#method-i-manual-upgrade)
    - [Method II: Upgrade using Cosmovisor](#method-ii-upgrade-using-cosmovisor)
      - [Manually preparing the binary](#manually-preparing-the-gaia-v800-binary)
        - [Preparation](#preparation)
        - [Expected upgrade result](#expected-upgrade-result)
      - [Auto-Downloading the Gaia v8.0.0 binary (not recommended!)](#auto-downloading-the-gaia-v800-binary-not-recommended)
        - [Preparation](#preparation-1)
        - [Expected result](#expected-result)
  - [Upgrade duration](#upgrade-duration)
  - [Rollback plan](#rollback-plan)
  - [Communications](#communications)
  - [Risks](#risks)
  - [Reference](#reference)

## On-chain governance proposal attains consensus

[Proposal #97](https://www.mintscan.io/cosmos/proposals/97) is the reference on-chain governance proposal for this upgrade, which has passed with overwhelming community support. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.

## Upgrade will take place Feb 16, 2023

The upgrade will take place at a block height of `14099412`. At the time of writing, and at current block times (around 7s/block), this block height corresponds approximately to `Thursday, 16-February-23 01:00:00 CET`. This date/time is approximate as blocks are not generated at a constant interval. You can stay up-to-date using this [live countdown](https://chain-monitor.cros-nest.com/d/Upgrades/upgrades?var-chain_id=cosmoshub-4&orgId=1&refresh=1m) page.

## Chain-id will remain the same

The chain-id of the network will remain the same, `cosmoshub-4`. This is because an in-place migration of state will take place, i.e., this upgrade does not export any state.

## Preparing for the upgrade

### System requirement

32GB RAM is recommended to ensure a smooth upgrade.

If you have less than 32GB RAM, you might try creating a swapfile to swap an idle program onto the hard disk to free up memory. This can
allow your machine to run the binary than it could run in RAM alone.

```shell
sudo fallocate -l 16G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### Backups

Prior to the upgrade, validators are encouraged to take a full data snapshot. Snapshotting depends heavily on infrastructure, but generally this can be done by backing up the `.gaia` directory.
If you use Cosmovisor to upgrade, by default, Cosmovisor will backup your data upon upgrade. See below [upgrade by cosmovisor](#method-ii-upgrade-using-cosmovisor-by-manually-preparing-the-gaia-v700-binary) section.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can run a [v8-Rho local testnet](https://github.com/cosmos/testnets/tree/master/local).

### Current runtime, cosmoshub-4 (pre-v8-Rho upgrade) is running Gaia v7.1.1

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v7.1.1](https://github.com/cosmos/gaia/releases/v7.1.1). We anticipate that operators who are running on v7.1.1, will be able to upgrade successfully; however, this is untested and it is up to operators to ensure that their systems are capable of performing the upgrade.

### Target runtime, cosmoshub-4 (post-v8-Rho upgrade) will run Gaia v8.0.0

The Cosmos Hub mainnet network, `cosmoshub-4`, will run [Gaia v8.0.0](https://github.com/cosmos/gaia/releases/tag/v8.0.0). Operators _MUST_ use this version post-upgrade to remain connected to the network.

## v8-Rho upgrade steps

There are 2 major ways to upgrade a node:

- Manual upgrade
- Upgrade using [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)
  - Either by manually preparing the new binary
  - Or by using the auto-download functionality (this is not yet recommended)

If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: Manual Upgrade

Run Gaia v7.1.1 till upgrade height, the node will panic:

```shell
ERR UPGRADE "v8-Rho" NEEDED at height: 14099412: upgrade to v7-Theta and applying upgrade "v8-Rho" at height:14099412
```

Stop the node, and install Gaia v8.0.0 and re-start by `gaiad start`.

It may take several minutes to a few hours until validators with a total sum voting power > 2/3 to complete their nodes upgrades. After that, the chain can continue to produce blocks.

### Method II: Upgrade using Cosmovisor

> **Warning**  <span style="color:red">**Please Read Before Proceeding**</span><br>
> **Using Cosmovisor 1.2.0 and higher requires a lowercase naming convention for upgrade version directory. For Cosmovisor 1.1.0 and earlier, the upgrade version is not lowercased.**       
> 
> **For Example:** <br>
> **Cosmovisor =< `1.1.0`: `/upgrades/v8-Rho/bin/gaiad`**<br>
> **Cosmovisor >= `1.2.0`: `/upgrades/v8-rho/bin/gaiad`**<br>

| Cosmovisor Version | Binary Name in Path |
|--------------------|---------------------|
| 1.3                | v8-rho              |
| 1.2                | v8-rho              |
| 1.1                | v8-Rho              |
| 1.0                | v8-Rho              |



### _Manually preparing the Gaia v8.0.0 binary_

##### Preparation

Install the latest version of Cosmovisor (`1.3.0`):

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

**Verify Cosmovisor Version**
```shell
cosmovisor version
cosmovisor version: v1.3.0
```

Create a cosmovisor folder:

create a Cosmovisor folder inside `$GAIA_HOME` and move Gaia v7.1.1 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
````

build Gaia v8.0.0, and move gaiad v8.0.0 to `$GAIA_HOME/cosmovisor/upgrades/v8-rho/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/v8-rho/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/v8-rho/bin
```

Then you should get the following structure:

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  #v7.1.1
└── upgrades
    └── v8-rho
        └── bin
            └── gaiad  #v8.0.0
```

Export the environmental variables:

```shell
export DAEMON_NAME=gaiad
# please change to your own gaia home dir
export DAEMON_HOME= $GAIA_HOME
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Start the node:

```shell
cosmovisor start --x-crisis-skip-assert-invariants
```

Skipping the invariant checks is strongly encouraged since it decreases the upgrade time significantly and since there are some other improvements coming to the crisis module in the next release of the Cosmos SDK.

#### Expected upgrade result

When the upgrade block height is reached, Gaia will panic and stop:

This may take 7 minutes to a few hours.
After upgrade, the chain will continue to produce blocks when validators with a total sum voting power > 2/3 complete their nodes upgrades.

### _Auto-Downloading the Gaia v8.0.0 binary (not recommended!)_
#### Preparation

Install the latest version of Cosmovisor (`1.3.0`):

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

Create a cosmovisor folder:

create a cosmovisor folder inside gaia home and move gaiad v7.1.1 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
```

```shell
.
├── current -> genesis or upgrades/<name>
└── genesis
     └── bin
        └── gaiad  #v7.1.1
```

Export the environmental variables:

```shell
export DAEMON_NAME=gaiad
# please change to your own gaia home dir
export DAEMON_HOME= $GAIA_HOME
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
```

Start the node:

```shell
cosmovisor start --x-crisis-skip-assert-invariants
```

Skipping the invariant checks is strongly encouraged since it decreases the upgrade time significantly and since there are some other improvements coming to the crisis module in the next release of the Cosmos SDK.

#### Expected result

When the upgrade block height is reached, you can find the following information in the log:

```shell
ERR UPGRADE "v8-Rho" NEEDED at height: 14099412: upgrade to v7-Theta and applying upgrade "v8-Rho" at height:14099412
```

Then the Cosmovisor will create `$GAIA_HOME/cosmovisor/upgrades/v8-rho/bin` and download the Gaia v8.0.0 binary to this folder according to links in the `--info` field of the upgrade proposal 97.
This may take 7 minutes to a few hours, afterwards, the chain will continue to produce blocks once validators with a total sum voting power > 2/3 complete their nodes upgrades.

_Please Note:_

- In general, auto-download comes with the risk that the verification of correct download is done automatically. If users want to have the highest guarantee users should confirm the check-sum manually. We hope more node operators will use the auto-download for this release but please be aware this is a risk and users should take at your own discretion.
- Users should use run node on v7.1.1 if they use the cosmovisor v1.1.0 with auto-download enabled for upgrade process.

## Upgrade duration

The upgrade may take a few minutes to several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v7.1.1 binary using the following command:

> gaiad start --unsafe-skip-upgrade 14099412

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#validators-verified` channel of the Cosmos Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand. Requests to join the `#validators-verified` channel can be sent to the `#general-support` channel.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it.

## Reference

[join Cosmos Hub Mainnet](https://github.com/cosmos/mainnet)

<!-- markdown-link-check-enable -->
