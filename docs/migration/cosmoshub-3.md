# Cosmos Hub 3 Upgrade Instructions

The following document describes the necessary steps involved that full-node operators
must take in order to upgrade from `cosmoshub-3` to `cosmoshub-4`. The Tendermint team
will post an official updated genesis file, but it is recommended that validators
execute the following instructions in order to verify the resulting genesis file.

There is a strong social consensus around proposal `Cosmos Hub 4 Upgrade Proposal`
on `cosmoshub-3`. Following proposals #[27](https://www.mintscan.io/cosmos/proposals/27), #[35](https://www.mintscan.io/cosmos/proposals/35) and #[36](https://www.mintscan.io/cosmos/proposals/36).
This indicates that the upgrade procedure should be performed on `February 18, 2021 at 06:00 UTC`.

  - [Preliminary](#preliminary)
  - [Major Updates](#major-updates)
  - [Risks](#risks)
  - [Recovery](#recovery)
  - [Upgrade Procedure](#upgrade-procedure)
  - [Notes for Service Providers](#notes-for-service-providers)

## Preliminary

Many changes have occurred to the Cosmos SDK and the Gaia application since the latest
major upgrade (`cosmoshub-3`). These changes notably consist of many new features,
protocol changes, and application structural changes that favor developer ergonomics
and application development.

First and foremost, [IBC](https://docs.cosmos.network/master/ibc/overview.html) following 
the [Interchain Standads](https://github.com/cosmos/ics#ibc-quick-references) will be enabled. 
This upgrade comes with several improvements in efficiency, node synchronization and following blockchain upgrades.
More details on the [Stargate Website](https://stargate.cosmos.network/).

__[Gaia](https://github.com/cosmos/gaia) application v4.0.0 is
what full node operators will upgrade to and run in this next major upgrade__.
Following Cosmos SDK version v0.41.0 and Tendermint v0.34.3.

## Major Updates

There are many notable features and changes in the upcoming release of the SDK. Many of these
are discussed at a high level 
[here](https://github.com/cosmos/stargate).

Some of the biggest changes to take note on when upgrading as a developer or client are the the following:

- **Protocol Buffers**: Initially the Cosmos SDK used Amino codecs for nearly all encoding and decoding. 
In this version a major upgrade to Protocol Buffers have been integrated. It is expected that with Protocol Buffers
applications gain in speed, readability, convinience and interoperability with many programming languages.
[Read more](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md#protocol-buffers)
- **CLI**: The CLI and the daemon for a blockchain were seperated in previous versions of the Cosmos SDK. This 
led to a `gaiad` and `gaiacli` binary which were seperated and could be used for different interactions with the
blockchain. Both of these have been merged into one `gaiad` which now supports the commands the `gaiacli` previously
supported.
- **Node Configuration**: Previously blockchain data and node configuration was stored in `~/.gaiad/`, these will
now reside in `~/.gaia/`, if you use scripts that make use of the configuration or blockchain data, make sure to update the path.

## Risks

As a validator performing the upgrade procedure on your consensus nodes carries a heightened risk of
double-signing and being slashed. The most important piece of this procedure is verifying your
software version and genesis file hash before starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade
procedure again during the network startup. If you discover a mistake in the process, the best thing
to do is wait for the network to start before correcting it. If the network is halted and you have
started with a different genesis file than the expected one, seek advice from a Tendermint developer
before resetting your validator.

## Recovery

Prior to exporting `cosmoshub-3` state, validators are encouraged to take a full data snapshot at the
export height before proceeding. Snapshotting depends heavily on infrastructure, but generally this
can be done by backing up the `.gaiad` directory.

It is critically important to back-up the `.gaiad/data/priv_validator_state.json` file after stopping your gaiad process. This file is updated every block as your validator participates in a consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to
gaia v2.0.15 with v0.37.15 of the _Cosmos SDK_ and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

__Note__: It is assumed you are currently operating a full-node running gaia v2.0.15 with v0.37.15 of the _Cosmos SDK_.

The version/commit hash of Gaia v2.0.15: `89cf7e6fc166eaabf47ad2755c443d455feda02e`

1. Verify you are currently running the correct version (v2.0.15) of _gaiad_:

   ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    client_name: gaiacli
    version: 2.0.15
    commit: 89cf7e6fc166eaabf47ad2755c443d455feda02e
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
   ```

1. Make sure your chain halts at the right time and date:
    February 18, 2021 at 06:00 UTC is in UNIX seconds: `1613628000`

    ```bash
    perl -i -pe 's/^halt-time =.*/halt-time = 1613628000/' ~/.gaiad/config/app.toml
    ```

 1. After the chain has halted, make a backup of your `.gaiad` directory

    ```bash
    mv ~/.gaiad ./gaiad_backup
    ```

    **NOTE**: It is recommended for validators and operators to take a full data snapshot at the export
   height before proceeding in case the upgrade does not go as planned or if not enough voting power
   comes online in a sufficient and agreed upon amount of time. In such a case, the chain will fallback
   to continue operating `cosmoshub-3`. See [Recovery](#recovery) for details on how to proceed.

1. Export existing state from `cosmoshub-3`:

   Before exporting state via the following command, the `gaiad` binary must be stopped!
   As a validator, you can see the last block height created in the 
   `~/.gaiad/config/data/priv_validator_state.json` - or now residing in `gaiad_backup` when you made
    a backup as in the last step - and obtain it with

   ```bash
   cat ~/.gaiad/config/data/priv_validator_state.json | jq '.height'
   ```

   ```bash
   $ gaiad export --for-zero-height --height=<height> > cosmoshub_3_genesis_export.json
   ```
   _this might take a while, you can expect an hour for this step_

1. Verify the SHA256 of the (sorted) exported genesis file:

    Compare this value with other validators / full node operators of the network. 
    Going forward it will be important that all parties can create the same genesis file export.

   ```bash
   $ jq -S -c -M '' cosmoshub_3_genesis_export.json | shasum -a 256
   [SHA256_VALUE]  cosmoshub_3_genesis_export.json
   ```

1. At this point you now have a valid exported genesis state! All further steps now require
v4.0.0 of [Gaia](https://github.com/cosmos/gaia). 
Cross check your genesis hash with other peers (other validators) in the chat rooms.

   **NOTE**: Go [1.15+](https://golang.org/dl/) is required!

   ```bash
   $ git clone https://github.com/cosmos/gaia.git && cd gaia && git checkout v4.0.0; make install
   ```

1. Verify you are currently running the correct version (v4.0.0) of the _Gaia_:

   ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    version: 4.0.0
    commit: 2bb04266266586468271c4ab322367acbf41188f
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
    build_deps:
    ...
   ```
    The version/commit hash of Gaia v4.0.0: `2bb04266266586468271c4ab322367acbf41188f`

1. Migrate exported state from the current v2.0.15 version to the new v4.0.0 version:

   The genesis time shall be the upgrade time of `2021-02-18T06:00:00Z` + `60` minutes with the subseconds truncated.

   This results in our genesis time for starting the blockchain:

    **2021-02-18T07:00:00Z**

   ```bash
   $ gaiad migrate cosmoshub_3_genesis_export.json --chain-id=cosmoshub-4 --genesis-time=2021-02-18T07:00:00Z> genesis.json
   ```

   This will migrate our exported state into the required `genesis.json` file to start the cosmoshub-4.

1. Verify the SHA256 of the final genesis JSON:

   ```bash
   $ jq -S -c -M '' genesis.json | shasum -a 256
   [SHA256_VALUE]  genesis.json
   ```

    Compare this value with other validators / full node operators of the network. 
    It is important that each party can reproduce the same genesis.json file from the steps accordingly.

1. Reset state:

   **NOTE**: Be sure you have a complete backed up state of your node before proceeding with this step.
   See [Recovery](#recovery) for details on how to proceed.

   ```bash
   $ gaiad unsafe-reset-all
   ```

1. Move the new `genesis.json` to your `.gaia/config/` directory

    ```bash
    cp genesis.json ~/.gaia/config/
    ```

1. Start your blockchain 

    ```bash
    gaiad start
    ```

    Automated audits of the genesis state can take 30-120 min using the crisis module. This can be disabled by 
    `gaiad start --x-crisis-skip-assert-invariants`.

## Notes for Service Providers

# REST server

In case you have been running REST server with the command `gaiacli rest-server` previously, running this command will not be necessary anymore.
API server is now in-process with daemon and can be enabled/disabled by API configuration in your `.gaia/config/app.toml`:

```
[api]
# Enable defines if the API server should be enabled.
enable = false
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
```

`swagger` setting refers to enabling/disabling swagger docs API, i.e, /swagger/ API endpoint.

# gRPC Configuration

gRPC configuration in your `.gaia/config/app.toml`

```yaml
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```

# State Sync

State Sync Configuration in your `.gaia/config/app.toml`

```yaml
# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

# Migrations

These chapters contains all the migration guides to update your app and modules to Cosmos v0.40 Stargate.

1. [App and Modules Migration](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md)
1. [Chain Upgrade Guide to v0.40](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/chain-upgrade-guide-040.md)
1. [REST Endpoints Migration](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/rest.md)

If you want to test the procedure before the update happens on 18th of February, please see this post accordingly:

https://github.com/cosmos/gaia/issues/569#issuecomment-767910963
