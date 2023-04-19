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

## [v9.0.3] - 2023-04-19
* (deps) [#2399](https://github.com/cosmos/gaia/pull/2399) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.15-ics](https://github.com/cosmos/cosmos
sdk/releases/tag/v0.45.15-ics) and migrate to [CometBFT](https://github.com/cometbft/cometbft).

## [v9.0.2] - 2023-04-03
* (feat) Bump [Interchain-Security](https://github.com/cosmos/interchain-security) [v1.1.0](https://github.com/cosmos/interchain-security/releases/tag/v1.1.0) provider module. See the [release notes](https://github.com/cosmos/interchain-security/releases/tag/v1.1.0) for details.
* (feat) Add two more msg types `/ibc.core.channel.v1.MsgTimeout` and `/ibc.core.channel.v1.MsgTimeoutOnClose` to default `bypass-min-fee-msg-types`.
* (feat) Change the bypassing gas usage criteria. Instead of requiring 200,000 gas per `bypass-min-fee-msg`, we will now allow a maximum total usage of 1,000,000 gas for all bypassed messages in a transaction. Note that all messages in the transaction must be the `bypass-min-fee-msg-types` for the bypass min fee to take effect, otherwise, fee payment will still apply.
* (fix) [#2087](https://github.com/cosmos/gaia/issues/2087) Fix `bypass-min-fee-msg-types` parsing in `app.toml`. Parsing of `bypass-min-fee-types` is changed to allow node operators to use empty bypass list. Removing the `bypass-min-fee-types` from `app.toml` applies the default message types. See [#2092](https://github.com/cosmos/gaia/pull/2092) for details.

 ## [v9.0.1] - 2023-03-09

 * (feat) [add spam prevention antehandler](https://github.com/cosmos/gaia/pull/2262) to alleviate recent governance spam issues.

## [v9.0.0] - 2023-02-21

* (feat) Add [Interchain-Security](https://github.com/cosmos/interchain-security) [v1.0.0](https://github.com/cosmos/interchain-security/releases/tag/v1.0.0) provider module. See the [ICS Spec](https://github.com/cosmos/ibc/blob/main/spec/app/ics-028-cross-chain-validation/README.md) for more details.
* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.13-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.13-ics). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/releases/tag/v0.45.13-ics) for details.
* (gaia) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.2.0](https://github.com/cosmos/ibc-go/blob/release/v4.2.x/CHANGELOG.md). See [v4.2 Release Notes](https://github.com/cosmos/ibc-go/releases/tag/v4.2.0) for details.
* (gaia) Bump [tendermint](https://github.com/informalsystems/tendermint) to [0.34.26](https://github.com/informalsystems/tendermint/tree/v0.34.26). See [CHANGELOG.md](https://github.com/informalsystems/tendermint/blob/v0.34.26/CHANGELOG.md#v03426) for details.
* (gaia) Bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v4.0.4](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v4.0.4).
* (tests) Add [E2E ccv tests](https://github.com/cosmos/gaia/blob/main/tests/e2e/e2e_gov_test.go#L138). Tests covering new functionality introduced by the provider module to add and remove a consumer chain via governance proposal.
* (tests) Add [integration ccv tests](https://github.com/cosmos/gaia/blob/main/tests/ics/interchain_security_test.go). Imports Interchain-Security's `TestCCVTestSuite` and implements Gaia as the provider chain.
* (fix) [#2017](https://github.com/cosmos/gaia/issues/2017) Fix Gaiad binary build tag for ubuntu system. See [#2018](https://github.com/cosmos/gaia/pull/2018) for details.

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

<!-- Release links -->
[Unreleased]: https://github.com/cosmos/gaia/compare/v9.0.3...release/v9.0.x
[v9.0.3]: https://github.com/cosmos/gaia/releases/tag/v9.0.3
[v9.0.2]: https://github.com/cosmos/gaia/releases/tag/v9.0.2
[v9.0.1]: https://github.com/cosmos/gaia/releases/tag/v9.0.1
[v9.0.0]: https://github.com/cosmos/gaia/releases/tag/v9.0.0