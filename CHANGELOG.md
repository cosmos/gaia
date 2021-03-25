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

[Unreleased]: https://github.com/cosmos/gaia/compare/v4.2.0...HEAD
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
