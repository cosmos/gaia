
### Long-Lived Version Branch Approach

Cherry-pick commits from `main` into the long-lived `release/vn.n.x` branch, e.g., `release/v3.0.x`. 
It is fine to create a long-lived branch from main if the last commit is the release commit.

### Release Procedure

- Start on `main`
- Create the release candidate branch `rc/v*` (going forward known as **RC**)
  and ensure it's protected against pushing from anyone except the release
  manager/coordinator
    - **no PRs targeting this branch should be merged unless exceptional circumstances arise**
- On the `RC` branch, prepare a new version section in the `CHANGELOG.md` and
  kick off a large round of simulation testing (e.g. 400 seeds for 2k blocks)
- If errors are found during the simulation testing, commit the fixes to `main`
  and create a new `RC` branch (making sure to increment the `rcN`)
- After simulation has successfully completed, create the release branch
  (`release/vX.XX.X`) from the `RC` branch
- Merge the release branch to `main` to incorporate the `CHANGELOG.md` updates
- Delete the `RC` branches

### Point Release Procedure

At the moment, only a single major release will be supported, so all point
releases will be based off of that release.

- start on `vX.XX.X`
- checkout a new branch `rcN/vX.X.X`
- cherry pick the desired changes from `main`
    - these changes should be small and NON-BREAKING (both API and state machine)
- add entries to CHANGELOG.md and remove corresponding pending log entries
- checkout a new branch `release/vX.X.X` based off of the previous release
- create a PR merging `rcN/vX.X.X` into `release/vX.X.X`
- run tests and simulations (noted in [Release Procedure](#release-procedure))
- after tests and simulation have successfully completed, merge the `RC` branch into `release/vX.X.X`
    - Make sure to delete the `RC` branch
- create a PR into `main` containing ONLY the CHANGELOG.md updates
- tag (use `git tag -a`) then push the tags (`git push --tags`)

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

