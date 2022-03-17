# Releasing

This document outlines the release process for https://github.com/cosmos/gaia. We use a [Long-Lived Version Branch Approach](x) because we work in parallel on a `main` branch, a `release` branch and a `release-prepare` branch. The reason for this is because our software is used to run a single live network that at any given moment only works with one major version of our software. We may make a major release, but there will be a time span after that release before the software is live on the network, and we'd like our `main` branch to always be compilable and (theoretically) able to be run with the current live network (if it's in main but not a release it comes with the caveat that there may be dragons).

We follow [Semver](https://semver.org/) in that any patch releases are non-breaking changes. It's important to note, that breaking changes in a Blockchain context include non-determinism. So if a code change is backwards compatible, it may still impact the amount of gas needed to execute an action, which means the change is in fact breaking as it results in a different apphash after the code is executed. It's important for non-breaking changes to be possible to be used on the live network prior to the release.

Each major release will have a release branch and patch releases will be tagged on this branch. No patched release has its own branch. (This branch strategy only applies to `v7` and later releases.)

## Long-Lived Version Branch Approach

In the Gaia repo, there are three categories of long-lived branches:

### Branch: `main` 
The `main` branch should be targeted for PRs that contain a bug-fix/feature/improvement that will be included in a release that is used on the currently live Cosmos Hub Network or to be backported to a previous release. The PRs against `main` might only be get merged when the release process starts.

### Branch: `release`
There are multiple long lived branches with the `release/` prefix. Each release series follows the format `release/vn.0.x`, e.g. `release/v7.0.x`. The branch `release/vn.0.x` should point to the most updated `vn` release, e.g. `release/v5.0.x` should be the same as `release/v5.0.8` if that's the latest `v5.0` release. When starting a new minor or patch release, begin with `release/vn.0.x` and checkout a new branch called `rc0/vn.0.m` where `m` is +1 the latest release number. This branch will be used in a PR into `release/vn.0.m` and eventually merged back to `release/vn.0.x`. When making a major release, the process should be the same as minor and patch except begin with a name like `release/vn.0.0-rc0` which contains cherry-picked commits from `release-prepare` branch or merged from `release-prepare` branch. 
  
### Branch: `release-prepare`
The `release-prepare` branch named `[upgradename]-main` is initially created from the latest `release/vn.0.x` of the previous major release. It serves the same purpose as `main` except for the next planned Cosmos Hub Upgrade. As soon as the upgrade has taken place, it is merged into main and won't exist until the next upgrade is in progress. During the period that `main` is the currently running network version and a new major release is being prepared, `release-prepare` is targeted for PRs that contain bug-fixes/features/improvements for that release. There may be some PRs that target `main` because they are relevant to the current network, but are also relevant to the upcoming upgrade. These should be cherry picked into `release-prepare` before being subsequently cherry picked into the respective `release/vn.0.x` that is being used for the next upgrade. The `release-prepare` branch aims to be a clean starting point for the subsequent `release/` branch. Therefore, only after the binary built from `release-prepare` branch passed dev-testnet action, will this branch be used in a subsequent `release/` branch. `release/` branch contains clean a squashed commit from `release-prepare` and an updated changedlog. `release/` will be finally merged into `release/vn.0.x` and deleted. This means we do not keep each release a branch. 

The `release-prepare` branch will become `main` after the relevant Cosmos Hub Network Upgrade successfully takes place.

## Release Procedure

### Minor & Patch Releases

TODO

### Major Release

For a new major release, `m`, after a previous release, `n`, start on `release/vn.0.x` and checkout a new branch called `upgradename-main`(e.g. `Theta-main`). This will be the `release-prepare` branch referenced above going forward until the network upgrades with the new Major Release. This branch should act like `main` in that it is the target for new PRs that come in with code meant for the upcoming release. Some PRs that target the real `main` will be relevant to this `upgradename-main`, and should similarly be cherry picked to this branch. Before beginning the initial release candidate process, this branch should pass the `dev-testnet` tests. Once passing, checkout a new branch, `release/vm.0.x` off  `release/vn.0.x`. From `release/vm.0.x` checkout two branched: `release/vm.0.0-rc0` and `rc0/vm.0.0-rc0`. Rebase `upgradename-main` into `rc0/vm.0.0-rc0`. Now open a PR from `rc0/vm.0.0-rc0` against `release/vn.0.0-rc0`. This PR is important as it's the last time code reviewers can see all changes going into the release at once. Last minute changes can go directly into `rc0/vm.0.0-rc0` as well as updaing the CHANGELOG.md. Once the PR has been merged, the `rc0/vm.0.0-rc` branch can be deleted. The `release/vm.0.0-rc0` can be tagged and the release can be published as a pre-release.

This `release/vm.0.0-rc0` will be used on the public testnet before the upgrade takes place. If there are no issues, it can be re-tagged and re-released as `release/vm.0.0` and used in the final upgrade. Should changes be needed follow the Minor & Patch Release process iterating on the final `rcn` number.

After a successful release, ensure `release/vm.0.0` is merged back into `release/vm.0.x` and that is merged back into `upgradename-main`.

### Dependency review

Check the `replace` line in `go.mod` of the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/blob/master/go.mod) for something like:
```
replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
```
Ensure that the same replace line is also used in Gaia's `go.mod` file.

### Tagging

The following steps are the default for tagging a specific branch commit (usually on a branch labeled `release/vX.X.X`):
1. Ensure you have checked out the commit you wish to tag
1. `git pull --tags --dry-run`
1. `git pull --tags`
1. `git tag -a v3.0.1 -m 'Release v3.0.1'`
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
1. To finalize the release, push the local tags, create a release based off the newly pushed tag, and attach the binary. 

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

