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

* (build) [#2018](https://github.com/cosmos/gaia/pull/2018) Consistently generate build tags on Make 4.3+
* (docs) [#2053](https://github.com/cosmos/gaia/pull/2053) Added Shapeshift Wallet

## [v8.0.1] - 2023-02-17

* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.14](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.14). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.26](https://github.com/informalsystems/tendermint/tree/v0.34.26). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.26/CHANGELOG.md) for details.

## [v8.0.0] - 2023-01-31

* (gaia) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v3.4.0](https://github.com/cosmos/ibc-go/blob/v3.4.0/CHANGELOG.md) to fix a vulnerability in ICA. See [v3.4.0 CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.9/CHANGELOG.md) and [v3.2.1 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v3.2.1) for details.
* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.12](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.12). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.24](https://github.com/informalsystems/tendermint/tree/v0.34.24). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.24/CHANGELOG.md) for details.
* (gaia) Bump [liquidity](https://github.com/Gravity-Devs/liquidity) to [v1.5.3](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.3).
* (gaia) Bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v3.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v3.1.1).
* (feat) Add [globalfee](https://github.com/cosmos/gaia/tree/main/x/globalfee) module. See [globalfee docs](https://github.com/cosmos/gaia/blob/main/docs/modules/globalfee.md) for more details.
* (feat) [#1845](https://github.com/cosmos/gaia/pull/1845) Add bech32-convert command to gaiad.
* (fix) [#2080](https://github.com/cosmos/gaia/issues/2074) Reintroduce deleted configuration for client rpc endpoints, transaction routes, and module REST routes in app.go.
* (fix) [Add new fee decorator](https://github.com/cosmos/gaia/pull/1961) to change `MaxBypassMinFeeMsgGasUsage` so importers of x/globalfee can change `MaxGas`.
* (fix) [#1870](https://github.com/cosmos/gaia/issues/1870) Fix bank denom metadata in migration. See [#1892](https://github.com/cosmos/gaia/pull/1892) for more details.
* (fix) [#1976](https://github.com/cosmos/gaia/pull/1976) Fix Quicksilver ICA exploit in migration. See [the bug fix forum post](https://forum.cosmos.network/t/upcoming-interchain-accounts-bugfix-release/8911) for more details.
* (tests) Add [E2E tests](https://github.com/cosmos/gaia/tree/main/tests/e2e). The tests cover transactions/queries tests of different modules, including Bank, Distribution, Encode, Evidence, FeeGrant, Global Fee, Gov, IBC, packet forwarding middleware, Slashing, Staking, and Vesting module.
* (tests) [#1941](https://github.com/cosmos/gaia/pull/1941) Fix packet forward configuration for e2e tests.
* (tests) Use gaiad to swap out [Ignite](https://github.com/ignite/cli) in [liveness tests](https://github.com/cosmos/gaia/blob/main/.github/workflows/test.yml).

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

<!-- Release links -->

[Unreleased]: https://github.com/cosmos/gaia/compare/v8.0.1...release/v8.0.x
[v8.0.1]: https://github.com/cosmos/gaia/releases/tag/v8.0.1
[v8.0.0]: https://github.com/cosmos/gaia/releases/tag/v8.0.0
