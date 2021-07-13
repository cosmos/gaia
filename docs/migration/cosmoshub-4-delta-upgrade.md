# Cosmos Hub 4, Delta Upgrade, Instructions

This document describes the steps for validator and full node operators for the successful execution of the [Delta Upgrade](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md#Delta-Upgrade), which adds the __Gravity DEX__ to the Cosmos Hub. 

TOC:
- [On-chain governance proposal attains consensus](#on-chain-governance-proposal-attains-consensus)
- [Upgrade will take place July 12, 2021](#upgrade-will-take-place-july-12-2021)
- [Chain-id will remain the same](#chain-id-will-remain-the-same)
- [Preparing for the upgrade](#preparing-for-the-upgrade)
  - [Backups](#backups)
  - [Testing](#testing)
  - [Public testnet](#public-testnet)
  - [Current runtime, cosmoshub-4 (pre-Delta upgrade) is running Gaia v4.2.1](#current-runtime-cosmoshub-4-pre-delta-upgrade-is-running-gaia-v421)
  - [Target runtime, cosmoshub-4 (post-Delta upgrade) will run Gaia v5.0.0](#target-runtime-cosmoshub-4-post-delta-upgrade-will-run-gaia-v500)
- [Delta upgrade steps](#delta-upgrade-steps)
- [Upgrade duration](#upgrade-duration)
- [Rollback plan](#rollback-plan)
- [Communications](#communications)
- [Risks](#risks)
- [FAQ](#faq)

## On-chain governance proposal attains consensus

[Proposal #51](https://www.mintscan.io/cosmos/proposals/51) is the reference on-chain governance proposal for the this upgrade, which has passed with overwhleming community support. Neither core developers nor core funding entities control the governance, and this governance proposal has passed in a _fully decentralized_ way.  

## Upgrade will take place July 12, 2021

The upgrade will take place at a block height of `6910000`. At current block times (around 7s/block), this block height corresponds approximately to `Mon Jul 12 2021 11:00:00 GMT+0000`. This date/time is approximate as blocks are not generated at a constant interval.

## Chain-id will remain the same

The chain-id of the network will remain the same, `cosmoshub-4`. This is because an in-place migration of state will take place, i.e., this upgrade does not export any state.

## Preparing for the upgrade

### Backups

Prior to the upgrade, validators are encouraged to take a full data snapshot. Snapshotting depends heavily on infrastructure, but generally this can be done by backing up the `.gaia` directory.

It is critically important for validator operators to back-up the `.gaia/data/priv_validator_state.json` file after stopping the gaiad process. This file is updated every block as your validator participates in consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

### Testing

For those validator and full node operators that are interested in ensuring preparedness for the impending upgrade, complete and detailed testing instructions are provided in the [gravity-dex-upgrade-test](https://github.com/b-harvest/gravity-dex-upgrade-test/) Github repository. This repository has been tested by members of the core Cosmos ecosystem, as well as ecosystem partners which include validators, exchanges, and service providers.

### Public testnet

Validator and full node operators that wish to test their systems on a public testnet are encouraged to join the Tendermint team's public testnet, described [here](https://github.com/b-harvest/gravity-dex-upgrade-test/#public-testnet-info).

### Current runtime, cosmoshub-4 (pre-Delta upgrade) is running Gaia v4.2.1

The Cosmos Hub mainnet network, `cosmoshub-4`, is currently running [Gaia v4.2.1](https://github.com/cosmos/gaia/releases/tag/v4.2.1). We anticipate that operators who are running earlier versions of Gaia, e.g., v4.2.x, will be able to upgrade successfully; however, this is untested and it is up to operators to ensure that their systems are capable of performing the upgrade. 

### Target runtime, cosmoshub-4 (post-Delta upgrade) will run Gaia v5.0.0

The Comsos Hub mainnet network, `cosmoshub-4`, will run [Gaia v5.0.0](https://github.com/cosmos/gaia/releases/tag/v5.0.0). Operators _MUST_ use this version post-upgrade to remain connected to the network. 

## Delta upgrade steps

The following steps assume that an operator is running v4.2.1 (running an earlier version is untested). The upgrade has only been tested with v4.2.1 and these instructions follow this prerequisite.

1. Prior to the upgrade, operators _MUST_ be running Gaia v4.2.1.
2. At the upgrade block height of [6910000](#Upgrade-will-take-place-July-12,-2021), the Gaia software will panic with a message similar to the below:

> ERR UPGRADE "Gravity-DEX" NEEDED at height: 6910000: v5.0.0-4760cf1f1266accec7a107f440d46d9724c6fd08
>
> panic: UPGRADE "Gravity-DEX" NEEDED at height: 6910000: v5.0.0-4760cf1f1266accec7a107f440d46d9724c6fd08

**IMPORTANT: PLEASE WAIT FOR THE BINARY TO HALT ON ITS OWN**. Do NOT shutdown the node yourself. If the node shuts down before the panic message, start the node and let it run until the panic stops the node for you. 

3. Important note to all validators: Although the upgrade path is essentially to replace the binary when the software panics and halts at the upgrade height, an important disaster recovery operation is to take a snapshot of your state after the halt and before starting v5.0.0.

```bash
cp -r ~/.gaia ./gaia_backup
```

Note: use the home directory relevant to your node's Gaia configuration (if different from `~/.gaia`). 
    
4. Replace the Gaia v4.2.1 binary with the Gaia v5.0.0 binary
5. Start the Gaia v5.0.0 binary using the following command (also applying any additional flags and parameters to the binary needed by the operator, e.g., `--home $HOME`):

> gaiad start --x-crisis-skip-assert-invariants

IMPORTANT: The flag `--x-crisis-skip-assert-invariants` is optional and can be used to reduce memory and processing requirements while the in-place ugprade takes place before resuming connecting to the network.

5. Wait until 2/3+ of voting power has upgraded for the network to start producing blocks
6. You can use the following commands to check peering status and state:

> curl -s http://127.0.0.1:26657/net_info | grep n_peers
> 
> curl -s localhost:26657/consensus_state | jq -r .result.round_state.height_vote_set[].prevotes_bit_array

## Upgrade duration

The upgrade may take several hours to complete because cosmoshub-4 participants operate globally with differing operating hours and it may take some time for operators to upgrade their binaries and connect to the network.

## Rollback plan

During the network upgrade, core Cosmos teams will be keeping an ever vigilant eye and communicating with operators on the status of their upgrades. During this time, the core teams will listen to operator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with operators and attaining social consensus, may choose to declare that the upgrade will be skipped. 

Steps to skip this upgrade proposal are simply to resume the cosmoshub-4 network with the (downgraded) v4.2.1 binary using the following command:

> gaiad start --unsafe-skip-upgrade 6910000

Note: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core Cosmos teams.

Important: A social consensus decision to skip the upgrade will be based solely on technical merits, thereby respecting and maintaining the decentralized governance process of the upgrade proposal's successful YES vote.

## Communications

Operators are encouraged to join the `#validators-verified` channel of the Cosmos Community Discord. This channel is the primary communication tool for operators to ask questions, report upgrade status, report technical issues, and to build social consensus should the need arise. This channel is restricted to known operators and requires verification beforehand - requests to join the `#validators-verified` channel can be sent to the `#validators-public` channel.  

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start before correcting it. 

## FAQ

1. If I am a new operator and I want to join the network, what should I do?

In order to join the cosmoshub-4 network after the Delta upgrade, you have two options:
 - Use a post-delta upgrade state snapshot, such as one provided by [quicksync](https://cosmos.quicksync.io/) and start a node using the gaia v5.0.0 binary. 
 - If not using a snapshot, or using a pre-delta upgrade snapshot, sync with the network using the gaia v4.2.1 binary until the upgrade height and panic, then switch the gaia binary for v5.0.0.
  
2. Does the post-Delta upgrade introduce any changes of note?

The core Cosmos SDK and Tendermint dependencies have only their minor versions bumped, so there are no significant changes of note to the API.

The only integration points that would be affected would be anything that parses all Cosmos SDK messages. The additional messages are [here](https://github.com/Gravity-Devs/liquidity/blob/master/proto/tendermint/liquidity/v1beta1/tx.proto).

3. Is Amino still supported in the post-Delta upgrade?

Amino is still supported. Amino support is still present in the master branch of the Cosmos SDK. No upgrade to remove Amino is currently scheduled.

4. Has the Gravity DEX module undergone a professional 3rd-party audit?

Yes, the audit was led by Least Authority, and have released the [audit report](https://leastauthority.com/blog/audit-of-cosmos-sdk-liquidity-module-for-all-in-bits/).

4. We have some self-healing node infrastructure in place. If the node starts failing when the chain halts, and we automatically spin up another 4.2.1 node with state from within the past couple of hours, is there a risk of it double signing transactions as it "catches up" to the point where block processing stops?

When the network is halted, there is no risk of double-signing since no blocks are being produced. You only need to ensure that the self-healing infrastructure does not launch multiple validators when the network resumes block production. As well, if any new node is spun up while the chain is halted, live peers will continue to share historical blocks without producing new blocks.
