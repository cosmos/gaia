---
title: Cosmos Hub 4, Lambda Upgrade
order: 1
---
<!-- markdown-link-check-disable -->
# Cosmos Hub 4, v9-Lambda Upgrade, Instructions

This document describes the steps for validator and full node operators for the successful execution of the [v9-Lambda Upgrade](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md#v9-lambda-upgrade-expected-q1-2023), which contains the following main new features/improvement:

- [Interchain-Security](https://github.com/cosmos/interchain-security) [v1.0.0](https://github.com/cosmos/interchain-security/releases/tag/v1.0.0) provider module. See the [ICS Spec](https://github.com/cosmos/ibc/blob/main/spec/app/ics-028-cross-chain-validation/README.md) for more details.
- [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.13-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.13-ics). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/releases/tag/v0.45.13-ics) for details.
- [ibc-go](https://github.com/cosmos/ibc-go) to [v4.2.0](https://github.com/cosmos/ibc-go/blob/release/v4.2.x/CHANGELOG.md). See [v4.2 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v4.2.0) for details.
- [tendermint](https://github.com/informalsystems/tendermint) to [0.34.26](https://github.com/informalsystems/tendermint/tree/v0.34.26). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.26/CHANGELOG.md#v03426) for details.
- [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v4.0.4](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v4.0.4).
- [E2E ccv tests](https://github.com/cosmos/gaia/blob/main/tests/e2e/e2e_gov_test.go#L138). Tests covering new functionality introduced by the provider module to add and remove a consumer chain via governance proposal.
- [integration ccv tests](https://github.com/cosmos/gaia/blob/main/tests/ics/interchain_security_test.go). Imports Interchain-Security's `TestCCVTestSuite` and implements Gaia as the provider chain.

TOC:

- [Cosmos Hub 4, v9-Lambda Upgrade, Instructions](#cosmos-hub-4-v9-lambda-upgrade-instructions)
    - [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
    - [Upgrade will take place March 15, 203](#upgrade-will-take-place-march-15-2023)
    - [Chain-id will remain the same](#chain-id-will-remain-the-same)
    - [Preparing for the upgrade](#preparing-for-the-upgrade)
        - [System requirement](#system-requirement)
        - [Backups](#backups)
        - [Testing](#testing)
        - [Current runtime, cosmoshub-4 (pre-v9-Lambda upgrade) is running Gaia v8.0.1](#current-runtime-cosmoshub-4-pre-v9-lambda-upgrade-is-running-gaia-v801)
        - [Target runtime, cosmoshub-4 (post-v9-Lambda upgrade) will run Gaia v9.0.0](#target-runtime-cosmoshub-4-post-v9-lambda-upgrade-will-run-gaia-v900)
    - [v9-Lambda upgrade steps](#v9-Lambda-upgrade-steps)
        - [Method I: Manual Upgrade](#method-i-manual-upgrade)
        - [Method II: Upgrade using Cosmovisor](#method-ii-upgrade-using-cosmovisor)
            - [Manually preparing the binary](#manually-preparing-the-gaia-v900-binary)
                - [Preparation](#preparation)
                - [Expected upgrade result](#expected-upgrade-result)
            - [Auto-Downloading the Gaia v9.0.0 binary (not recommended!)](#auto-downloading-the-gaia-v900-binary-not-recommended)
                - [Preparation](#preparation-1)
                - [Expected result](#expected-result)
    - [Upgrade duration](#upgrade-duration)
    - [Rollback plan](#rollback-plan)
    - [Communications](#communications)
    - [Risks](#risks)
    - [Reference](#reference)

## On-chain governance proposal attains consensus

[Proposal #187](https://www.mintscan.io/cosmos/proposals/187) is the reference on-chain governance proposal for this upgrade, which is still in its voting period. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.

## Upgrade will take place March 14-16, 2023

The upgrade will take place at a block height of `14470501`. The date/time of the upgrade is subject to change as blocks are not generated at a constant interval. You can stay up-to-date using this [live countdown](https://www.mintscan.io/cosmos/blocks/14470501) page.

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

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can run a [v8-Rho local testnet](https://github.com/cosmos/testnets/tree/master/local) or join in our [v9-Lambda public-testnet](https://github.com/cosmos/testnets/tree/master/public).

### Current runtime, cosmoshub-4 (pre-v9-Lambda upgrade) is running Gaia v8.0.1

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v8.0.1](https://github.com/cosmos/gaia/releases/v8.0.1). We anticipate that operators who are running on v8.0.1, will be able to upgrade successfully. Validators are expected to ensure that their systems are up to date and capable of performing the upgrade. This includes running the correct binary, or if building from source, building with go `1.18`.

### Target runtime, cosmoshub-4 (post-v9-Lambda upgrade) will run Gaia v9.0.0

The Cosmos Hub mainnet network, `cosmoshub-4`, will run [Gaia v9.0.0](https://github.com/cosmos/gaia/releases/tag/v9.0.0). Operators _MUST_ use this version post-upgrade to remain connected to the network.

## v9-Lambda upgrade steps

There are 2 major ways to upgrade a node:

- Manual upgrade
- Upgrade using [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)
    - Either by manually preparing the new binary
    - Or by using the auto-download functionality (this is not yet recommended)

If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: Manual Upgrade

Make sure Gaia v9.0.0 is installed by either downloading a [compatable binary](https://github.com/cosmos/gaia/releases/tag/v9.0.0), or building from source. Building from source requires go 1.18.

Run Gaia v8.0.1 till upgrade height, the node will panic:

```shell
ERR UPGRADE "v9-Lambda" NEEDED at height: 14470501: upgrade to v9-Lambda and applying upgrade "v9-Lambda" at height:14470501
```

Stop the node, and switch the binary to Gaia v9.0.0 and re-start by `gaiad start`.

It may take several minutes to a few hours until validators with a total sum voting power > 2/3 to complete their node upgrades. After that, the chain can continue to produce blocks.

### Method II: Upgrade using Cosmovisor

::: warning
<span style="color:red">**Please Read Before Proceeding**</span><br>
Using Cosmovisor 1.2.0 and higher requires a lowercase naming convention for upgrade version directory. For Cosmovisor 1.1.0 and earlier, the upgrade version is not lowercased.
:::

> **For Example:** <br>
> **Cosmovisor =< `1.1.0`: `/upgrades/v9-Lambda/bin/gaiad`**<br>
> **Cosmovisor >= `1.2.0`: `/upgrades/v9-lambda/bin/gaiad`**<br>

| Cosmovisor Version | Binary Name in Path |
|--------------------|---------------------|
| 1.3                | v9-lambda           |
| 1.2                | v9-lambda           |
| 1.1                | v9-Lambda           |
| 1.0                | v9-Lambda           |


### _Manually preparing the Gaia v9.0.0 binary_

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

create a Cosmovisor folder inside `$GAIA_HOME` and move Gaia v8.0.1 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
````

build Gaia v9.0.0, and move gaiad v9.0.0 to `$GAIA_HOME/cosmovisor/upgrades/v9-lambda/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/v9-lambda/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/v9-lambda/bin
```

Then you should get the following structure:

```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  #v8.0.1
└── upgrades
    └── v9-lambda
        └── bin
            └── gaiad  #v9.0.0
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

### _Auto-Downloading the Gaia v9.0.0 binary (not recommended!)_
#### Preparation

Install the latest version of Cosmovisor (`1.3.0`):

```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

Create a cosmovisor folder:

create a cosmovisor folder inside gaia home and move gaiad v8.0.1 into `$GAIA_HOME/cosmovisor/genesis/bin`

```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
```

```shell
.
├── current -> genesis or upgrades/<name>
└── genesis
     └── bin
        └── gaiad  #v8.0.1
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
ERR UPGRADE "v9-Lambda" NEEDED at height: 14470501: upgrade to v9-Lambda and applying upgrade "v9-Lambda" at height:14470501
```

Then the Cosmovisor will create `$GAIA_HOME/cosmovisor/upgrades/v9-lambda/bin` and download the Gaia v9.0.0 binary to this folder according to links in the `--info` field of the upgrade proposal 97.
This may take 7 minutes to a few hours, afterwards, the chain will continue to produce blocks once validators with a total sum voting power > 2/3 complete their nodes upgrades.

_Please Note:_

- In general, auto-download comes with the risk that the verification of correct download is done automatically. If users want to have the highest guarantee users should confirm the check-sum manually. We hope more node operators will use the auto-download for this release but please be aware this is a risk and users should take at your own discretion.
- Users should use run node on v8.0.1 if they use the cosmovisor v1.3.0 with auto-download enabled for upgrade process.

## Upgrade duration

The upgrade may take a few minutes to several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v8.0.1 binary using the following command:

> gaiad start --unsafe-skip-upgrade 14470501

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
