<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes.
"State Machine Breaking" for breaking the AppState

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

## [v7.0.3] -2022-07-28

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.6). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.6/CHANGELOG.md#v0456---2022-06-28) for details.
* (gaia) bump [liquidity](https://github.com/gravity-devs/liquidity) to [v1.5.1](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.1) to be compatible with cosmos-sdk v0.45.6.

## [v7.0.2] -2022-05-09

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.4). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.4/CHANGELOG.md#v0454---2022-04-25) for details.
* (gaia) [#1447](https://github.com/cosmos/gaia/pull/1447) Support custom message types to bypass minimum fee checks for. If a transaction contains only bypassed message types, the transaction will not have minimum fee checks performed during `CheckTx`. Operators can supply these message types via the `bypass-min-fee-msg-types`
  configuration in `app.toml`. Note, by default they include various IBC message types.

## [v7.0.1] -2022-04-13

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.3). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.3/CHANGELOG.md#v0453---2022-04-12) for details.
* (gaia) bump [tendermint](https://github.com/tendermint/tendermint) to [0.34.19](https://github.com/tendermint/tendermint/tree/v0.34.19). See [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/v0.34.19/CHANGELOG.md#v03419) for details.
* (gaia) bump [tm-db](https://github.com/tendermint/tm-db) to [v0.6.7](https://github.com/tendermint/tm-db/tree/v0.6.7). See [CHANGELOG.md](https://github.com/tendermint/tm-db/blob/v0.6.7/CHANGELOG.md#067) for details.

## [v7.0.0] - 2022-03-24

* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.1/CHANGELOG.md#v0451---2022-02-03) for details.
* (gaia) bump [ibc-go](https://github.com/cosmos/ibc-go) module to [v3.0.0](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0). See [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v3.0.0/CHANGELOG.md#v300---2022-03-15) for details.
* (gaia) add [interchain account](https://github.com/cosmos/ibc-go/tree/main/modules/apps/27-interchain-accounts) module (interhchain-account module is part of ibc-go module).
* (gaia) bump [liquidity](https://github.com/gravity-devs/liquidity) module to [v1.5.0](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.0). See [CHANGELOG.md](https://github.com/Gravity-Devs/liquidity/blob/v1.5.0/CHANGELOG.md#v150---20220223) for details.
* (gaia) bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) module to [v2.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v2.1.1).
* (gaia) add migration logs for upgrade process.

## [v6.0.4] - 2022-03-10

* (gaia) Bump [Liquidity](https://github.com/gravity-devs/liquidity) module to [v1.4.6](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.4.6).
* (gaia) Bump [IBC](https://github.com/cosmos/ibc-go) module to [2.0.3](https://github.com/cosmos/ibc-go/releases/tag/v2.0.3).
* (gaia) [#1230](https://github.com/cosmos/gaia/pull/1230) Fix: update gRPC Web Configuration in `contrib/testnets/test_platform`.
* (gaia) [#1135](https://github.com/cosmos/gaia/pull/1135) Fix rocksdb build tag usage.
* (gaia) [#1160](https://github.com/cosmos/gaia/pull/1160) Improvement: update state sync configs.
* (gaia) [#1208](https://github.com/cosmos/gaia/pull/1208) Update statesync.bash.
* * (gaia) Bump [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) to [v0.44.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.6)
* (gaia) Bump [Versions](https://github.com/cosmos/gaia/pull/1100) of various smaller dependencies, remove the Cosmos SDK replace statement, update `initiClientCtx` params, ensure `stdout` and `stderr` are handled correctly in the CLI.

## [v6.0.3] - 2022-02-18

 * This is a reverted release that is the same as v6.0.0

## [v6.0.2] - 2022-02-17

 * Unusable release

## [v6.0.1] - 2022-02-10

 * Unusable release

## [v6.0.0] - 2021-11-24

 * (gaia) Add NewSetUpContextDecorator to anteDecorators
 * (gaia) Reconfigure SetUpgradeHandler to ensure vesting is configured after auth and new modules have InitGenesis run.
 * (golang) Bump golang prerequisite to 1.17. 
 * (gaia) Bump [Liquidity](https://github.com/gravity-devs/liquidity) module to [v1.4.2](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.4.2).
 * (gaia) Bump [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) to [v0.44.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.3). See the [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.44.x/CHANGELOG.md#v0443---2021-10-21) for details.
 * (gaia) Add [IBC](https://github.com/cosmos/ibc-go) as a standalone module from the Cosmos SDK using version [v2.0.0](https://github.com/cosmos/ibc-go/releases/tag/v2.0.0). See the [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v2.0.0/CHANGELOG.md) for details.
 * (gaia) Add [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) [v1.0.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v1.0.1).
 * (gaia) [#969](https://github.com/cosmos/gaia/issues/969) Remove legacy migration code.

## [v5.0.8] - 2021-10-14

* (gaia) This release includes a new AnteHandler that rejects redundant IBC transactions to save relayers fees.

## [v5.0.8] - 2021-10-14

* (gaia) This release includes a new AnteHandler that rejects redundant IBC transactions to save relayers fees.

## [v5.0.7] - 2021-09-30

  * (gaia) Bump Cosmos SDK to 0.42.10

## [v5.0.6] - 2021-09-16

 * (gaia) Bump tendermint to 0.34.13

 
## [v5.0.5] - 2021-08-05

 * (gaia) Bump SDK to [0.42.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.9) to resolve IBC channel restart issue ([9800](https://github.com/cosmos/cosmos-sdk/issues/9800)).

## [v5.0.4] - 2021-07-31
 * (chore) Fix release to include intended items from `v5.0.3`.

## [v5.0.3] - 2021-07-30

* (gaia) Bump SDK to [0.42.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.8) to resolve tx query issues.
* (gaia) Bump SDK to [0.42.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.7) to resolve state sync issues.

## [v5.0.2] - 2021-07-15

* (gaia) Fix misspelling in RELEASE.md
* (gaia) Add releases to .gitignore

## [v5.0.1] - 2021-07-15

* (gaia) Configure gaiad command to add back `config` capabilities.

## [v5.0.0] - 2021-06-28

* (golang) Bump golang prerequisite from 1.15 to 1.16.
* (gaia) Add [Liquidity](https://github.com/gravity-devs/liquidity) module [v1.2.9](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.2.9).
* (sdk)  Bump SDK version to [v0.42.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.6).
* (tendermint) Bump Tendermint version to [v0.34.11](https://github.com/tendermint/tendermint/releases/tag/v0.34.11).

## [v4.2.1] - 2021-04-08

A critical security vulnerability was identified in Tendermint Core, which impacts Tendermint Lite Client.

This release fixes the identified security vulnerability.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.4)
* (tendermint) Bump Tendermint version to [v0.34.9](https://github.com/tendermint/tendermint/releases/tag/v0.34.9).

## [v4.2.0] - 2021-03-25

A critical security vulnerability has been identified in Gaia v4.1.x.
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This release fixes the identified security vulnerability.

If the chain halts before or during the upgrade, validators with sufficient voting power need to upgrade
and come online in order for the chain to resume.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.3)

## [v4.1.2] - 2021-03-22

This release removes unnecessary dependencies.

### Bug Fixes

* (gaia)  [\#781](https://github.com/cosmos/gaia/pull/781) Remove unnecessary dependencies

## [v4.1.1] - 2021-03-19

This release bring improvements to keyring UX, tx search results, and multi-sig account migrations.
See the Cosmos SDK [release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.2) for details.

### Bug Fixes

* (sdk)  Bump SDK version to [v0.42.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.2)

## [v4.1.0] - 2021-03-10

### Bug Fixes

A critical security vulnerability has been identified in Gaia v4.0.x.
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This release fixes the identified security vulnerability.

If the chain halts before or during the upgrade, validators with sufficient voting power need to upgrade
and come online in order for the chain to resume.

## [v4.0.6] - 2021-03-09

### Bug Fixes

This release bumps the Cosmos SDK, which includes an important security fix for all non
Cosmos Hub chains (e.g. any chain that does not use the default cosmos bech32 prefix),
and a few performance improvements. The SDK also applies a security fix for validator
address conversion in evidence handling, and the full header is now emitted on an
IBC UpdateClient message event.

* (sdk)  Bump SDK version to [v0.42.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.0)

## [v4.0.5] - 2021-03-02

### Bug Fixes

* (tendermint) Bump Tendermint version to [v0.34.8](https://github.com/tendermint/tendermint/releases/tag/v0.34.8).
* (sdk)  Bump SDK version to [v0.41.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.4), which reduces startup time with the cosmoshub-4 mainnet genesis without invariant checks.

## [v4.0.4] - 2021-02-19

### Bug Fixes

This release applies a patched version to grpc dependencies in order to resolve some queries; no explicit version bumps are included.

## [v4.0.3] - 2021-02-18

### Bug Fixes

This release fixes build failures caused by a small API breakage introduced in tendermint v0.34.7.

* (sdk)  Bump SDK version to [v0.41.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.3).

## [v4.0.2] - 2021-02-18

### Bug Fixes

This release fixes a downstream security issue which impacts Cosmos SDK users.
See the [Tendermint v0.34.7 SDK changelog](https://github.com/tendermint/tendermint/blob/v0.34.x/CHANGELOG.md#v0347) for details.

* (sdk) [\#640](https://github.com/cosmos/gaia/pull/640) Bump SDK version to [v0.41.2](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.2).
* (tendermint) [\#640](https://github.com/cosmos/gaia/pull/640) Bump Tendermint version to [v0.34.7](https://github.com/tendermint/tendermint/releases/tag/v0.34.7).

## [v4.0.1] - 2021-02-17

### Bug Fixes

* (sdk) [\#579](https://github.com/cosmos/gaia/pull/635) Bump SDK version to [v0.41.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.1).
* (tendermint) [\#622](https://github.com/cosmos/gaia/pull/622) Bump Tendermint version to [v0.34.4](https://github.com/tendermint/tendermint/releases/tag/v0.34.4).

## [v4.0.0] - 2021-01-26

### Improvements

* (app) [\#564](https://github.com/cosmos/gaia/pull/564) Add client denomination metadata for atoms.

### Bug Fixes

* (cmd) [\#563](https://github.com/cosmos/gaia/pull/563) Add balance coin to supply when adding a new genesis account
* (sdk) [\#579](https://github.com/cosmos/gaia/pull/579) Bump SDK version to [v0.41.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.0).

## [v3.0.1] - 2021-01-19

### Improvements

* (protobuf) [\#553](https://github.com/cosmos/gaia/pull/553) Bump gogo protobuf deps to v1.3.3
* (github) [\#543](https://github.com/cosmos/gaia/pull/543) Add docker deployment
* (starport) [\#535](https://github.com/cosmos/gaia/pull/535) Add config.yml
* (docker) [\#534](https://github.com/cosmos/gaia/pull/534) Update to python3

### Bug Fixes

* (sdk) Bump SDK version to [v0.40.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.1).
* (tendermint) Bump Tendermint version to [v0.34.3](https://github.com/tendermint/tendermint/releases/tag/v0.34.3).
* (github) [\#544](https://github.com/cosmos/gaia/pull/544) Deploy from main not master
* (docs) [\#550](https://github.com/cosmos/gaia/pull/550) Bump vuepress-theme-cosmos to 1.0.180
* (docker) [\#537](https://github.com/cosmos/gaia/pull/537) Fix single-node.sh setup script

## [v3.0.0] - 2021-01-09

### Improvements

* (sdk) Bump SDK version to [v0.40.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.0).
* (tendermint) Bump Tendermint version to [v0.34.1](https://github.com/tendermint/tendermint/releases/tag/v0.34.1).

## [v2.0.14] - 2020-12-10

* (sdk) Bump SDK version to [v0.37.15](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.15).
* (tendermint) Bump Tendermint version to [v0.32.14](https://github.com/tendermint/tendermint/releases/tag/v0.32.14).

## [v2.0.13] - 2020-08-13

* (sdk) Bump SDK version to [v0.37.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.14).

## [v2.0.12] - 2020-08-13

* This version did not contain the update to [v0.37.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.14). Please use v2.0.13

## [v2.0.11] - 2020-05-06

* (sdk) Bump SDK version to [v0.37.13](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.13).

## [v2.0.10] - 2020-05-06

* (sdk) Bump SDK version to [v0.37.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.12).

## [v2.0.9] - 2020-04-23

* (sdk) Bump SDK version to [v0.37.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.11).

## [v2.0.8] - 2020-04-09

### Improvements

* (sdk) Bump SDK version to [v0.37.9](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.9).

## [v2.0.7] - 2020-03-11

### Improvements

* (sdk) Bump SDK version to [v0.37.8](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.8).

## [v2.0.6] - 2020-02-10

### Improvements

* (sdk) Bump SDK version to [v0.37.7](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.7).

## [v2.0.5] - 2020-01-21

### Improvements

* (sdk) Bump SDK version to [v0.37.6](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.6).
* (tendermint) Bump Tendermint version to [v0.32.9](https://github.com/tendermint/tendermint/releases/tag/v0.32.9).

## [v2.0.4] - 2020-01-09

### Improvements

* (sdk) Bump SDK version to [v0.37.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.5).
* (tendermint) Bump Tendermint version to [v0.32.8](https://github.com/tendermint/tendermint/releases/tag/v0.32.8).

### Bug Fixes

* (cli) Fixed `gaiacli query txs` to use `events` instead of `tags`. Events take the form of `'{eventType}.{eventAttribute}={value}'`. Please
  see the [events doc](https://github.com/cosmos/cosmos-sdk/blob/master/docs/core/events.md#events-1)
  for further documentation.

## [v2.0.3] - 2019-11-04

### Improvements

* (sdk) Bump SDK version to [v0.37.4](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.4).
* (tendermint) Bump Tendermint version to [v0.32.7](https://github.com/tendermint/tendermint/releases/tag/v0.32.7).

## [v2.0.2] - 2019-10-12

### Improvements

* (sdk) Bump SDK version to [v0.37.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.3).
* (tendermint) Bump Tendermint version to [v0.32.6](https://github.com/tendermint/tendermint/releases/tag/v0.32.6).

## [v2.0.1] - 2019-09-20

### Features

* (gaiad) [\#119](https://github.com/cosmos/gaia/pull/119) Add support for the `--halt-time` CLI flag and configuration.

### Improvements

* [\#119](https://github.com/cosmos/gaia/pull/119) Refactor and upgrade Circle CI
  configuration.
* (sdk) Update SDK version to v0.37.1

## [v2.0.0] - 2019-08-22

### Bug Fixes

* [\#104](https://github.com/cosmos/gaia/issues/104) Fix `ModuleAccountAddrs` to
  not rely on the `x/supply` keeper to get module account addresses for blacklisting.

### Breaking Changes

* (sdk) Update SDK version to v0.37.0

## [v1.0.0] - 2019-08-13

### Bug Fixes

* (gaiad) [\#4113](https://github.com/cosmos/cosmos-sdk/issues/4113) Fix incorrect `$GOBIN` in `Install Go`
* (gaiacli) [\#3945](https://github.com/cosmos/cosmos-sdk/issues/3945) There's no check for chain-id in TxBuilder.SignStdTx
* (gaiacli) [\#4190](https://github.com/cosmos/cosmos-sdk/issues/4190) Fix redelegations-from by using the correct params and query endpoint.
* (gaiacli) [\#4219](https://github.com/cosmos/cosmos-sdk/issues/4219) Return an error when an empty mnemonic is provided during key recovery.
* (gaiacli) [\#4345](https://github.com/cosmos/cosmos-sdk/issues/4345) Improved Ledger Nano X detection

### Breaking Changes

* (sdk) Update SDK version to v0.36.0
* (gaiad) [\#3985](https://github.com/cosmos/cosmos-sdk/issues/3985) ValidatorPowerRank uses potential consensus power
* (gaiad) [\#4027](https://github.com/cosmos/cosmos-sdk/issues/4027) gaiad version command does not return the checksum of the go.sum file shipped along with the source release tarball.
  Go modules feature guarantees dependencies reproducibility and as long as binaries are built via the Makefile shipped with the sources, no dependendencies can break such guarantee.
* (gaiad) [\#4159](https://github.com/cosmos/cosmos-sdk/issues/4159) use module pattern and module manager for initialization
* (gaiad) [\#4272](https://github.com/cosmos/cosmos-sdk/issues/4272) Merge gaiareplay functionality into gaiad replay.
  Drop `gaiareplay` in favor of new `gaiad replay` command.
* (gaiacli) [\#3715](https://github.com/cosmos/cosmos-sdk/issues/3715) query distr rewards returns per-validator
  rewards along with rewards total amount.
* (gaiacli) [\#40](https://github.com/cosmos/cosmos-sdk/issues/40) rest-server's --cors option is now gone.
* (gaiacli) [\#4027](https://github.com/cosmos/cosmos-sdk/issues/4027) gaiacli version command dooes not return the checksum of the go.sum file anymore.
* (gaiacli) [\#4142](https://github.com/cosmos/cosmos-sdk/issues/4142) Turn gaiacli tx send's --from into a required argument.
  New shorter syntax: `gaiacli tx send FROM TO AMOUNT`
* (gaiacli) [\#4228](https://github.com/cosmos/cosmos-sdk/issues/4228) Merge gaiakeyutil functionality into gaiacli keys.
  Drop `gaiakeyutil` in favor of new `gaiacli keys parse` command. Syntax and semantic are preserved.
* (rest) [\#3715](https://github.com/cosmos/cosmos-sdk/issues/3715) Update /distribution/delegators/{delegatorAddr}/rewards GET endpoint
  as per new specs. For a given delegation, the endpoint now returns the
  comprehensive list of validator-reward tuples along with the grand total.
* (rest) [\#3942](https://github.com/cosmos/cosmos-sdk/issues/3942) Update pagination data in txs query.
* (rest) [\#4049](https://github.com/cosmos/cosmos-sdk/issues/4049) update tag MsgWithdrawValidatorCommission to match type
* (rest) The `/auth/accounts/{address}` now returns a `height` in the response. The
  account is now nested under `account`.

### Features

* (gaiad) Add `migrate` command to `gaiad` to provide the ability to migrate exported
  genesis state from one version to another.
* (gaiad) Update Gaia for community pool spend proposals per Cosmos Hub governance proposal [\#7](https://github.com/cosmos/cosmos-sdk/issues/7) "Activate the Community Pool"

### Improvements

* (gaiad) [\#4042](https://github.com/cosmos/cosmos-sdk/issues/4042) Update docs and scripts to include the correct `GO111MODULE=on` environment variable.
* (gaiad) [\#4066](https://github.com/cosmos/cosmos-sdk/issues/4066) Fix 'ExportGenesisFile() incorrectly overwrites genesis'
* (gaiad) [\#4064](https://github.com/cosmos/cosmos-sdk/issues/4064) Remove `dep` and `vendor` from `doc` and `version`.
* (gaiad) [\#4080](https://github.com/cosmos/cosmos-sdk/issues/4080) add missing invariants during simulations
* (gaiad) [\#4343](https://github.com/cosmos/cosmos-sdk/issues/4343) Upgrade toolchain to Go 1.12.5.
* (gaiacli) [\#4068](https://github.com/cosmos/cosmos-sdk/issues/4068) Remove redundant account check on `gaiacli`
* (gaiacli) [\#4227](https://github.com/cosmos/cosmos-sdk/issues/4227) Support for Ledger App v1.5
* (rest) [\#2007](https://github.com/cosmos/cosmos-sdk/issues/2007) Return 200 status code on empty results
* (rest) [\#4123](https://github.com/cosmos/cosmos-sdk/issues/4123) Fix typo, url error and outdated command description of doc clients.
* (rest) [\#4129](https://github.com/cosmos/cosmos-sdk/issues/4129) Translate doc clients to chinese.
* (rest) [\#4141](https://github.com/cosmos/cosmos-sdk/issues/4141) Fix /txs/encode endpoint

<!-- Release links -->

[Unreleased]: https://github.com/cosmos/gaia/compare/v7.0.3...HEAD
[v7.0.3]: https://github.com/cosmos/gaia/releases/tag/v7.0.3
[v7.0.2]: https://github.com/cosmos/gaia/releases/tag/v7.0.2
[v7.0.1]: https://github.com/cosmos/gaia/releases/tag/v7.0.1
[v7.0.0]: https://github.com/cosmos/gaia/releases/tag/v7.0.0
[v6.0.4]: https://github.com/cosmos/gaia/releases/tag/v6.0.4
[v6.0.3]: https://github.com/cosmos/gaia/releases/tag/v6.0.3
[v6.0.2]: https://github.com/cosmos/gaia/releases/tag/v6.0.2
[v6.0.1]: https://github.com/cosmos/gaia/releases/tag/v6.0.1
[v6.0.0]: https://github.com/cosmos/gaia/releases/tag/v6.0.0
[v5.0.8]: https://github.com/cosmos/gaia/releases/tag/v5.0.8
[v5.0.7]: https://github.com/cosmos/gaia/releases/tag/v5.0.7
[v5.0.6]: https://github.com/cosmos/gaia/releases/tag/v5.0.6
[v5.0.5]: https://github.com/cosmos/gaia/releases/tag/v5.0.5
[v5.0.4]: https://github.com/cosmos/gaia/releases/tag/v5.0.4
[v5.0.3]: https://github.com/cosmos/gaia/releases/tag/v5.0.3
[v5.0.2]: https://github.com/cosmos/gaia/releases/tag/v5.0.2
[v5.0.1]: https://github.com/cosmos/gaia/releases/tag/v5.0.1
[v5.0.0]: https://github.com/cosmos/gaia/releases/tag/v5.0.0
[v4.2.1]: https://github.com/cosmos/gaia/releases/tag/v4.2.1
[v4.2.0]: https://github.com/cosmos/gaia/releases/tag/v4.2.0
[v4.1.2]: https://github.com/cosmos/gaia/releases/tag/v4.1.2
[v4.1.1]: https://github.com/cosmos/gaia/releases/tag/v4.1.1
[v4.1.0]: https://github.com/cosmos/gaia/releases/tag/v4.1.0
[v4.0.6]: https://github.com/cosmos/gaia/releases/tag/v4.0.6
[v4.0.5]: https://github.com/cosmos/gaia/releases/tag/v4.0.5
[v4.0.4]: https://github.com/cosmos/gaia/releases/tag/v4.0.4
[v4.0.3]: https://github.com/cosmos/gaia/releases/tag/v4.0.3
[v4.0.2]: https://github.com/cosmos/gaia/releases/tag/v4.0.2
[v4.0.1]: https://github.com/cosmos/gaia/releases/tag/v4.0.1
[v4.0.0]: https://github.com/cosmos/gaia/releases/tag/v4.0.0
[v3.0.1]: https://github.com/cosmos/gaia/releases/tag/v3.0.1
[v3.0.0]: https://github.com/cosmos/gaia/releases/tag/v3.0.0
[v2.0.14]: https://github.com/cosmos/gaia/releases/tag/v2.0.14
[v2.0.13]: https://github.com/cosmos/gaia/releases/tag/v2.0.13
[v2.0.12]: https://github.com/cosmos/gaia/releases/tag/v2.0.12
[v2.0.11]: https://github.com/cosmos/gaia/releases/tag/v2.0.11
[v2.0.10]: https://github.com/cosmos/gaia/releases/tag/v2.0.10
[v2.0.9]: https://github.com/cosmos/gaia/releases/tag/v2.0.9
[v2.0.8]: https://github.com/cosmos/gaia/releases/tag/v2.0.8
[v2.0.7]: https://github.com/cosmos/gaia/releases/tag/v2.0.7
[v2.0.6]: https://github.com/cosmos/gaia/releases/tag/v2.0.6
[v2.0.5]: https://github.com/cosmos/gaia/releases/tag/v2.0.5
[v2.0.4]: https://github.com/cosmos/gaia/releases/tag/v2.0.4
[v2.0.3]: https://github.com/cosmos/gaia/releases/tag/v2.0.3
[v2.0.2]: https://github.com/cosmos/gaia/releases/tag/v2.0.2
[v2.0.1]: https://github.com/cosmos/gaia/releases/tag/v2.0.1
[v2.0.0]: https://github.com/cosmos/gaia/releases/tag/v2.0.0
[v1.0.0]: https://github.com/cosmos/gaia/releases/tag/v1.0.0
