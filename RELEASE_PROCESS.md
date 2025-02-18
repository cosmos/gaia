# Release Process

- [Release Process](#release-process)
    - [Breaking Changes](#breaking-changes)
  - [Major Release Procedure](#major-release-procedure)
    - [Changelog](#changelog)
      - [Creating a new release branch](#creating-a-new-release-branch)
      - [Cutting a new release](#cutting-a-new-release)
      - [Update the changelog on main](#update-the-changelog-on-main)
    - [Release Notes](#release-notes)
    - [Tagging Procedure](#tagging-procedure)
      - [Test building artifacts](#test-building-artifacts)
      - [Installing goreleaser](#installing-goreleaser)
  - [Non-major Release Procedure](#non-major-release-procedure)
  - [Major Release Maintenance](#major-release-maintenance)
  - [Stable Release Policy](#stable-release-policy)


This document outlines the release process for Cosmos Hub (Gaia).

Gaia follows [semantic versioning](https://semver.org), but with the following deviations to account for state-machine and API breaking changes: 

- State-machine breaking changes will result in an increase of the major version X (X.y.z).
- Emergency releases & API breaking changes will result in an increase of the minor version Y (x.Y.z | x > 0).
- All other changes will result in an increase of the patch version Z (x.y.Z | x > 0).

**Note:** In case a major release is deprecated before ending up on the network (due to potential bugs), 
it is replaced by a minor release (eg: `v14.0.0` → `v14.1.0`). 
As a result, this minor release is considered state-machine breaking.

### Breaking Changes

A change is considered to be ***state-machine breaking*** if it requires a coordinated upgrade for the network to preserve [state compatibility](./STATE-COMPATIBILITY.md). 
Note that when bumping the dependencies of [Cosmos SDK](https://github.com/cosmos/cosmos-sdk), [IBC](https://github.com/cosmos/ibc-go), and [ICS](https://github.com/cosmos/interchain-security) we will only treat patch releases as non state-machine breaking.

A change is considered to be ***API breaking*** if it modifies the provided API. This includes events, queries, CLI interfaces. 

## Major Release Procedure

A _major release_ is an increment of the first number (eg: `v9.1.0` → `v10.0.0`). Each major release opens a _stable release series_ and receives updates outlined in the [Major Release Maintenance](#major-release-maintenance) section.

**Note**: Generally, PRs should target either `main` or a long-lived feature branch (see [CONTRIBUTING.md](./CONTRIBUTING.md#pull-requests)).
An exception are PRs open via the Github mergify integration (i.e., backported PRs). 

* Once the team feels that `main` is _**feature complete**_, we create a `release/vY` branch (going forward known as release branch), 
  where `Y` is the version number, with the minor and patch part substituted to `x` (eg: 11.x). 
  * Update the [GitHub mergify integration](./.mergify.yml) by adding instructions for automatically backporting commits from `main` to the `release/vY` using the `A:backport/vY` label.
  * **PRs targeting directly a release branch can be merged _only_ when exceptional circumstances arise**.
* In the release branch 
  * Create a new version section in the `CHANGELOG.md` (follow the procedure described [below](#changelog))
  * Create release notes, in `RELEASE_NOTES.md`, highlighting the new features and changes in the version. 
    This is needed so the bot knows which entries to add to the release page on GitHub.
  * (To be added in the future) ~~Additionally verify that the `UPGRADING.md` file is up to date and contains all the necessary information for upgrading to the new version.~~
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
  * **Note:** The final release MUST have the same commit hash as the latest corresponding release candidate.
  * Create a new annotated git tag in the release branch (follow the [Tagging Procedure](#tagging-procedure)). This will trigger the automated release process (which will also create the release artifacts).
  * Once the release process completes, modify release notes if needed.

### Changelog

For PRs that are changing production code, please add a changelog entry in `CHANGELOG.md` (for details, see 
[contributing guidelines](./CONTRIBUTING.md#changelog)). 

#### Creating a new release branch 

Unreleased changes are collected on `main` in `CHANGELOG.md`. 
Thus, when creating a new release branch (e.g., `release/v11.x`), the following steps are necessary:

- create a new release branch, e.g., `release/v11.x`
    ```bash 
    git checkout main
    git pull 
    git checkout -b release/v11.x
    ```
- push the release branch upstream 
    ```bash 
    git push
    ```

#### Cutting a new release

Before cutting a _**release candidate**_ (e.g., `v11.0.0-rc0`), the following steps are necessary:

- move to the release branch, e.g., `release/v11.x`
    ```bash 
    git checkout release/v11.x
    ```
- move all entries in "CHANGELOG.md" from the `UNRELEASED` section to a new section with the proper release version,
  e.g., `v11.0.0`
- open a PR (from this new created branch) against the release branch, e.g., `release/v11.x`

Now you can cut the release candidate, e.g., v11.0.0-rc0 (follow the [Tagging Procedure](#tagging-procedure)).

#### Update the changelog on main

Once the **final release** is cut, the new changelog section must be added to main:

- checkout a new branch from the `main` branch, i.e.,
    ```bash
    git checkout main
    git pull 
    git checkout -b <username>/backport_changelog
    ```
- bring the new changelog section from the release branch into this branch
    ```bash
    git merge release/v11.x
    ```
- Note that if new entries have been created in the Changelog since the release was cut, you should preserve those 
  in the `UNRELEASED` section. That means you may have to do some manual cleanup here.
- Open a PR (from this new created branch) against `main`
  
### Release Notes

Release notes will be created using the `RELEASE_NOTES.md` from the release branch. 
Once the automated releases process is completed, please add any missing information the release notes using Github UI.

With every release, the `goreleaser` tool will create a file with all the build artifact checksums and upload it alongside the artifacts.
The file is called `SHA256SUMS-{{.version}}.txt` and contains the following:
```
098b00ed78ca01456c388d7f1f22d09a93927d7a234429681071b45d94730a05  gaiad_0.0.4_windows_arm64.exe
15b2b9146d99426a64c19d219234cd0fa725589c7dc84e9d4dc4d531ccc58bec  gaiad_0.0.4_darwin_amd64
604912ee7800055b0a1ac36ed31021d2161d7404cea8db8776287eb512cd67a9  gaiad_0.0.4_darwin_arm64
76e5ff7751d66807ee85bc5301484d0f0bcc5c90582d4ba1692acefc189392be  gaiad_0.0.4_linux_arm64
bcbca82da2cb2387ad6d24c1f6401b229a9b4752156573327250d37e5cc9bb1c  gaiad_0.0.4_windows_amd64.exe
f39552cbfcfb2b06f1bd66fd324af54ac9ee06625cfa652b71eba1869efe8670  gaiad_0.0.4_linux_amd64
```

### Tagging Procedure

**Important**: _**Always create tags from your local machine**_ since all release tags should be signed and annotated.
Using Github UI will create a `lightweight` tag, so it's possible that `gaiad version` returns a commit hash, instead of a tag.
This is important because most operators build from source, and having incorrect information when you run `make install && gaiad version` raises confusion.

The following steps are the default for tagging a specific branch commit using git on your local machine. Usually, release branches are labeled `release/v*`:

Ensure you have checked out the commit you wish to tag and then do:
```bash
git pull --tags

# test tag creation and releasing using goreleaser
make create-release-dry-run TAG=v11.0.0

# after successful test push the tag
make create-release TAG=v11.0.0
```

To re-create a tag:
```bash
# delete a tag locally
git tag -d v11.0.0  

# push the deletion to the remote
git push --delete origin v11.0.0 

# redo create-release
make create-release-dry-run TAG=v11.0.0
make create-release TAG=v11.0.0
```

#### Test building artifacts

Before tagging a new version, please test the building releasing artifacts by running:

```bash
TM_VERSION=$(go list -m github.com/tendermint/tendermint | sed 's:.* ::') goreleaser release --snapshot --clean --debug
```

#### Installing goreleaser
Check the instructions for installing goreleaser locally for your platform
* https://goreleaser.com/install/

## Release Policy

### Definitions

A `major` release is an increment of the _point number_ (eg: `v9.X.X → v10.X.X`).  
A `minor` release is an increment of the _point number_ (eg: `v9.0.X → v9.1.X`).  
A `patch` release is an increment of the patch number (eg: `v10.0.0` → `v10.0.1`).

### Policy

A `major` release will only be done via a governance gated upgrade. It can contain state breaking changes, and will 
also generally include new features, major changes to existing features, and/or large updates to key dependency 
packages such as CometBFT or the Cosmos SDK.

A `minor` release may be done via a governance gated upgrade, or via a coordinated upgrade on a predefined block 
height. It will contain breaking changes which require a coordinated upgrade, but the scope of these changes is 
limited to essential updates such as fixes for security vulnerabilities. 

Each vulnerability which requires a state breaking upgrade will be evaluated individually by the maintainers of the 
software and the maintainers will make a determination on whether to include the changes into a minor release.

A `patch` release will be created for changes which are strictly not state breaking. The latest patch release for a 
given release version is generally the recommended release, however, validator updates can be rolled out 
asynchronously without risking the state of a network running the software.

The intention of the Release Policy is to ensure that the latest gaia release is maintained with the following
categories of fixes:

- Tooling improvements (including code formatting, linting, static analysis and updates to testing frameworks)
- Performance enhancements for running archival and syncing nodes
- Test and benchmarking suites, ensuring that fixes are sound and there are no performance regressions
- Library updates including point releases for core libraries such as IBC-Go, Cosmos SDK, Tendermint and other dependencies
- General maintenance improvements, that are deemed necessary by the stewarding team, that help align different releases and reduce the workload on the stewarding team
- Security fixes

## Non-major Release Procedure

Updates to the release branch should come from `main` by backporting PRs 
(usually done by automatic cherry-pick followed by a PRs to the release branch). 
The backports must be marked using `backport/Y` label in PR for main.
It is the PR author's responsibility to fix merge conflicts, update changelog entries, and
ensure CI passes. If a PR originates from an external contributor, a member of the stewarding team assumes
responsibility to perform this process instead of the original author.

After the release branch has all commits required for the next patch release:

* Update the [changelog](#changelog) and the [release notes](#release-notes).
* Create a new annotated git tag in the release branch (follow the [Tagging Procedure](#tagging-procedure)). This will trigger the automated release process (which will also create the release artifacts).
* Once the release process completes, modify release notes if needed.
