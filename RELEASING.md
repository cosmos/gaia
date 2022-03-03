
### Long-Lived Version Branch Approach

In Gaia repo, there are three categories of long-lived branches:
##`main` 
`main` allows PR merge if that PR is targeting a bug/feature/improvement that will be included in a backport release or present release. In order to keep `main` always pointing to the most updated version that  can be used on live cosmoshub net. The PRs against `main` might only be get merged when the release process starts.
## `release` branch
Each new release line start with `release/vn.0.x`, e.g. `release/v7.0.x`, `release/vn.0.x` should point to the most updated `vn` releases. Each release will checkout its own branch but merged back to  `release/vn.0.x`. The first release on `release/vn.0.x` should be `release/vn.0.0-rc0` which contains cherry-picked commits from `release-prepare` branch or merged from `release-prepare` branch. 
  
## `release-prepare` branch
`release-prepare` branch will cherry-pick the commits from main or be directly commits to. `release-prepare` aims at a clean `release-branch`. Therefore, only till the binary built from `release-prepare` branch passes the dev-testnet, that this branch can be cherry-pick/merge to `release` branch. 

The `release` branch will merge back to `main` when the live cosmoshub net upgrades to this release version line.

### Release Procedure

 start on `release/v(n-1).0.x`, checkout `upgradename-main`(e.g. `theta-main`) branch, cherry-pick/merge commits from main, or commits directly to `upgradename-main` branch. When  `upgradename-main` branch pass `dev-testnet` tests, checkout a new branch  `release/vn.0.x` off  `release/v(n-1).0.x`, checkout  `release/vn.0.0-rc0` from  `release/vn.0.x`, merge commits from `upgradename-main` branch to `release/vn.0.0-rc0`, modify change log, tag `release/vn.0.0-rc0`.






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

