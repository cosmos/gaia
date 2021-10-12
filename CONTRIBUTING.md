# Contributing

- [Contributing](#contributing)
  - [Pull Requests](#pull-requests)
    - [Process for reviewing PRs](#process-for-reviewing-prs)
    - [Updating Documentation](#updating-documentation)
  - [Forking](#forking)
  - [Dependencies](#dependencies)
  - [Testing](#testing)
  - [Branching Model and Release](#branching-model-and-release)
    - [PR Targeting](#pr-targeting)
    - [Development Procedure](#development-procedure)
    - [Pull Merge Procedure](#pull-merge-procedure)
    - [Release Procedure](#release-procedure)
    - [Point Release Procedure](#point-release-procedure)

Thank you for considering making contributions to Gaia and related
repositories!

Contributing to this repo can mean many things such as participating in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, a general procedure for contributing has been established:

1. Either [open](https://github.com/cosmos/gaia/issues/new/choose) or
   [find](https://github.com/cosmos/gaia/issues) an issue you'd like to help with
2. Participate in thoughtful discussion on that issue
3. If you would like to contribute:
   1. If the issue is a proposal, ensure that the proposal has been accepted
   2. Ensure that nobody else has already begun working on this issue, if they have
      make sure to contact them to collaborate
   3. If nobody has been assigned the issue and you would like to work on it
      make a comment on the issue to inform the community of your intentions
      to begin work
   4. Follow standard Github best practices: fork the repo, branch from the
      HEAD of `main`, make some commits, and submit a PR to `main`
      - For core developers working within the Gaia repo, to ensure a clear
      ownership of branches, branches must be named with the convention
      `{moniker}/{issue#}-branch-name`
   5. Be sure to submit the PR in `Draft` mode. Submit your PR early, even if
      it's incomplete as this indicates to the community you're working on
      something and allows them to provide comments early in the development process
   6. When the code is complete it can be marked `Ready for Review`
   7. Be sure to include a relevant change log entry in the `Unreleased` section
      of `CHANGELOG.md` (see file for log format)

Note that for very small or blatantly obvious problems (such as typos) it is
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a github issue, that PR runs a high likelihood of being rejected.

Take a peek at our [coding repo](https://github.com/tendermint/coding) for
overall information on repository workflow and standards. Note, we use `make
tools` for installing the linting tools.

Other notes:

- Looking for a good place to start contributing? How about checking out some
  [good first issues](https://github.com/cosmos/gaia/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
- Please make sure to run `make format` before every commit - the easiest way
  to do this is have your editor run it for you upon saving a file. Additionally
  please ensure that your code is lint compliant by running `golangci-lint run`.
  A convenience git `pre-commit` hook that runs the formatters automatically
  before each commit is available in the `contrib/githooks/` directory.

## Submissions

Generally, when structuring a submission for the Cosmos ecosystem, the submission might span multiple layers
of the technology stack, including the Cosmos SDK and Tendermint dependencies.

### Structuring a submission

To accommodate review process we suggest that PRs are categorically broken up.
Ideally each PR addresses only a single issue. Additionally, as much as possible
code refactoring and cleanup should be submitted as a separate PRs from bugfixes/feature-additions.

In order to simplify reviewing large changes, submissions should have a created an issue
with a description of the submission, a description tracking the changes and relevant discussions,
and a checklist of changes and tasks to be done. 

The issue can then be used to develop multiple well-scoped PRs that are easy to review.

The following PR structuring checklist can be used when submitting changes to the Gaia repository for review:
- [ ] Proto files: PR updating proto files. As a suggested next step, don't regenerate updated protobuf 
   implementations using `protgen`, since this will break existing code.   
- [ ] Broken code: If `protogen` is run, a PR disabling broken code
- [ ] Validation: PR with validation of types
- [ ] Functionality: PR integrating supporting functionality
- [ ] Servers: PR for `msgserver` and `queryserver`
- [ ] CLI: PR for CLI commands
- [ ] Orchestrators: PR for any orchestrators
- [ ] Genesis: PR for genesis
- [ ] Upgrades: PR for upgrades

### Process for reviewing PRs

All PRs require at least one review before merge (except docs changes, or variable name-changes which only require one). When reviewing PRs please use the following review explanations:

- `LGTM` without an explicit approval means that the changes look good, but you haven't pulled down the code, run tests locally and thoroughly reviewed it.
- `Approval` through the GH UI means that you understand the code, documentation/spec is updated in the right places, you have pulled down and tested the code locally. In addition:
  - You must also think through anything which ought to be included but is not
  - You must think through whether any added code could be partially combined (DRYed) with existing code
  - You must think through any potential security issues or incentive-compatibility flaws introduced by the changes
  - Naming must be consistent with conventions and the rest of the codebase
  - Code must live in a reasonable location, considering dependency structures (e.g. not importing testing modules in production code, or including example code modules in production code).
  - if you approve of the PR, you are responsible for fixing any of the issues mentioned here and more
- If you sat down with the PR submitter and did a pairing review please note that in the `Approval`, or your PR comments.
- If you are only making "surface level" reviews, submit any notes as `Comments` without adding a review.

### Updating Documentation

If you open a PR in Gaia, it is mandatory to update the relevant documentation in /docs.

- If your changes relate specifically to the gaia application, please modify the docs/ folder.

## Forking

Please note that Go requires code to live under absolute paths, which complicates forking.
While my fork lives at `https://github.com/rigeyrigerige/gaia`,
the code should never exist at  `$GOPATH/src/github.com/rigeyrigerige/gaia`.
Instead, we use `git remote` to add the fork as a new remote for the original repo,
`$GOPATH/src/github.com/cosmos/gaia`, and do all the work there.

For instance, to create a fork and work on a branch of it, I would:

- Create the fork on github, using the fork button.
- Go to the original repo checked out locally (i.e. `$GOPATH/src/github.com/cosmos/gaia`)
- `git remote rename origin upstream`
- `git remote add origin git@github.com:rigeyrigerige/gaia.git`

Now `origin` refers to my fork and `upstream` refers to the Gaia version.
So I can `git push -u origin main` to update my fork, and make pull requests to Gaia from there.
Of course, replace `rigeyrigerige` with your git handle.

To pull in updates from the origin repo, run

- `git fetch upstream`
- `git rebase upstream/main` (or whatever branch you want)

Please don't make Pull Requests to `main`.

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

## Testing

All repos should be hooked up to [CircleCI](https://circleci.com/).

If they have `.go` files in the root directory, they will be automatically
tested by circle using `go test -v -race ./...`. If not, they will need a
`circle.yml`. Ideally, every repo has a `Makefile` that defines `make test` and
includes its continuous integration status using a badge in the `README.md`.

We expect tests to use `require` or `assert` rather than `t.Skip` or `t.Fail`,
unless there is a reason to do otherwise.
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

## Branching Model and Release

User-facing repos should adhere to the [trunk based development branching model](https://trunkbaseddevelopment.com/).

Libraries need not follow the model strictly, but would be wise to.

Gaia utilizes [semantic versioning](https://semver.org/).

### PR Targeting

Ensure that you base and target your PR on the `main` branch.

All feature additions should be targeted against `main`. Bug fixes for an outstanding release candidate
should be targeted against the release candidate branch. Release candidate branches themselves should be the
only pull requests targeted directly against main.

### Development Procedure

- the latest state of development is on `main`
- `main` must never fail `make test` or `make test_cli`
- `main` should not fail `make lint`
- no `--force` onto `main` (except when reverting a broken commit, which should seldom happen)
- create a development branch either on `https://github.com/cosmos/gaia`, or your fork (using `git remote add origin`)
- before submitting a pull request, begin `git rebase` on top of `main`

### Pull Merge Procedure

- ensure pull branch is rebased on `main`
- run `make test` and `make test_cli` to ensure that all tests pass
- merge pull request
