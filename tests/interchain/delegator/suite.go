package delegator

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

// Suite is a common base suite for all delegator tests.
type Suite struct {
	*chainsuite.Suite
	DelegatorWallet  ibc.Wallet
	DelegatorWallet2 ibc.Wallet
	DelegatorWallet3 ibc.Wallet
}

func (s *Suite) SetupSuite() {
	s.Suite.SetupSuite()
	wallet, err := s.Chain.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.DelegatorWallet = wallet
	s.Require().NoError(s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DelegatorWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}))

	wallet, err = s.Chain.BuildWallet(s.GetContext(), "delegator2", "")
	s.Require().NoError(err)
	s.DelegatorWallet2 = wallet

	wallet, err = s.Chain.BuildWallet(s.GetContext(), "delegator3", "")
	s.Require().NoError(err)
	s.DelegatorWallet3 = wallet

	s.Require().NoError(s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DelegatorWallet2.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}))

	s.Require().NoError(s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DelegatorWallet3.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}))
}
