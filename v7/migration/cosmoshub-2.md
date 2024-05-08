# Cosmos Hub 2 Upgrade Instructions

The following document describes the necessary steps involved that full-node operators
must take in order to upgrade from `cosmoshub-2` to `cosmoshub-3`. The Tendermint team
will post an official updated genesis file, but it is recommended that validators
execute the following instructions in order to verify the resulting genesis file.

There is a strong social consensus around proposal `Cosmos Hub 3 Upgrade Proposal E`
on `cosmoshub-2`. This indicates that the upgrade procedure should be performed
on `December 11, 2019 at or around 14:27 UTC` on block `2,902,000`.

  - [Preliminary](#preliminary)
  - [Major Updates](#major-updates)
  - [Risks](#risks)
  - [Recovery](#recovery)
  - [Upgrade Procedure](#upgrade-procedure)
  - [Notes for Service Providers](#notes-for-service-providers)

## Preliminary

Many changes have occurred to the Cosmos SDK and the Gaia application since the latest
major upgrade (`cosmoshub-2`). These changes notably consist of many new features,
protocol changes, and application structural changes that favor developer ergonomics
and application development.

First and foremost, the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) and the
[Gaia](https://github.com/cosmos/gaia) application have been split into separate
repositories. This allows for both the Cosmos SDK and Gaia to evolve naturally
and independently. Thus, any future [releases](https://github.com/cosmos/gaia/releases)
of Gaia going forward, including this one, will be built and tagged from this
repository not the Cosmos SDK.

Since the Cosmos SDK and Gaia have now been split into separate repositories, their
versioning will also naturally diverge. In an attempt to decrease community confusion and strive for
semantic versioning, the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) will continue
on its current versioning path (i.e. v0.36.x ) and the [Gaia](https://github.com/cosmos/gaia)
application will become v2.0.x.

__[Gaia](https://github.com/cosmos/gaia) application v2.0.3 is
what full node operators will upgrade to and run in this next major upgrade__.

## Major Updates

There are many notable features and changes in the upcoming release of the SDK. Many of these
are discussed at a high level in July's Cosmos development update found
[here](https://blog.cosmos.network/cosmos-development-update-july-2019-8df2ade5ba0a).

Some of the biggest changes to take note on when upgrading as a developer or client are the the following:

- **Tagging/Events**: The entire system of what we used to call tags has been replaced by a more
  robust and flexible system called events. Any client that depended on querying or subscribing to
  tags should take note on the new format as old queries will not work and must be updated. More in
  depth docs on the events system can be found [here](https://github.com/tendermint/tendermint/blob/master/rpc/core/events.go).
  In addition, each module documents its own events in the specs (e.g. [slashing](https://github.com/cosmos/cosmos-sdk/blob/v0.36.0/docs/spec/slashing/06_events.md)).
- **Height Queries**: Both the CLI and REST clients now (re-)enable height queries via the
  `--height` and `?height` arguments respectively. An important note to keep in mind are that height
  queries against pruning nodes will return errors when a pruned height is queried against. When no
  height is provided, the latest height will be used by default keeping current behavior intact. In
  addition, many REST responses now wrap the query results in a new structure `{"height": ..., "result": ...}`.
  That is, the height is now returned to the client for which the resource was queried at.

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

Prior to exporting `cosmoshub-2` state, validators are encouraged to take a full data snapshot at the
export height before proceeding. Snapshotting depends heavily on infrastructure, but generally this
can be done by backing up the `.gaia` directories.

It is critically important to back-up the `.gaia/data/priv_validator_state.json` file after stopping your gaiad process. This file is updated every block as your validator participates in a consensus rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails and the previous chain needs to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to
v0.34.6+ of the _Cosmos SDK_ and restore to their latest snapshot before restarting their nodes.

## Upgrade Procedure

__Note__: It is assumed you are currently operating a full-node running v0.34.6+ of the _Cosmos SDK_.

- The version/commit hash of Gaia v2.0.3: `2f6783e298f25ff4e12cb84549777053ab88749a`
- The upgrade height as agreed upon by governance: **2,902,000**
- You may obtain the canonical UTC timestamp of the exported block by any of the following methods:
  - Block explorer (e.g. [Hubble](https://hubble.figment.network/cosmos/chains/cosmoshub-2/blocks/2902000?format=json&kind=block))
  - Through manually querying an RPC node (e.g. `/block?height=2902000`)
  - Through manually querying a Gaia REST client (e.g. `/blocks/2902000`)

1. Verify you are currently running the correct version (v0.34.6+) of the _Cosmos SDK_:

   ```bash
   $ gaiad version --long
   cosmos-sdk: 0.34.6
   git commit: 80234baf91a15dd9a7df8dca38677b66b8d148c1
   vendor hash: f60176672270c09455c01e9d880079ba36130df4f5cd89df58b6701f50b13aad
   build tags: netgo ledger
   go version go1.12.2 linux/amd64
   ```

2. Export existing state from `cosmoshub-2`:

   **NOTE**: It is recommended for validators and operators to take a full data snapshot at the export
   height before proceeding in case the upgrade does not go as planned or if not enough voting power
   comes online in a sufficient and agreed upon amount of time. In such a case, the chain will fallback
   to continue operating `cosmoshub-2`. See [Recovery](#recovery) for details on how to proceed.

   Before exporting state via the following command, the `gaiad` binary must be stopped!

   ```bash
   $ gaiad export --for-zero-height --height=2902000 > cosmoshub_2_genesis_export.json
   ```

3. Verify the SHA256 of the (sorted) exported genesis file:

   ```bash
   $ jq -S -c -M '' cosmoshub_2_genesis_export.json | shasum -a 256
   [PLACEHOLDER]  cosmoshub_2_genesis_export.json
   ```

4. At this point you now have a valid exported genesis state! All further steps now require
v2.0.3 of [Gaia](https://github.com/cosmos/gaia).

   **NOTE**: Go [1.13+](https://golang.org/dl/) is required!

   ```bash
   $ git clone https://github.com/cosmos/gaia.git && cd gaia && git checkout v2.0.3; make install
   ```

5. Verify you are currently running the correct version (v2.0.3) of the _Gaia_:

   ```bash
   $ gaiad version --long
   name: gaia
   server_name: gaiad
   client_name: gaiacli
   version: 2.0.3
   commit: 2f6783e298f25ff4e12cb84549777053ab88749a
   build_tags: netgo,ledger
   go: go version go1.13.3 darwin/amd64
   ```

6. Migrate exported state from the current v0.34.6+ version to the new v2.0.3 version:

   ```bash
   $ gaiad migrate v0.36 cosmoshub_2_genesis_export.json --chain-id=cosmoshub-3 --genesis-time=[PLACEHOLDER]> genesis.json
   ```

   **NOTE**: The `migrate` command takes an input genesis state and migrates it to a targeted version.
   Both v0.36 and v0.37 are compatible as far as state structure is concerned.

   Genesis time should be computed relative to the blocktime of `2,902,000`. The genesis time
   shall be the blocktime of `2,902,000` + `60` minutes with the subseconds truncated.

   An example shell command(tested on OS X Mojave) to compute this values is:

   ```bash
   curl https://stargate.cosmos.network:26657/block\?height\=2902000 | jq -r '.result["block_meta"]["header"]["time"]'|xargs -0 date -v +60M  -j  -f "%Y-%m-%dT%H:%M:%S" +"%Y-%m-%dT%H:%M:%SZ"
   ```

7. Now we must update all parameters that have been agreed upon through governance. There is only a
single parameter, `max_validators`, that we're upgrading based on [proposal 10](https://www.mintscan.io/proposals/10)

   ```bash
   $ cat genesis.json | jq '.app_state["staking"]["params"]["max_validators"]=125' > tmp_genesis.json && mv tmp_genesis.json genesis.json
   ```

8. Verify the SHA256 of the final genesis JSON:

   ```bash
   $ jq -S -c -M '' genesis.json | shasum -a 256
   [PLACEHOLDER]  genesis.json
   ```

9. Reset state:

   **NOTE**: Be sure you have a complete backed up state of your node before proceeding with this step.
   See [Recovery](#recovery) for details on how to proceed.

   ```bash
   $ gaiad unsafe-reset-all
   ```

10. Move the new `genesis.json` to your `.gaia/config/` directory
11. Replace the `db_backend` on `.gaia/config/config.toml` to:

    ```toml
    db_backend = "goleveldb"
    ```

12. Note, if you have any application configuration in `gaiad.toml`, that file has now been renamed to `app.toml`:

    ```bash
    $ mv .gaia/config/gaiad.toml .gaia/config/app.toml
    ```

## Notes for Service Providers

1. The transition from `cosmoshub-2` to `cosmoshub-3` contains an unusual amount of API breakage.
   After this upgrade will maintain the CosmosSDK API stability guarantee to avoid breaking APIs for at
   least 6 months and hopefully long.
2. Anyone running signing infrastructure(wallets and exchanges) should be conscious that the `type:`
   field on `StdTx` will have changed from `"type":"auth/StdTx","value":...` to  `"type":"cosmos-sdk/StdTx","value":...`
3. As mentioned in the notes and SDK CHANGELOG, many queries to cosmos cli are wrapped with `height` fields now.
4. We highly recommend standing up a [testnet](https://github.com/cosmos/gaia/blob/master/docs/deploy-testnet.md)
   with the `gaia-2.0` release or joining the gaia-13006 testnet. More info for joining the testnet can be
   found in the [riot validator room](https://riot.im/app/#/room/#cosmos-validators:matrix.org).
5. We expect that developers with iOS or Android based apps may have to notify their users of downtime
   and ship an upgrade for cosmoshub-3 compatibility unless they have some kind of switch they can throw
   for the new tx formats. Server side applications should experience briefer service interruptions and
   be able to just spin up new nodes and migrate to the new apis.
