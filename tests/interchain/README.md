# Interchain tests for gaia.


These tests use [interchaintest](https://github.com/strangelove-ventures/interchaintest/) to
create, upgrade, and test chains.

They dockerize the validators, so they depend on a `gaia` docker image being built.
You can build a docker image using the [docker-push](../../.github/workflows/docker-push.yml) workflow.
`docker-push` runs nightly on the `main` branch, and for all new releases, but you can also
[run it manually on any branch](https://github.com/cosmos/gaia/actions/workflows/docker-push.yml)

Once the `gaia` image is built, the `docker-push` action workflow automatically invokes the
[interchain-test](../../.github/workflows/interchain-test.yml) workflow.

Read on to learn how these tests work.

## Upgrade testing

The tests will make sure it's possible to upgrade from a previous version of
`gaia` to the current version being tested. It does so by starting a chain from genesis
on the previous version, then upgrading it to the current version.

## Version selection

The `interchain-test` workflow will start by selecting versions to test upgrading from.

The [`matrix_tool`](./matrix_tool/main.go) tool will take the tag of the image
being tested (e.g. `v18.0.0`, or `main`, or `some-feature-branch`), and figure
out a corresponding semver. If the tag is already a valid semver, that's the
version. Otherwise, it will take the major version from the module line in `go.mod`,
and append `.999.0`. Given that semver, it'll figure out:

* The previous rc (if the current version is itself an rc)
* The previous minor version (if applicable)
* The previous major version

For instance, for `v15.1.0-rc1`, we'll test upgrading from:
* `v15.1.0-rc0`
* `v15.0.0`
* `v14.2.0`

The workflow will then test upgrading from each of those three to the current
version. When it's a major upgrade, it will do so via governance proposal,
otherwise it'll simply stop the old image and start the new one.

## Test Suites

Each of the *_test.go files in this directory contains a test suite.  These
share some common scaffolding (a `SetupSuite`) to create and upgrade a chain,
and then run a set of tests on that chain.

So, for instance, a transactions suite:

```go
type TxSuite struct {
	*chainsuite.Suite
}
```

It extends `chainsuite.Suite,` so its SetupSuite will create and upgrade a
chain (more on this later). The individual `Test*` methods then run the gaia
version being tested:

```go
func (s *TxSuite) TestBankSend() {
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.Chain.ValidatorWallets[1].Address, chainsuite.Uatom)
	s.Require().NoError(err)

	_, err = s.Chain.Validators[0].ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Moniker,
		"bank", "send",
		s.Chain.ValidatorWallets[0].Address, s.Chain.ValidatorWallets[1].Address, txAmountUatom(),
	)
	s.Require().NoError(err)

	balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.Chain.ValidatorWallets[1].Address, chainsuite.Uatom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.Add(sdkmath.NewInt(txAmount)), balanceAfter)
}
```

Because of how testify works, we have to instantiate each suite to run it.
This is also where we tell the suite to run an upgrade on Setup:

```go
func TestTransactions(t *testing.T) {
	txSuite := TxSuite{chainsuite.NewSuite(chainsuite.SuiteConfig{UpgradeOnSetup: true})}
	suite.Run(t, &txSuite)
}
```

Of course, we can also parameterize the test suites themselves. This enables us
to write tests once and run them a bunch of times on different configurations:

```go
type ConsumerLaunchSuite struct {
	*chainsuite.Suite
	OtherChain            string
	OtherChainVersion     string
	ShouldCopyProviderKey [chainsuite.ValidatorCount]bool
}

func TestICS40ChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v4.0.0",
		ShouldCopyProviderKey: noProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS33ConsumerAllKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v3.3.0",
		ShouldCopyProviderKey: allProviderKeysCopied(),
	}
	suite.Run(t, s)
}
```

Notice also how `UpgradeOnSetup` isn't set here: the ConsumerLaunchSuite needs
to be handed a pre-upgrade chain so it can make sure that a consumer chain that
launched before the upgrade keeps working after the upgrade.


## Writing new tests

All you need to start writing new tests is a test suite as described above.
The suite will have an `s.Chain` that you can test. Check out utilities in
[`chainsuite/chain.go`](./chainsuite/chain.go) and
[`chain_ics.go`](./chainsuite/chain_ics.go) for some convenience methods.

In addition, the s.Chain object extends the `interchaintest` chain object, so
check out [the docs](https://pkg.go.dev/github.com/strangelove-ventures/interchaintest/v7) to
see what else is available.
