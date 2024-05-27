---
title: Cosmos Hub 4, Gaia v14 Upgrade
order: 9
---
<!-- markdown-link-check-disable -->

This document describes the steps for validators, full node operators and relayer operators, to upgrade successfully for the Gaia v14 release.

For more details on the release, please see the [release notes](https://github.com/cosmos/gaia/releases/tag/v14.1.0)

**Relayer Operators** for the Cosmos Hub and consumer chains, will also need to update to use [Hermes 1.7.3](https://github.com/informalsystems/hermes/releases/tag/v1.7.3) or higher, see [Relayer Operations](#relayer-operations) or more details.

## Release Binary

> Please note that the **v14.0.0** binary is depreceated and **ALL** validators **MUST** use the **v14.1.0** binary instead.

## Instructions
  
- [Cosmos Hub 4, Gaia v14 Upgrade, Instructions](#cosmos-hub-4-gaia-v14-upgrade-instructions)
  - [Release Binary](#release-binary)
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
  - [Upgrade duration](#upgrade-duration)
  - [Rollback plan](#rollback-plan)
  - [Communications](#communications)
  - [Risks](#risks)
  - [Relayer Operations](#relayer-operations)
  - [Reference](#reference)

## On-chain governance proposal attains consensus

[Proposal 854](https://www.mintscan.io/cosmos/proposals/854) is the reference on-chain governance proposal for this upgrade, which is still in its voting period. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.

## Upgrade date

The upgrade will take place at a block height of `18262000`. The date/time of the upgrade is subject to change as blocks are not generated at a constant interval. You can stay up-to-date using this [live countdown](https://www.mintscan.io/cosmos/blocks/18262000) page.

## Chain-id will remain the same

The chain-id of the network will remain the same, `cosmoshub-4`. This is because an in-place migration of state will take place, i.e., this upgrade does not export any state.

## Preparing for the upgrade

System requirements for validator nodes can be found [here](../../../getting-started/system-requirements.md).

### Backups

Prior to the upgrade, validators are encouraged to take a full data snapshot. Snapshotting depends heavily on infrastructure, but generally this can be done by backing up the `.gaia` directory.
If you use Cosmovisor to upgrade, by default, Cosmovisor will backup your data upon upgrade. See below [upgrade using cosmovisor](#method-ii-upgrade-using-cosmovisor) section.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can run a [v14 Local Testnet](https://github.com/cosmos/testnets/tree/master/local) or join in our [Cosmos Hub Public Testnet](https://github.com/cosmos/testnets/tree/master/public).

### Current runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v13.0.0](https://github.com/cosmos/gaia/releases/v13.0.0). We anticipate that operators who are running on v13.0.x, will be able to upgrade successfully. Validators are expected to ensure that their systems are up to date and capable of performing the upgrade. This includes running the correct binary, or if building from source, building with go `1.20`.

### Target runtime

The Cosmos Hub mainnet network, `cosmoshub-4`, will run **[Gaia v14.1.0](https://github.com/cosmos/gaia/releases/tag/v14.1.0)**. Operators _**MUST**_ use this version post-upgrade to remain connected to the network.

## Upgrade steps

There are 2 major ways to upgrade a node:

- Manual upgrade
- Upgrade using [Cosmovisor](https://pkg.go.dev/cosmossdk.io/tools/cosmovisor)
    - Either by manually preparing the new binary
    - Or by using the auto-download functionality (this is not yet recommended)

If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: Manual Upgrade

Make sure **Gaia v14.1.0** is installed by either downloading a [compatible binary](https://github.com/cosmos/gaia/releases/tag/v13.0.0), or building from source. Building from source requires **Golang 1.20.x**.

Run Gaia v13.0.0 till upgrade height, the node will panic:

```shell
ERR UPGRADE "v14" NEEDED at height: 18262000: upgrade to v14 and applying upgrade "v14" at height:18262000
```

Stop the node, and switch the binary to **Gaia v14.1.0** and re-start by `gaiad start`.

It may take several minutes to a few hours until validators with a total sum voting power > 2/3 to complete their node upgrades. After that, the chain can continue to produce blocks.

### Method II: Upgrade using Cosmovisor

### Manually preparing the binary

##### Preparation

Install the latest version of Cosmovisor (`1.5.0`):

```shell
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest
```

**Verify Cosmovisor Version**
```shell
cosmovisor version
cosmovisor version: v1.5.0
```

Create a cosmovisor folder:

create a Cosmovisor folder inside `$GAIA_HOME` and move Gaia v13.0.0 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
````

Build Gaia **v14.1.0**, and move gaiad **v14.1.0** to `$GAIA_HOME/cosmovisor/upgrades/v14/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/v14/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/v14/bin
```

Then you should get the following structure:

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  #v13.0.x
└── upgrades
    └── v14
        └── bin
            └── gaiad  #v14.1.0
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

This may take a few minutes to a few hours.
After upgrade, the chain will continue to produce blocks when validators with a total sum voting power > 2/3 complete their node upgrades.

### Auto-Downloading the Gaia binary

**This method is not recommended!**

## Upgrade duration

The upgrade may take a few minutes to several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v13.0.x binary using the following command:

> gaiad start --unsafe-skip-upgrade 18262000

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#cosmos-hub-validators-verified` channel of the Cosmos Hub Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand. Requests to join the `#cosmos-hub-validators-verified` channel can be sent to the `#general-support` channel.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it.

## Relayer Operations

The Gaia `v14.1.0` upgrade brings forth the cryptographic verification of equivocation feature from ICS `v2.4.0-lsm`. This important security enhancement empowers external agents to promptly submit evidence evidence of light client and double signing attacks observed on a consumer chain. Operators can seize the control of this feature using either the dedicated ICS CLI commands or unleash the power of the Hermes IBC relayer in “evidence” mode. 

This feature is supported by an updated [Hermes v1.7.3](https://github.com/informalsystems/hermes/releases/tag/v1.7.3).

### **1. Hermes “evidence” mode**

Ensure you have a well-configured Hermes `v1.7.3+` relayer effectively relaying packets between a consumer and a provider chain. The following command demonstrates how to run a Hermes instance in “evidence” mode to detect misbehaviors on a consumer chain.

```sh
hermes evidence --chain <CONSUMER-CHAIN-ID>
```

**Tip**: this command takes a `--check-past-blocks` option giving the possibility to look for older evidences (default is `100`).

### **2. ICS CLI**

The ICS provider module offers two commands for submitting evidence of misbehavior originating from a consumer chain. Here are two examples illustrating the process:

To submit evidence of a double-vote:

```sh
gaiad tx provider submit-consumer-double-voting [path/to/evidence.json] [path/to/infraction_header.json] --from node0 --home ../node0 --chain-id $CID
```

And for a light client attack:

```sh
gaiad tx provider submit-consumer-misbehaviour [path/to/misbehaviour.json] --from node0 --home ../node0 --chain-id $CID
```

## Reference

[Join Cosmos Hub Mainnet](https://github.com/cosmos/mainnet)

<!-- markdown-link-check-enable -->
