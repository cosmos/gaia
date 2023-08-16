---
title: Cosmos Hub 4, Gaia v11 Upgrade
order: 1
---
<!-- markdown-link-check-disable -->
# Cosmos Hub 4, Gaia v11 Upgrade, Instructions

This document describes the steps for validators and full node operators, to upgrade successfully to the Gaia v11 release. 
For more details on the release, please see the [release notes](https://github.com/cosmos/gaia/releases/tag/v11.0.0)

## Instructions
  
- [Cosmos Hub 4, Gaia v11 Upgrade, Instructions](#cosmos-hub-4-gaia-v11-upgrade-instructions)
  - [Instructions](#instructions)
  - [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
  - [Upgrade date](#upgrade-date)
  - [Chain-id will remain the same](#chain-id-will-remain-the-same)
  - [Preparing for the upgrade](#preparing-for-the-upgrade)
    - [System requirement](#system-requirement)
    - [Backups](#backups)
    - [Testing](#testing)
    - [Current runtime](#current-runtime)
    - [Target runtime](#target-runtime)
  - [Upgrade steps](#upgrade-steps)
    - [Method I: Manual Upgrade](#method-i-manual-upgrade)
    - [Method II: Upgrade using Cosmovisor](#method-ii-upgrade-using-cosmovisor)
    - [Manually preparing the binary](#manually-preparing-the-binary)
        - [Preparation](#preparation)
      - [Expected upgrade result](#expected-upgrade-result)
    - [Auto-Downloading the Gaia binary](#auto-downloading-the-gaia-binary)
      - [Preparation](#preparation-1)
      - [Expected result](#expected-result)
  - [Upgrade duration](#upgrade-duration)
  - [Rollback plan](#rollback-plan)
  - [Communications](#communications)
  - [Risks](#risks)
  - [Reference](#reference)

## On-chain governance proposal attains consensus

[Proposal 804](https://www.mintscan.io/cosmos/proposals/804) is the reference on-chain governance proposal for this upgrade, which is still in its voting period. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.

## Upgrade date

The upgrade will take place at a block height of `16596000`. The date/time of the upgrade is subject to change as blocks are not generated at a constant interval. You can stay up-to-date using this [live countdown](https://www.mintscan.io/cosmos/blocks/16596000) page.

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
If you use Cosmovisor to upgrade, by default, Cosmovisor will backup your data upon upgrade. See below [upgrade using cosmovisor](#method-ii-upgrade-using-cosmovisor) section.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can run a [v11 Local Testnet](https://github.com/cosmos/testnets/tree/master/local) or join in our [Cosmos Hub Public Testnet](https://github.com/cosmos/testnets/tree/master/public).

### Current runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v10.0.2](https://github.com/cosmos/gaia/releases/v10.0.2). We anticipate that operators who are running on v10.0.x, will be able to upgrade successfully. Validators are expected to ensure that their systems are up to date and capable of performing the upgrade. This includes running the correct binary, or if building from source, building with go `1.20`.

### Target runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, will run [Gaia v11.0.0](https://github.com/cosmos/gaia/releases/tag/v11.0.0). Operators _**MUST**_ use this version post-upgrade to remain connected to the network.

## Upgrade steps

There are 2 major ways to upgrade a node:

- Manual upgrade
- Upgrade using [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)
    - Either by manually preparing the new binary
    - Or by using the auto-download functionality (this is not yet recommended)

If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: Manual Upgrade

Make sure Gaia v11.0.0 is installed by either downloading a [compatible binary](https://github.com/cosmos/gaia/releases/tag/v11.0.0), or building from source. Building from source requires **Golang 1.20**.

Run Gaia v10.0.x till upgrade height, the node will panic:

```shell
ERR UPGRADE "v11" NEEDED at height: 16596000: upgrade to v11 and applying upgrade "v11" at height:16596000
```

Stop the node, and switch the binary to Gaia v11.0.0 and re-start by `gaiad start`.

It may take several minutes to a few hours until validators with a total sum voting power > 2/3 to complete their node upgrades. After that, the chain can continue to produce blocks.

### Method II: Upgrade using Cosmovisor

### Manually preparing the binary

##### Preparation

Install the latest version of Cosmovisor (`1.5.0`):

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

**Verify Cosmovisor Version**
```shell
cosmovisor version
cosmovisor version: v1.5.0
```

Create a cosmovisor folder:

create a Cosmovisor folder inside `$GAIA_HOME` and move Gaia v9.1.1 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
````

build Gaia v11.0.0, and move gaiad v11.0.0 to `$GAIA_HOME/cosmovisor/upgrades/v11/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/v11/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/v11/bin
```

Then you should get the following structure:

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  #v10.0.x
└── upgrades
    └── v11
        └── bin
            └── gaiad  #v11.0.0
```

Export the environmental variables:

```shell
export DAEMON_NAME=gaiad
# please change to your own gaia home dir
# please note `DAEMON_HOME` has to be absolute path
export DAEMON_HOME=$GAIA_HOME
export DAEMON_RESTART_AFTER_UPGRADE=true
```

Start the node:

```shell
cosmovisor run  start --x-crisis-skip-assert-invariants --home $DAEMON_HOME
```

Skipping the invariant checks is strongly encouraged since it decreases the upgrade time significantly and since there are some other improvements coming to the crisis module in the next release of the Cosmos SDK.

#### Expected upgrade result

When the upgrade block height is reached, Gaia will panic and stop:

This may take 7 minutes to a few hours.
After upgrade, the chain will continue to produce blocks when validators with a total sum voting power > 2/3 complete their node upgrades.

### Auto-Downloading the Gaia binary

**This method is not recommended!**

#### Preparation

Install the latest version of Cosmovisor (`1.5.0`):

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

Create a cosmovisor folder:

create a cosmovisor folder inside gaia home and move gaiad v10.0.x into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
```

```shell
.
├── current -> genesis or upgrades/<name>
└── genesis
     └── bin
        └── gaiad  #v10.0.x
```

Export the environmental variables:

```shell
export DAEMON_NAME=gaiad
# please change to your own gaia home dir
export DAEMON_HOME=$GAIA_HOME
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
```

Start the node:

```shell
cosmovisor run start --x-crisis-skip-assert-invariants --home $DAEMON_HOME
```

Skipping the invariant checks can help decrease the upgrade time significantly.

#### Expected result

When the upgrade block height is reached, you can find the following information in the log:

```shell
ERR UPGRADE "v11" NEEDED at height: 16596000: upgrade to v11 and applying upgrade "v11" at height:16596000
```

Then the Cosmovisor will create `$GAIA_HOME/cosmovisor/upgrades/v11/bin` and download the Gaia v11.0.0 binary to this folder according to links in the `--info` field of the upgrade proposal.
This may take 7 minutes to a few hours, afterwards, the chain will continue to produce blocks once validators with a total sum voting power > 2/3 complete their nodes upgrades.

_Please Note:_

- In general, auto-download comes with the risk that the verification of correct download is done automatically. If users want to have the highest guarantee users should confirm the check-sum manually. We hope more node operators will use the auto-download for this release but please be aware this is a risk and users should take at your own discretion.
- Users should run their node on v10.0.x if they use the cosmovisor v1.5.0 with auto-download enabled for upgrade process.

## Upgrade duration

The upgrade may take a few minutes to several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v10.0.2 binary using the following command:

> gaiad start --unsafe-skip-upgrade 16596000

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#cosmos-hub-validators-verified` channel of the Cosmos Hub Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand. Requests to join the `#cosmos-hub-validators-verified` channel can be sent to the `#general-support` channel.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it.

## Reference

[Join Cosmos Hub Mainnet](https://github.com/cosmos/mainnet)

<!-- markdown-link-check-enable -->
