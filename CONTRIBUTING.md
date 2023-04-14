# Contributing

- [Contributing](#contributing)
  - [Team Organization](#team-organization)
  - [Development Procedure](#development-procedure)
    - [Testing](#testing)
    - [Pull Requests](#pull-requests)
    - [Pull Request Templates](#pull-request-templates)
    - [Requesting Reviews](#requesting-reviews)
    - [Updating Documentation](#updating-documentation)
  - [Dependencies](#dependencies)
  - [Protobuf](#protobuf)
  - [Branching Model and Release](#branching-model-and-release)
    - [PR Targeting](#pr-targeting)

Thank you for considering making contributions to Gaia! 🎉👍

Contributing to this repo can mean many things such as participating in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, a general procedure for contributing has been established:

1. Start by browsing [existing issues](https://github.com/cosmos/gaia/issues) and [discussions](https://github.com/cosmos/gaia/discussions). If you are looking for something interesting or if you have something in your mind, there is a chance it had been discussed.
   * Looking for a good place to start contributing? How about checking out some [good first issues](https://github.com/cosmos/gaia/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) or [bugs](https://github.com/cosmos/gaia/issues?q=is%3Aopen+is%3Aissue+label%3Abug)?
2. Determine whether a GitHub issue or discussion is more appropriate for your needs:
   1. If you want to propose something new that requires specification or an additional design, or you would like to change a process, start with a [new discussion](https://github.com/cosmos/gaia/discussions/new/choose). With discussions, we can better handle the design process using discussion threads. A discussion usually leads to one or more issues.
   2. If the issue you want addressed is a specific proposal or a bug, then open a [new issue](https://github.com/cosmos/gaia/issues/new/choose).
   3. Review existing [issues](https://github.com/cosmos/gaia/issues) to find an issue you'd like to help with.
3. Participate in thoughtful discussion on that issue.
4. If you would like to contribute:
   1. Ensure that the proposal has been accepted.
   2. Ensure that nobody else has already begun working on this issue. If they have,
      make sure to contact them to collaborate.
   3. If nobody has been assigned for the issue and you would like to work on it,
      make a comment on the issue to inform the community of your intentions
      to begin work.
5. To submit your work as a contribution to the repository, follow standard GitHub best practices. See [pull request guideline](#pull-requests) below.

**Note:** For very small or blatantly obvious problems such as typos, you are
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a GitHub issue, that PR runs a high likelihood of being rejected 
without review.

## Team Organization

Gaia has many stakeholders contributing and shaping the project. The Core Gaia team is composed of Informal Systems developers. Any long-term contributors and additional maintainers from other projects are welcome. We use self-organizing principles to coordinate and collaborate across organizations in structured "EPIC" that focus on specific problem domains or architectural components of Gaia. For details, see the [GitHub Project board](https://github.com/orgs/cosmos/projects/28/views/11). 

The developers work in sprints, which are available in a [GitHub Project](https://github.com/orgs/cosmos/projects/28/views/2). 

## Development Procedure

* The latest state of development is on `main`.
* `main` must never fail `make lint run-tests`.
* No `--force` onto `main` (except when reverting a broken commit, which should seldom happen).
* Create a branch to start work:
    * Fork the repo (core developers must create a branch directly in the Gaia repo),
    branch from the HEAD of `main`, make some commits, and submit a PR to `main`.
    * For core developers working within the `gaia` repo, follow branch name conventions to ensure clear
    ownership of branches: `{moniker}/{issue#}-branch-name`.
    * See [Branching Model](#branching-model-and-release) for more details.
* Be sure to run `make format` before every commit. The easiest way
  to do this is have your editor run it for you upon saving a file (most of the editors
  will do it anyway using a pre-configured setup of the programming language mode).
  A convenience git `pre-commit` hook that runs the formatters automatically
  before each commit is available in the `contrib/githooks/` directory.
* Follow the [CODING GUIDELINES](docs/guidelines/code-guidelines.md), which defines criteria for designing and coding a software.

Code is merged into main through pull request procedure.

### Testing

Tests can be executed by running `make run-tests` at the top level of the Gaia repository.
For running the e2e tests, make sure to build the docker images by running `make docker-build-all`.

When testing a function under a variety of different inputs, we prefer to use
[table driven tests](https://github.com/golang/go/wiki/TableDrivenTests).
Table driven test error messages should follow the following format
`<desc>, tc #<index>, i #<index>`.
`<desc>` is an optional short description of whats failing, `tc` is the
index within the table of the testcase that is failing, and `i` is when there
is a loop, exactly which iteration of the loop failed.
The idea is you should be able to see the
error message and figure out exactly what failed.
Here is an example check:

```go
<some table>
for tcIndex, tc := range cases {
<some code>
for i := 0; i < tc.numTxsToTest; i++ {
<some code>
require.Equal(t, expectedTx[:32], calculatedTx[:32],
"First 32 bytes of the txs differed. tc #%d, i #%d", tcIndex, i)
```

### Pull Requests

Before submitting a pull request:

* merge the latest main `git merge origin/main`,
* run `make lint run-tests` to ensure that all checks and tests pass.

Then:

1. If you have something to show, **start with a `Draft` PR**. It's good to have early validation of your work and we highly recommend this practice. A Draft PR also indicates to the community that the work is in progress.
   Draft PRs also help the core team provide early feedback and ensure the work is in the right direction.
2. When the code is complete, change your PR from `Draft` to `Ready for Review`.
3. Go through the actions for each checkbox present in the PR template description. The PR actions are automatically provided for each new PR.
4. Be sure to include a relevant changelog entry in the `Unreleased` section of `CHANGELOG.md` (see file for log format). The entry should be on top of all others changes in the section.

PRs must have a category prefix that is based on the type of changes being made (for example, `fix`, `feat`,
`refactor`, `docs`, and so on). The *type* must be included in the PR title as a prefix (for example,
`fix: <description>`). This convention ensures that all changes that are committed to the base branch follow the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
Additionally, **each PR should only address a single issue**.

Pull requests are merged automatically using [`A:automerge` action](https://mergify.io/features/auto-merge).

NOTE: when merging, GitHub will squash commits and rebase on top of the main.

### Pull Request Templates

There are three PR templates. The [default template](./.github/PULL_REQUEST_TEMPLATE.md) is for types `fix`, `feat`, and `refactor`. We also have a [docs template](./.github/PULL_REQUEST_TEMPLATE/docs.md) for documentation changes and an [other template](./.github/PULL_REQUEST_TEMPLATE/other.md) for changes that do not affect production code. When previewing a PR before it has been opened, you can change the template by adding one of the following parameters to the url:

* `template=docs.md`
* `template=other.md`

### Requesting Reviews

In order to accommodate the review process, the author of the PR must complete the author checklist
(from the pull request template)
to the best of their abilities before marking the PR as "Ready for Review". If you would like to
receive early feedback on the PR, open the PR as a "Draft" and leave a comment in the PR indicating
that you would like early feedback and tagging whoever you would like to receive feedback from.

Codeowners are marked automatically as the reviewers.

All PRs require at least two review approvals before they can be merged (one review might be acceptable in
the case of minor changes to [docs](./.github/PULL_REQUEST_TEMPLATE/docs.md) or [other](./.github/PULL_REQUEST_TEMPLATE/other.md) changes that do not affect production code). Each PR template has a reviewers checklist that must be completed before the PR can be merged. Each reviewer is responsible
for all checked items unless they have indicated otherwise by leaving their handle next to specific
items. In addition, use the following review explanations:

* `LGTM` without an explicit approval means that the changes look good, but you haven't thoroughly reviewed the reviewer checklist items.
* `Approval` means that you have completed some or all of the reviewer checklist items. If you only reviewed selected items, you must add your handle next to the items that you have reviewed. In addition, follow these guidelines:
    * You must also think through anything which ought to be included but is not
    * You must think through whether any added code could be partially combined (DRYed) with existing code
    * You must think through any potential security issues or incentive-compatibility flaws introduced by the changes
    * Naming must be consistent with conventions and the rest of the codebase
    * Code must live in a reasonable location, considering dependency structures (for example, not importing testing modules in production code, or including example code modules in production code).
    * If you approve the PR, you are responsible for any issues mentioned here and any issues that should have been addressed after thoroughly reviewing the reviewer checklist items in the pull request template.
* If you sat down with the PR submitter and did a pairing review, add this information in the `Approval` or your PR comments.
* If you are only making "surface level" reviews, submit notes as a `comment` review.

### Updating Documentation

If you open a PR in Gaia, it is mandatory to update the relevant documentation in `/docs`.

## Dependencies

We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage
dependency versions.

The main branch of every Cosmos repository should just build with `go get`,
which means they should be kept up-to-date with their dependencies so we can
get away with telling people they can just `go get` our software.

When dependencies in Gaia's `go.mod` are changed, it is generally accepted practice
to delete `go.sum` and then run `go mod tidy`.

Since some dependencies are not under our control, a third party may break our
build, in which case we can fall back on `go mod tidy -v`.

## Protobuf

We use [Protocol Buffers](https://developers.google.com/protocol-buffers) along with [gogoproto](https://github.com/cosmos/gogoproto) to generate code for use in Gaia.

For deterministic behavior around Protobuf tooling, everything is containerized using Docker. Make sure to have Docker installed on your machine, or head to [Docker's website](https://docs.docker.com/get-docker/) to install it.

To generate the protobuf stubs, you can run `make proto-gen`.

## Branching Model and Release

User-facing repos should adhere to the trunk based development branching model: https://trunkbaseddevelopment.com. User branches should start with a user name, example: `{moniker}/{issue#}-branch-name`.

Gaia follows [semantic versioning](https://semver.org), but with the following deviations to account for state-machine and API breaking changes: 

- State-machine breaking changes will result in an increase of the major version X (X.y.z).
- API breaking changes will result in an increase of the minor version Y (x.Y.z | x > 0).
- All other changes will result in an increase of the patch version Z (x.y.Z | x > 0).

### PR Targeting

Ensure that you base and target your PR on the `main` branch.

All feature additions and all bug fixes must be targeted against `main`. Exception is for bug fixes which are only related to a released version. In that case, the related bug fix PRs must target against the release branch.

If needed, we backport a commit from `main` to a release branch (excluding consensus breaking feature, API breaking and similar).

