package validator_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gaia/v25/tests/interchain/chainsuite"
	"github.com/cosmos/interchaintest/v10"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type UnbondingSuite struct {
	*chainsuite.Suite
}

func (s *UnbondingSuite) TestUnbondValidator() {
	_, stake := chainsuite.DefaultGenesisAmounts(s.Chain.Config().Denom)(5)
	txhash, err := s.Chain.Validators[5].ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[5].Moniker,
		"staking", "unbond", s.Chain.ValidatorWallets[5].ValoperAddress, stake.String(),
	)
	s.Require().NoError(err)
	validator, err := s.Chain.StakingQueryValidator(s.GetContext(), s.Chain.ValidatorWallets[5].ValoperAddress)
	s.Require().NoError(err)
	s.Require().Equal(stakingtypes.Unbonding, validator.Status)

	tx, err := s.Chain.GetTransaction(txhash)
	s.Require().NoError(err)

	_, err = s.Chain.Validators[5].ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[5].Moniker,
		"staking", "cancel-unbond", s.Chain.ValidatorWallets[5].ValoperAddress,
		stake.String(), fmt.Sprintf("%d", tx.Height),
	)
	s.Require().ErrorContains(err, "validator for this address is currently jailed")
}

func TestUnbonding(t *testing.T) {
	txSuite := UnbondingSuite{chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		ChainSpec: &interchaintest.ChainSpec{
			NumValidators: &chainsuite.SixValidators,
		},
	})}
	suite.Run(t, &txSuite)
}
