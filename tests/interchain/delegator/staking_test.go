package delegator_test

import (
	"testing"
	"time"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type StakingSuite struct {
	*delegator.Suite
}

func (s *StakingSuite) TestDelegateWithdrawUnbond() {
	// delegate tokens
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(),
	)
	s.Require().NoError(err)

	startingBalance, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	time.Sleep(20 * time.Second)
	// Withdraw rewards
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"distribution", "withdraw-rewards", s.Chain.ValidatorWallets[0].ValoperAddress,
	)
	s.Require().NoError(err)
	endingBalance, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	s.Require().Truef(endingBalance.GT(startingBalance), "endingBalance: %s, startingBalance: %s", endingBalance, startingBalance)

	// Unbond tokens
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"staking", "unbond", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(),
	)
	s.Require().NoError(err)
}

func TestStaking(t *testing.T) {
	s := &StakingSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
