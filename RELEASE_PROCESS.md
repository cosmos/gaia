# Release Process

- [Release Process](#release-process)
  - [Major Release Procedure](#major-release-procedure)
    - [Tagging Procedure](#tagging-procedure)
  - [Patch Release Procedure](#patch-release-procedure)
  - [Major Release Maintenance](#major-release-maintenance)
  - [Stable Release Policy](#stable-release-policy)
- [Old Release Process](#old-release-process)
  - [Release Procedure](#release-procedure)
    - [Checks and tests](#checks-and-tests)
    - [Major and minor Release](#major-and-minor-release)
      - [example of releasing `v8.0.0-rc0`](#example-of-releasing-v800-rc0)
      - [example of releasing `v8.0.0`](#example-of-releasing-v800)
      - [example of releasing `v8.0.1`](#example-of-releasing-v801)
    - [backport release](#backport-release)
      - [example of backport release `v7.0.5`](#example-of-backport-release-v705)
    - [Test building artifacts](#test-building-artifacts)
    - [Release notes](#release-notes)


This document outlines the release process for Cosmos Hub (Gaia).

Gaia follows [semantic versioning](https://semver.org), but with the following deviations to account for state-machine breaking changes: 
- Changes that requires upgrade trough governance will result in an increase of the major version X (X.y.z). Note that most likely this entails state-machine breaking changes.
- Changes that are state-machine breaking but do not require upgrade through governance (e.g., emergency upgrades) will result in an increase of the minor version Y (x.Y.z | x > 0).
- Changes that are not state-machine breaking will result in an increase of the patch version Z (x.x.Z | x > 0).

**Note**: State-machine breaking changes include changes that impact the amount of gas needed to execute a transaction as it results in a different `apphash` after the code is executed.

Every major release will have a release branch and patch releases will be tagged on this branch. No patch releases have their own branch. (This branch strategy only applies to `v7` and later releases.)

## Major Release Procedure

A _major release_ is an increment of the first number (eg: `v7.1.0` → `v8.0.0`) or the _point number_ (eg: `v7.0.0 → v7.1.0`, also called _point release_). Each major release opens a _stable release series_ and receives updates outlined in the [Major Release Maintenance](#major-release-maintenance) section.

> Note: Generally, PRs should target `main` (expect PRs open via the Github mergify integration). 

* Once the team feels that `main` is feature complete, we create a `release/vY` branch (going forward known a release branch), 
  where `Y` is the version number, with the patch part substituted to `x` (eg: 8.0.x). 
  * Update the [GitHub mergify integration](./.mergify.yml) by adding instructions for automatically backporting commits from `main` to the `release/vY` using the `A:backport/vY` label.
  * **PRs targeting directly a release branch can be merged _only_ when exceptional circumstances arise**.
* In the release branch prepare a new version section in the `CHANGELOG.md`
    * All links must point to their respective pull request.
    * The `CHANGELOG.md` must contain only the changes of that specific released version. 
      All other changelog entries must be deleted and linked to the `main` branch changelog ([example]([TBA](https://github.com/cosmos/gaia/blob/release/v9.0.x/CHANGELOG.md))).
    * Create release notes, in `RELEASE_NOTES.md`, highlighting the new features and changes in the version. 
      This is needed so the bot knows which entries to add to the release page on GitHub.
    * Additionally verify that the `UPGRADING.md` file is up to date and contains all the necessary information for upgrading to the new version.
* We freeze the release branch from receiving any new features and focus on releasing a release candidate.
  * Finish audits and reviews.
  * Add more tests.
  * Fix bugs as they are discovered.
* After the team feels that the release branch works fine (i.e., has `~90%` chance of reaching mainnet), we cut a release candidate.
  * Create a new annotated git tag for a release candidate in the release branch (follow the [Tagging Procedure](#tagging-procedure)).
  * The release verification on public testnets must pass. 
  * When bugs are found, create a PR for `main`, and backport fixes to the release branch.
  * Create new release candidate tags after bugs are fixed.
* After the team feels the release candidate is mainnet ready, create a full release:
  * Update `CHANGELOG.md`.
  * Run `gofumpt -w -l .` to format the code.
  * Create a new annotated git tag in the release branch (follow the [Tagging Procedure](#tagging-procedure)).
  * Create a GitHub release.

### Tagging Procedure

The following steps are the default for creating a new annotated git tag in a release branch

1. Ensure you have checked out the commit you wish to tag
2. `git pull --tags --dry-run`
3. `git pull --tags`
4. `git tag -a -s v9.0.1 -m 'Release v9.0.1'`
   - optional, add the `-s` tag to create a signed commit using your PGP key (which should be added to github beforehand)
5. `git push --tags --dry-run`
6. `git push --tags`

To re-create a tag:

1. `git tag -d v9.0.1` to delete a tag locally
2. `git push --delete origin v9.0.1`, to push the deletion to the remote
3. Proceed with the above steps to create a tag

To tag and build without a public release (e.g., as part of a timed security release):

1. Follow the steps above for tagging locally, but do not push the tags to the repository.
2. After adding the tag locally, you can build the binary, e.g., `make build-reproducible`.
3. To finalize the release, push the local tags, create a release based off the newly pushed tag, and attach the binaries.

## Patch Release Procedure

A _patch release_ is an increment of the patch number (eg: `v8.0.0` → `v8.0.1`).

**Patch release must not break consensus.**

Updates to the release branch should come from `main` by backporting PRs 
(usually done by automatic cherry pick followed by a PRs to the release branch). 
The backports must be marked using `backport/Y` label in PR for main.
It is the PR author's responsibility to fix merge conflicts, update changelog entries, and
ensure CI passes. If a PR originates from an external contributor, a core team member assumes
responsibility to perform this process instead of the original author.
Lastly, it is core team's responsibility to ensure that the PR meets all the Stable Release Update (SRU) criteria.

Point Release must follow the [Stable Release Policy](#stable-release-policy).

After the release branch has all commits required for the next patch release:

* Update `CHANGELOG.md` and `RELEASE_NOTES.md` (if applicable).
* Create a new annotated git tag in the release branch (follow the [Tagging Procedure](#tagging-procedure)).
* Create a GitHub release (if applicable).

## Major Release Maintenance

Major Release series continue to receive bug fixes (released as a Patch Release) until they reach **End Of Life**.
Major Release series is maintained in compliance with the **Stable Release Policy** as described in this document.
Note: not every Major Release is denoted as stable releases.

After two major releases, a supported major release will be transitioned to unsupported and will be deemed EOL with no further updates.
For example, `release/v7.1.x` is deemed EOL once the network upgrades to `release/v9.0.x`. 

## Stable Release Policy

Once a Gaia release has been completed and published, updates for it are released under certain circumstances
and must follow the [Patch Release Procedure](#patch-release-procedure).

The intention of the Stable Release Policy is to ensure that all major release series that are not EOL, 
are maintained with the following categories of fixes:

- Tooling improvements (including code formatting, linting, static analysis and updates to testing frameworks)
- Performance enhancements for running archival and synching nodes
- Test and benchmarking suites, ensuring that fixes are sound and there are no performance regressions
- Library updates including point releases for core libraries such as IBC-Go, Cosmos SDK, Tendermint and other dependencies
- General maintenance improvements, that are deemed necessary by the stewarding team, that help align different releases and reduce the workload on the stewarding team
- Security fixes

Issues that are likely excluded, are any issues that impact operating a block producing network.

---------

# Old Release Process

## Release Procedure

### Checks and tests

Before merge and release, the following tests checks need to be conducted:

- check the `replace` line in `go.mod`, check all the versions in `go.mod` are correct.
- run tests and simulations by `make run-tests`.
- test version compatibilities for minor releases.

### Major and minor Release

For a new major release `n`, checkout `release/vn.0.x` from `main`. Merge or use mergify to merge the commits to `release/vn.0.x`, and tag the version.
For minor release. Merge or use mergify to merge the commits to `release/vn.0.x`, and tag the version.

Usually the first release on the `release/vn.0.x` is a release candidate.

#### example of releasing `v8.0.0-rc0`

1. checkout `release/v8.0.x` off `main`
1. get the `v8-prepare-branch` ready including CHANGELOG.md, create a PR to merge `v8-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v8-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.0-rc0`.

#### example of releasing `v8.0.0`

1. get the `v800-prepare-branch` ready including CHANGELOG.md, create a PR to merge `v800-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v800-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.0`.

#### example of releasing `v8.0.1`

1. get the `v801-prepare-branch`(off `main`) ready including CHANGELOG.md, create a PR to merge `v801-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v801-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.1`.

### backport release

For a backport release, checkout a new branch from the right release branch, for example, `release/vn-1.0.x`. Commits to this new branch and merge into `release/vn-1.0.x`, tag the backport version from `release/vn-1.0.x`.

#### example of backport release `v7.0.5`

assume main branch is at `v8`.

1. checkout `v705-prepare-branch` off `release/v7.0.x`, get the backport changes ready including CHANGELOG.md on `v705-prepare-branch`.
1. create a PR to merge `v705-prepare-branch` to `release/v7.0.x`, and merge.
1. checkout `release/v7.0.x`  tag `v7.0.5`.

### Test building artifacts

Before tagging the version, please test the building releasing artifacts by

```bash
make distclean build-reproducible
```

The above command will generate a directory
`gaia/artifacts` with different os and architecture binaries. If the above command runs sucessfully, delete the directory `rm -r gaia/artifacts`.



### Release notes

Ensure you run the reproducible build in order to generate sha256 hashes and platform binaries;
these artifacts should be included in the release.

```bash
make distclean build-reproducible
```

This runs the docker image [tendermintdev/rbuilder](https://hub.docker.com/r/tendermintdev/rbuilder) with a copy of the [rbuilder](https://github.com/tendermint/images/tree/master/rbuilder) docker file.

Then use the following release text template:

```markdown
# Gaia v4.0.0 Release Notes

Note, that this specific release will be updated with a newer changelog, and the below hashes and binaries will also be updated.

This release includes bug fixes for prop29, as well as additional support for IBC and Ledger signing.

As there is a breaking change from Gaia v3, the Gaia module has been incremented to v4.

See the [Cosmos SDK v0.41.0 Release](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.0) for details.

```bash
$ make distclean build-reproducible
App: gaiad
Version: 4.0.0
Commit: 2bb04266266586468271c4ab322367acbf41188f
Files:
 2e801c7424ef67e6d9fc092c2b75c2d3  gaiad-4.0.0-darwin-amd64
 dc21eb861480e0f55af876f271b512fe  gaiad-4.0.0-linux-amd64
 bda165f91bc065afb8a445e72be9a868  gaiad-4.0.0-linux-arm64
 c7203d53bd596679b39b6a94d1dbe0dc  gaiad-4.0.0-windows-amd64.exe
 81299b602e1760078e03c97813edda60  gaiad-4.0.0.tar.gz
Checksums-Sha256:
 de764e52acc31dd98fa49d8d0eef851f3b7b53e4f1d4fbfda2c07b1a8b115b91  gaiad-4.0.0-darwin-amd64
 e5244ccd98a05479cc35753da1bb5b6bd873f6d8ebe6f8c5d112cf4d9e2761b4  gaiad-4.0.0-linux-amd64
 7b7c4ea3e2de5f228436dcbb177455906239603b11eca1fb1015f33973d7b567  gaiad-4.0.0-linux-arm64
 b418c5f296ee6f946f44da8497af594c6ad0ece2b1da09a93a45d7d1b1457f27  gaiad-4.0.0-windows-amd64.exe
 3895518436b74be8b042d7d0b868a60d03e1656e2556b12132be0f25bcb061ef  gaiad-4.0.0.tar.gz
```



