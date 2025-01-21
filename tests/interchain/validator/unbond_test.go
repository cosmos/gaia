package validator_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type UnbondingSuite struct {
	*chainsuite.Suite
}

func (s *UnbondingSuite) TestUnbondValidator() {
	_, err := s.Chain.Validators[5].ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[5].Moniker,
		"staking", "unbond-validator",
	)
	s.Require().NoError(err)
	validator, err := s.Chain.StakingQueryValidator(s.GetContext(), s.Chain.ValidatorWallets[5].ValoperAddress)
	s.Require().NoError(err)
	s.Require().Equal(stakingtypes.Unbonding, validator.Status)

	_, err = s.Chain.Validators[5].ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[5].Moniker,
		"slashing", "unjail",
	)
	s.Require().NoError(err)

	validator, err = s.Chain.StakingQueryValidator(s.GetContext(), s.Chain.ValidatorWallets[5].ValoperAddress)
	s.Require().NoError(err)
	s.Require().Equal(stakingtypes.Bonded, validator.Status)
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
