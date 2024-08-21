# Upgrading Gaia

This guide provides instructions for upgrading Gaia from v18.1.x. to v19.1.x.

This document describes the steps for validators, full node operators and relayer operators, to upgrade successfully for the Gaia v18 release.

For more details on the release, please see the [release notes](https://github.com/cosmos/gaia/releases/tag/v19.1.0)

**Relayer Operators** for the Cosmos Hub and consumer chains, will also need to update to use [Hermes v1.10.0](https://github.com/informalsystems/hermes/releases/tag/v1.10.0) or higher. You may need to restart your relayer software after a major chain upgrade.

## Release Binary

Please use the correct release binary: `v19.1.0`.

## Instructions

- [Upgrading Gaia](#upgrading-gaia)
  - [Release Binary](#release-binary)
  - [Instructions](#instructions)
  - [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
  - [Upgrade date](#upgrade-date)
  - [Preparing for the upgrade](#preparing-for-the-upgrade)
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
  - [Upgrade duration](#upgrade-duration)
  - [Rollback plan](#rollback-plan)
  - [Communications](#communications)
  - [Risks](#risks)
  - [Reference](#reference)

## On-chain governance proposal attains consensus

Once a software upgrade governance proposal is submitted to the Cosmos Hub, both a reference to this proposal and an `UPGRADE_HEIGHT` are added to the [release notes](https://github.com/cosmos/gaia/releases/tag/v19.1.0).
If and when this proposal reaches consensus, the upgrade height will be used to halt the "old" chain binaries. You can check the proposal on one of the block explorers or using the `gaiad` CLI tool.

## Upgrade date

The date/time of the upgrade is subject to change as blocks are not generated at a constant interval. You can stay up-to-date by checking the estimated estimated time until the block is produced one of the block explorers (e.g. https://www.mintscan.io/cosmos/blocks/`UPGRADE_HEIGHT`).

## Preparing for the upgrade

### Backups

Prior to the upgrade, validators are encouraged to take a full data snapshot. Snapshotting depends heavily on infrastructure, but generally this can be done by backing up the `.gaia` directory.
If you use Cosmovisor to upgrade, by default, Cosmovisor will backup your data upon upgrade. See below [upgrade using cosmovisor](#method-ii-upgrade-using-cosmovisor) section.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can run a [v19 Local Testnet](https://github.com/cosmos/testnets/tree/master/local) or join in our [Cosmos Hub Public Testnet](https://github.com/cosmos/testnets/tree/master/public).

### Current runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v18.1.0](https://github.com/cosmos/gaia/releases/v18.1.0). We anticipate that operators who are running on v18.1.0, will be able to upgrade successfully. Validators are expected to ensure that their systems are up to date and capable of performing the upgrade. This includes running the correct binary and if building from source, building with the appropriate `go` version.

### Target runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, will run **[Gaia v19.1.0](https://github.com/cosmos/gaia/releases/tag/v19.1.0)**. Operators _**MUST**_ use this version post-upgrade to remain connected to the network. The new version requires `go v1.22` to build successfully.

## Upgrade steps

There are 2 ways to upgrade a node:

- Manual upgrade
- Upgrade using [Cosmovisor](https://pkg.go.dev/cosmossdk.io/tools/cosmovisor)
    - Either by manually preparing the new binary
    - Or by using the auto-download functionality (this is not yet recommended)

If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: Manual Upgrade

Make sure **Gaia v18.1.0** is installed by either downloading a [compatible binary](https://github.com/cosmos/gaia/releases/tag/v18.1.0), or building from source. Check the required version to build this binary in the `Makefile`.

Run Gaia v18.1.0 till upgrade height, the node will panic:

```shell
ERR UPGRADE "v19" NEEDED at height: <UPGRADE_HEIGHT>: upgrade to v19 and applying upgrade "v19" at height:<UPGRADE_HEIGHT>
```

Stop the node, and switch the binary to **Gaia v19.1.0** and re-start by `gaiad start`.

It may take several minutes to a few hours until validators with a total sum voting power > 2/3 to complete their node upgrades. After that, the chain can continue to produce blocks.

### Method II: Upgrade using Cosmovisor

#### Manually preparing the binary

##### Preparation

- Install the latest version of Cosmovisor (`1.5.0`):

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
cosmovisor version
# cosmovisor version: v1.5.0
```

- Create a `cosmovisor` folder inside `$GAIA_HOME` and move Gaia `v18.1.0` into `$GAIA_HOME/cosmovisor/genesis/bin`:

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
```

- Build Gaia `v19.1.0`, and move gaiad `v19.1.0` to `$GAIA_HOME/cosmovisor/upgrades/v19/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/v19/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/v19/bin
```

At this moment, you should have the following structure:

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  # old: v18.1.0
└── upgrades
    └── v19
        └── bin
            └── gaiad  # new: v19.1.0
```

- Export the environmental variables:

```shell
export DAEMON_NAME=gaiad
# please change to your own gaia home dir
# please note `DAEMON_HOME` has to be absolute path
export DAEMON_HOME=$GAIA_HOME
export DAEMON_RESTART_AFTER_UPGRADE=true
```

- Start the node:

```shell
cosmovisor run start --x-crisis-skip-assert-invariants --home $DAEMON_HOME
```

Skipping the invariant checks is strongly encouraged since it decreases the upgrade time significantly and since there are some other improvements coming to the crisis module in the next release of the Cosmos SDK.

##### Expected upgrade result

When the upgrade block height is reached, Gaia will panic and stop:

This may take a few minutes.
After upgrade, the chain will continue to produce blocks when validators with a total sum voting power > 2/3 complete their node upgrades.

#### Auto-Downloading the Gaia binary

## Upgrade duration

The upgrade may take a few minutes to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v18.1.0 binary using the following command:

```shell
gaiad start --unsafe-skip-upgrade <UPGRADE_HEIGHT>
```

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#cosmos-hub-validators-verified` channel of the Cosmos Hub Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand. Requests to join the `#cosmos-hub-validators-verified` channel can be sent to the `#general-support` channel.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it.

## Reference

[Join Cosmos Hub Mainnet](https://github.com/cosmos/mainnet)
