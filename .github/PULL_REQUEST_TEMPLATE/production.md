<!--
The production pull request template is for types feat, fix, deps, or refactor.
-->

## Description

Closes: #XXXX

<!-- Add a description of the changes that this PR introduces and the files that
are the most critical to review. -->

<!-- Pull requests that sit inactive for longer than 14 days will be closed.  -->

---

### Author Checklist

*All items are required. Please add a note to the item if the item is not applicable and
please add links to any relevant follow up issues.*

I have...

* [ ] Included the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
* [ ] Added `!` to the type prefix if API, client, or state breaking change (i.e., requires minor or major version bump)
* [ ] Targeted the correct branch (see [PR Targeting](https://github.com/cosmos/gaia/blob/main/CONTRIBUTING.md#pr-targeting))
* [ ] Provided a link to the relevant issue or specification
* [ ] Followed the guidelines for [building SDK modules](https://github.com/cosmos/cosmos-sdk/blob/main/docs/docs/building-modules)
* [ ] Included the necessary unit and integration [tests](https://github.com/cosmos/gaia/blob/main/CONTRIBUTING.md#testing)
* [ ] Added a changelog entry in `.changelog` (for details, see [contributing guidelines](https://github.com/cosmos/gaia/blob/main/CONTRIBUTING.md#changelog))
* [ ] Included comments for [documenting Go code](https://blog.golang.org/godoc)
* [ ] Updated the relevant documentation or specification
* [ ] Reviewed "Files changed" and left comments if necessary <!-- relevant if the changes are not obvious  -->
* [ ] Confirmed all CI checks have passed

### Reviewers Checklist

*All items are required. Please add a note if the item is not applicable and please add
your handle next to the items reviewed if you only reviewed selected items.*

I have...

* [ ] confirmed the correct [type prefix](https://github.com/commitizen/conventional-commit-types/blob/v3.0.0/index.json) in the PR title
* [ ] confirmed `!` in the type prefix if API or client breaking change
* [ ] confirmed all author checklist items have been addressed 
* [ ] reviewed state machine logic
* [ ] reviewed API design and naming
* [ ] reviewed documentation is accurate
* [ ] reviewed tests and test coverage
