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

<<<<<<< HEAD
=======
* (deps) [#2543](https://github.com/cosmos/gaia/pull/2543) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.1](https://github.com/cosmos/ibc-go/releases/tag/v4.4.1).

### Improvements
* (test) [#2440](https://github.com/cosmos/gaia/pull/2440) Add vulncheck to nightly builds
* (gaia) [#2442](https://github.com/cosmos/gaia/pull/2442) Bump [Interchain-Security](https://github.com/cosmos/interchain-security) to [v1.1.1](https://github.com/cosmos/interchain-security/tree/v1.1.1).

### State Machine Breaking

* (feat!) [#2424](https://github.com/cosmos/gaia/pull/2424) Add `bypass-min-fee-msg-types` and `maxTotalBypassMinFeeMsgGagUsage` to globalfee params. Note that this change is both state breaking and API breaking. The previous API endpoint was "/gaia/globalfee/v1beta1/minimum_gas_prices," and the new API endpoint is "/gaia/globalfee/v1beta1/params."
* (feat!) [#2352](https://github.com/cosmos/gaia/pull/2352) Create the upgrade handler and params migration for the new Gloabal Fee module parameters introduced in [#2424](https://github.com/cosmos/gaia/pull/2424).
Update the CI upgrade tests from v9 to the v10 and check that the parameters are successfully migrated.
* (feat!) [#2447](https://github.com/cosmos/gaia/pull/2447) Update Global Fee's AnteHandler to check tx fees against the network min gas prices in DeliverTx mode.

>>>>>>> eb6980e (deps: bump IBC to v4.4.1 (#2543))
## [v10.0.0] 2023-05-19

* (deps) [#2498](https://github.com/cosmos/gaia/pull/2498) Bump multiple dependencies. 
  * Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.16-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16-ics). See the [v0.45.16 release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16) for details. 
  * Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.0](https://github.com/cosmos/ibc-go/releases/tag/v4.4.0).
  * Bump [CometBFT](https://github.com/cometbft/cometbft) to [v0.34.28](https://github.com/cometbft/cometbft/releases/tag/v0.34.28).
* (gaia) Bump Golang prerequisite from 1.18 to 1.20. See (https://go.dev/blog/go1.20) for details.

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

<!-- Release links -->
[Unreleased]: https://github.com/cosmos/gaia/compare/v10.0.0...release/v10.0.x
[v10.0.0]: https://github.com/cosmos/gaia/releases/tag/v10.0.0
