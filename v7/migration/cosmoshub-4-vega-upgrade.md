# Cosmos Hub 4, Vega Upgrade, Instructions


This document describes the steps for validator and full node operators for the successful execution of the [Vega Upgrade](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md#vega-upgrade-expected-q4-2021), which contains the following main new features:
-  [authz](https://github.com/cosmos/cosmos-sdk/tree/v0.44.3/x/authz/spec) and [feegrant modules](https://github.com/cosmos/cosmos-sdk/tree/v0.44.3/x/feegrant/spec)
- [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware)
-  [IBC](https://github.com/cosmos/ibc-go) as a standalone module

TOC:
- [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
- [Upgrade will take place December 14, 2021](#upgrade-will-take-place-december-14-2021)
- [Chain-id will remain the same](#chain-id-will-remain-the-same)
- [Preparing for the upgrade](#preparing-for-the-upgrade)
    - [System requirement](#system-requirement)
    - [Backups](#backups)
    - [Testing](#testing)
    - [Current runtime, cosmoshub-4 (pre-Vega upgrade) is running Gaia v5.0.0](#current-runtime-cosmoshub-4-pre-vega-upgrade-is-running-gaia-v500)
    - [Target runtime, cosmoshub-4 (post-Vega upgrade) will run Gaia v6.0.0](#target-runtime-cosmoshub-4-post-vega-upgrade-will-run-gaia-v600)
- [Vega upgrade steps](#vega-upgrade-steps)
    - [Method I: manual upgrade](#method-i-manual-upgrade)
    -  - [Method II: upgrade using Cosmovisor by manually preparing the Gaia v6.0.0 binary](#method-ii-upgrade-using-cosmovisor-by-manually-preparing-the-gaia-v600-binary)
  - [Method III: upgrade using Cosmovisor by auto-downloading the Gaia v6.0.0 binary (not recommended!)](#method-iii-upgrade-using-cosmovisor-by-auto-downloading-the-gaia-v600-binary-not-recommended)
- [Upgrade duration](#upgrade-duration)
- [Rollback plan](#rollback-plan)
- [Communications](#communications)
- [Risks](#risks)
- [Reference](#reference)


## On-chain governance proposal attains consensus

[Proposal #59](https://www.mintscan.io/cosmos/proposals/59) is the reference on-chain governance proposal for this upgrade, which has passed with overwhleming community support. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.

## Upgrade will take place December 14, 2021

The upgrade will take place at a block height of `8695000`. At the time of writing, and at current block times (around 7s/block), this block height corresponds approximately to `Tuesday, 14-Dec-21 14:49:50 UTC`. This date/time is approximate as blocks are not generated at a constant interval. You can stay up-to-date using this [live countdown](https://chain-monitor.cros-nest.com/d/Upgrades/upgrades?var-chain_id=cosmoshub-4&orgId=1&refresh=1m) page.

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
If you use Cosmovisor to upgrade, by default, Cosmovisor will backup your data upon upgrade. See below [upgrade by cosmovisor](#method-ii-upgrade-using-cosmovisor-by-manually-preparing-the-gaia-v600-binary) section.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, you can join in our [Vega public-testnet](https://github.com/cosmos/vega-test/tree/master/public-testnet) or run a [Vega local testnet](https://github.com/cosmos/vega-test/tree/master/local-testnet).

### Current runtime, cosmoshub-4 (pre-Vega upgrade) is running Gaia v5.0.0

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v5.0.0](https://github.com/cosmos/gaia/releases/tag/v5.0.0). We anticipate that operators who are running on v5.0.x, will be able to upgrade successfully; however, this is untested and it is up to operators to ensure that their systems are capable of performing the upgrade.

### Target runtime, cosmoshub-4 (post-Vega upgrade) will run Gaia v6.0.0

The Comsos Hub mainnet network, `cosmoshub-4`, will run [Gaia v6.0.0](https://github.com/cosmos/gaia/releases/tag/v6.0.0). Operators _MUST_ use this version post-upgrade to remain connected to the network.

## Vega upgrade steps
There are 2 major ways to upgrade a node:
- Manual upgrade
- Upgrade using [Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)
  - Either by manually preparing the new binary
  - Or by using the auto-download functionality (this is not yet recommended)
    
If you prefer to use Cosmovisor to upgrade, some preparation work is needed before upgrade.

### Method I: manual upgrade
Run Gaia v5.0.x till upgrade height, the node will panic:
```shell
ERR UPGRADE "Vega" NEEDED at height: 8695000

panic: UPGRADE "Vega" NEEDED at height: 8695000
```
Stop the node, and install Gaia v6.0.0 and re-start by `gaiad start`.

It may take 20 min to a few hours until validators with a total sum voting power > 2/3 to complete their nodes upgrades. After that, the chain can continue to produce blocks.

### Method II: upgrade using Cosmovisor by manually preparing the Gaia v6.0.0 binary
#### Preparation

Install the latest version of Cosmovisor:
```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest
```

Create a cosmovisor folder:

create a Cosmovisor folder inside `$GAIA_HOME` and move Gaia v5.0.0 into ` $GAIA_HOME/cosmovisor/genesis/bin`
```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
````
build Gaia v6.0.0, and move gaiad v6.0.0 to `$GAIA_HOME/cosmovisor/upgrades/Vega/bin`

```shell
mkdir -p  $GAIA_HOME/cosmovisor/upgrades/Vega/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/upgrades/Vega/bin
```
Then you should get the following structure:
```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad  #v5.0.x
└── upgrades
└── Vega
└── bin
    └── gaiad  #v6.0.0
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
 
#### Expected ugprade result
When the upgrade block height is reached, you can find the following information in the log: 
```shell
ERR UPGRADE "Vega" NEEDED at height: 8695000: upgrade to Vega and applying upgrade "Vega" at height:8695000.
```
 This may take 20 min to a few hours.
 After this, the chain will continue to produce blocks when validators with a total sum voting power > 2/3 complete their nodes upgrades.

### Method III: upgrade using Cosmovisor by auto-downloading the Gaia v6.0.0 binary (not recommended!)
#### Preparation
Install Cosmovisor v0.1
```shell
go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v0.1.0
```
Create a cosmovisor folder:

create a cosmovisor folder inside gaia home and move gaiad v5.0.0 into ` $GAIA_HOME/cosmovisor/genesis/bin`
```shell
mkdir -p $GAIA_HOME/cosmovisor/genesis/bin
cp $(which gaiad) $GAIA_HOME/cosmovisor/genesis/bin
```

```shell
.
├── current -> genesis or upgrades/<name>
└── genesis
     └── bin
        └── gaiad  #v5.0.x
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
ERR UPGRADE "Vega" NEEDED at height: 8695000: upgrade to Vega and applying upgrade "Vega" at height:8695000
```
Then the Cosmovisor will create `$GAIA_HOME/cosmovisor/upgrades/Vega/bin` and download Gaia v6.0.0 binary to this folder according to links in the ` --info` field of the upgrade proposal 59.
This may take 20 min to a few hours, afterwards, the chain will continue to produce blocks once validators with a total sum voting power > 2/3 complete their nodes upgrades.

*Please Note:*

Auto-download the new binary is not recommended for the following reasons:
- In general, auto-download comes with the risk that the verification of correct download is done automatically. If you want to have the highest guarantee you should confirm the check-sum manually. We hope more node operators will use the auto-download for this release but please be aware this is a risk you should take at your own discretion.
- For the Vega upgrade, Gaia will upgrade its dependency on Cosmos SDK v0.42 to Cosmos SDK v0.44, this will require [Cosmovisor v0.1](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv0.1.0). Later versions of Cosmovisor do not support Cosmos SDK v0.42 or earlier if the auto-download option is enabled.
- By using Cosmovisor v0.1 you might experience a [node hanging issue](https://github.com/cosmos/cosmos-sdk/issues/9875) when querying a result with a large output size. For example, `gaiad q gov proposals` will hang the node being queried, this issue will not appear for Cosmovisor versions newer than v0.1.

## Upgrade duration

The upgrade may take several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped.

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v5.0.x binary using the following command:

> gaiad start --unsafe-skip-upgrade 8695000

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#validators-verified` channel of the Cosmos Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand. Requests to join the `#validators-verified` channel can be sent to the `#general-support` channel.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it.

## Reference
[cosmos/vega-test](https://github.com/cosmos/vega-test)

[Delta upgrade instruction](https://github.com/cosmos/gaia/blob/main/docs/migration/cosmoshub-4-delta-upgrade.md)
