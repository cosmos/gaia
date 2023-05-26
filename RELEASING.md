# Releasing

This document outlines the release process for <https://github.com/cosmos/gaia>. We use [Long-Lived Version Branch Approach](x) on a `main` branch and a `release` branch.

We follow [Semver](https://semver.org/) in that any patch releases are non-breaking changes. It's important to note, that breaking changes in a Blockchain context include non-determinism. So if a code change is backwards compatible, it may still impact the amount of gas needed to execute an action, which means the change is in fact breaking as it results in a different apphash after the code is executed. It's important for non-breaking changes to be possible to be used on the live network prior to the release.

Each major release will have a release branch and patch releases will be tagged on this branch. No patched release has its own branch. (This branch strategy only applies to `v7` and later releases.)

## Long-Lived Version Branch Approach

In the Gaia repo, there are two categories of long-lived branches:

### Branch: `main`

The `main` branch should be targeted for PRs that contain a bug-fix/feature/improvement that will be included for the next release.

### Branch: `release`

There are multiple long-lived branches with the `release/` prefix. Each release series follows the format `release/vn.0.x`, e.g. `release/v7.0.x`. The branch `release/vn.0.x` should point to the most updated `vn` release, e.g. `release/v5.0.x` should be the same as `release/v5.0.8` if that's the latest `v5.0` release.

## Other Branches

### branches for the next release

Other feature/fix branches targeting at `main` contain commits preparing for the next release. When the `release-prepare-branch` is ready for next release, add label `A:backport/vn.0.x` to the PR of `release-prepare-branch` against `main`, then the mergifybot will create a new PR of `mergify/bp/release/vn.0.x`  against `Release/vn.0.x`.

### branches for the backport release

If the feature/fix branches are for a backport release, `main` branch already contains the commits for the next major release  `vn`, the feature/fix branch's PR should target at `Release/vn-1` rather than `main`.

## Release Procedure
### Concept

The release procedure always contains these steps:
* prepare release branch and add `CHANGELOG.md` with relevant changes
* create a tag (using git or Github UI)
* wait for the automated release process to complete and upload artifacts
* modify release notes if needed

### Checks and tests

Before merge and release, the following tests checks need to be conducted:

- check the `replace` line in `go.mod`, check all the versions in `go.mod` are correct.
- run tests and simulations by `make run-tests`.
- test version compatibilities for minor releases.

### Major and minor Release

For a new major release `n`, checkout `release/vn.0.x` from `main`. Merge or use mergify to merge the commits to `release/vn.0.x`, and tag the version.
For minor release. Merge or use mergify to merge the commits to `release/vn.0.x`, and tag the version.

Usually the first release on the `release/vn.0.x` is a release candidate.

You can create releases using `git` CLI or using Github UI. When using Github UI, you will end up with 2 releases due to how release workflow is setup - simply delete the release that you do not need.


#### example of releasing `v8.0.0-rc0`

1. checkout `release/v8.0.x` off `main`
2. create a PR against `main` with changes CHANGELOG.md if the CHANGELOG.md is not up-to-date on `release/v8.0.x`, label this PR `A:backport/v8.0.x`.
3. mergifybot will create a new PR targeting `release/v8.0.x` with changes from 2. Check the PR, and merge it.
4. checkout `release/v8.0.x` branch locally and tag it as `v8.0.0-rc0`.

#### example of releasing `v8.0.0`

1. create a PR against `main` with changes CHANGELOG.md if the CHANGELOG.md is not up-to-date on `release/v8.0.x`, label this PR `A:backport/v8.0.x`.
2. mergifybot will create a new PR targeting `release/v8.0.x` with changes from 1. Check the PR, and merge it.
3. checkout `release/v8.0.x` branch locally and tag it as `v8.0.0`.

#### example of releasing `v8.0.1`

1. create a PR against `main` with changes CHANGELOG.md if the CHANGELOG.md is not up-to-date on `release/v8.0.x`, label this PR `A:backport/v8.0.x`.
2. mergifybot will create a new PR targeting `release/v8.0.x` with changes from 1. Check the PR, and merge it.
3. checkout `release/v8.0.x` branch locally and tag it as `v8.0.0`.

### backport release

For a backport release, checkout a new branch from the right release branch, for example, `release/vn-1.0.x`. Commits to this new branch and merge into `release/vn-1.0.x`, tag the backport version from `release/vn-1.0.x`.
Create the tag the same way as for regular releases.

### Test building artifacts

Before tagging the version, please test the building releasing artifacts by running:

```bash
TM_VERSION=$(go list -m github.com/tendermint/tendermint | sed 's:.* ::') goreleaser release --snapshot --clean --debug
```

This step requires go-releaser. To install it follow [this link](https://goreleaser.com/install/).

### Tagging

The following steps are the default for tagging a specific branch commit using git on your local machine. Usually branches are labeled `release/vX.X.X`:

Ensure you have checked out the commit you wish to tag and then do:
```bash
git pull --tags --dry-run
git pull --tags
# -s creates a signed commit using your PGP key (which should be added to github beforehand)
git tag -s v3.0.1 -m 'Release v3.0.1'
git push --tags --dry-run
git push --tags
```

To re-create a tag:
```bash
git tag -d v4.0.0  # delete a tag locally
git push --delete origin v4.0.0 # push the deletion to the remote
```

Proceed with the above steps to create a tag

### Release notes

Release notes will be created using the `CHANGELOG.md` from the `release/v*` branch. Feel free to add any missing information the the release notes using Github UI.

With every release the `goreleaser` tool will create a file with all the build artifact checksums and upload it alongside the artifacts.
The file is called `SHA256SUMS-{{.version}}.txt` and contains the following:
```
098b00ed78ca01456c388d7f1f22d09a93927d7a234429681071b45d94730a05  gaiad_0.0.4_windows_arm64.exe
15b2b9146d99426a64c19d219234cd0fa725589c7dc84e9d4dc4d531ccc58bec  gaiad_0.0.4_darwin_amd64
604912ee7800055b0a1ac36ed31021d2161d7404cea8db8776287eb512cd67a9  gaiad_0.0.4_darwin_arm64
76e5ff7751d66807ee85bc5301484d0f0bcc5c90582d4ba1692acefc189392be  gaiad_0.0.4_linux_arm64
bcbca82da2cb2387ad6d24c1f6401b229a9b4752156573327250d37e5cc9bb1c  gaiad_0.0.4_windows_amd64.exe
f39552cbfcfb2b06f1bd66fd324af54ac9ee06625cfa652b71eba1869efe8670  gaiad_0.0.4_linux_amd64
```

# Major Release Maintenance

Major Release series continue to receive bug fixes (released as a Patch Release) until they reach End Of Life.
Major Release series are maintained in compliance with the Stable Release Policy as described in this document. Note: not every Major Release is denoted as a stable release.

After two releases, a supported version will be transitioned to unsupported and will be deemed EOL with no further updates.

### Example
```
v9 latest, stable
v8 supported
v7 EOL, not supported
v6 EOL, not supported
```

# Stable Release Policy

The intention of the Stable Release Policy is to ensure that all major release series that are not EOL, are maintained with the following categories of fixes:

- Tooling improvements (including code formatting, linting, static analysis and updates to testing frameworks)
- Performance enhancements for running archival and synching nodes
- Test and benchmarking suites, ensuring that fixes are sound and there are no performance regressions
- Library updates including point releases for core libraries such as IBC-Go, Cosmos SDK, Tendermint and other dependencies
- General maintenance improvements, that are deemed necessary by the stewarding team, that help align different releases and reduce the workload on the stewarding team
- Security fixes

Issues that are likely excluded, are any issues that impact operating a block producing network.
