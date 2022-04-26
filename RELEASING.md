# Releasing

This document outlines the release process for https://github.com/cosmos/gaia. We use [Long-Lived Version Branch Approach](x) on a `main` branch and a `release` branch.

We follow [Semver](https://semver.org/) in that any patch releases are non-breaking changes. It's important to note, that breaking changes in a Blockchain context include non-determinism. So if a code change is backwards compatible, it may still impact the amount of gas needed to execute an action, which means the change is in fact breaking as it results in a different apphash after the code is executed. It's important for non-breaking changes to be possible to be used on the live network prior to the release.

Each major release will have a release branch and patch releases will be tagged on this branch. No patched release has its own branch. (This branch strategy only applies to `v7` and later releases.)

## Long-Lived Version Branch Approach

In the Gaia repo, there are two categories of long-lived branches:

### Branch: `main`
The `main` branch should be targeted for PRs that contain a bug-fix/feature/improvement that will be included for the next release. 

### Branch: `release`
There are multiple long-lived branches with the `release/` prefix. Each release series follows the format `release/vn.0.x`, e.g. `release/v7.0.x`. The branch `release/vn.0.x` should point to the most updated `vn` release, e.g. `release/v5.0.x` should be the same as `release/v5.0.8` if that's the latest `v5.0` release.

## Other Branches
### branches for the next release:

Other feature/fix branches targeting at `main` contain commits preparing for the next release. When the `release-prepare-branch` is ready for next release, add label `A:backport/vn.0.x` to the PR of `release-prepare-branch` against `main`, then the mergifybot will create a new PR of `mergify/bp/release/vn.0.x`  against `Release/vn.0.x`.

### branches for the backport release:

If the feature/fix branches are for a backport release, `main` branch already contains the commits for the next major release  `vn`, the feature/fix branch's PR should target at `Release/vn-1` rather than `main`. 

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

#### example of releasing `v8.0.0-rc0`:

1. checkout `release/v8.0.x` off `main`
1. get the `v8-prepare-branch` ready including CHANGELOG.md, create a PR to merge `v8-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v8-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.0-rc0`.

#### example of releasing `v8.0.0`:

1. get the `v800-prepare-branch` ready including CHANGELOG.md, create a PR to merge `v800-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v800-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.0`.

#### example of releasing `v8.0.1`:

1. get the `v801-prepare-branch`(off `main`) ready including CHANGELOG.md, create a PR to merge `v801-prepare-branch` to `main`, label this PR `A:backport/v8.0.x`.
1. after merge  `v801-prepare-branch` to `main`, mergifybot will create a new PR of  `mergify/bp/release/v8.0.x` to `release/v8.0.x`. Check the PR, and merge this PR.
1. checkout  `release/v8.0.x` and tag `v8.0.1`.

### backport release
For a backport release, checkout a new branch from the right release branch, for example, `release/vn-1.0.x`. Commits to this new branch and merge into `release/vn-1.0.x`, tag the backport version from `release/vn-1.0.x`.

#### example of backport release `v7.0.5`:
assume main branch is at `v8`.
1. checkout `v705-prepare-branch` off `release/v7.0.x`, get the backport changes ready including CHANGELOG.md on `v705-prepare-branch`.
1. create a PR to merge `v705-prepare-branch` to `release/v7.0.x`, and merge.
1. checkout `release/v7.0.x`  tag `v7.0.5`.

### Tagging

The following steps are the default for tagging a specific branch commit (usually on a branch labeled `release/vX.X.X`):
1. Ensure you have checked out the commit you wish to tag
1. `git pull --tags --dry-run`
1. `git pull --tags`
1. `git tag -s v3.0.1 -m 'Release v3.0.1'`
   1. optional, add the `-s` tag to create a signed commit using your PGP key (which should be added to github beforehand)
1. `git push --tags --dry-run`
1. `git push --tags`

To re-create a tag:
1. `git tag -d v4.0.0` to delete a tag locally
1. `git push --delete origin v4.0.0`, to push the deletion to the remote
1. Proceed with the above steps to create a tag

To tag and build without a public release (e.g., as part of a timed security release):
1. Follow the steps above for tagging locally, but do not push the tags to the repository.
1. After adding the tag locally, you can build the binary, e.g., `make build-reproducible`.
1. To finalize the release, push the local tags, create a release based off the newly pushed tag, and attach the binaries.

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

