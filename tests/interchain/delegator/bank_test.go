package delegator_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type BankSuite struct {
	*delegator.Suite
}

func (s *BankSuite) TestSend() {
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), s.Chain.Config().Denom)
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet2.FormattedAddress(), txAmountUatom(),
	)
	s.Require().NoError(err)

	balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), s.Chain.Config().Denom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.Add(sdkmath.NewInt(txAmount)), balanceAfter)
}

func TestBank(t *testing.T) {
	s := &BankSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
