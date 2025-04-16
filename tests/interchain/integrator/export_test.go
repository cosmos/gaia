package integrator_test

import (
	"fmt"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"
)

const (
	maxValidators          = 5
	maxConsensusValidators = 4
)

type ExportSuite struct {
	*chainsuite.Suite
}

func (s *ExportSuite) TestExportAndImportValidators() {
	height, err := s.Chain.Height(s.GetContext())
	s.Require().NoError(err)

	s.Require().NoError(s.Chain.StopAllNodes(s.GetContext()))

	exported, err := s.Chain.ExportState(s.GetContext(), height)
	s.Require().NoError(err)
	newConfig := chainsuite.DefaultChainSpec(s.Env).ChainConfig
	newConfig.ModifyGenesis = func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		return []byte(exported), nil
	}
	newConfig.PreGenesis = func(c ibc.Chain) error {
		for i, val := range s.Chain.Validators {
			key, err := val.PrivValFileContent(s.GetContext())
			if err != nil {
				return fmt.Errorf("failed to get key for validator %d: %w", i, err)
			}
			cosmosChain := c.(*cosmos.CosmosChain)
			err = cosmosChain.Validators[i].OverwritePrivValFile(s.GetContext(), key)
			if err != nil {
				return fmt.Errorf("failed to overwrite priv val file for validator %d: %w", i, err)
			}
		}
		return nil
	}
	zero := 0
	newSpec := &interchaintest.ChainSpec{
		Name:          "gaia",
		Version:       s.Env.NewGaiaImageVersion,
		NumValidators: &chainsuite.SixValidators,
		NumFullNodes:  &zero,
		ChainConfig:   newConfig,
	}
	newChain, err := chainsuite.CreateChain(s.GetContext(), s.T(), newSpec)
	s.Require().NoError(err)

	validators := []*stakingtypes.Validator{}
	// We go through the original chain's wallets because we want to make sure those validators still exist in the new chain.
	for _, wallet := range s.Chain.ValidatorWallets {
		validator, err := newChain.StakingQueryValidator(s.GetContext(), wallet.ValoperAddress)
		s.Require().NoError(err)
		validators = append(validators, validator)
	}
	for i := 0; i < maxValidators; i++ {
		s.Require().Equal(stakingtypes.Bonded, validators[i].Status)
	}
	s.Require().Equal(stakingtypes.Unbonded, validators[maxValidators].Status)

	vals, err := newChain.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(maxConsensusValidators, len(vals.Array()), vals)
	for i := 0; i < maxConsensusValidators; i++ {
		valCons := vals.Array()[i].Get("address").String()
		s.Require().NoError(err)
		s.Require().Equal(s.Chain.ValidatorWallets[i].ValConsAddress, valCons)
	}
}

func TestExport(t *testing.T) {
	genesis := chainsuite.DefaultGenesis()
	genesis = append(genesis,
		cosmos.NewGenesisKV("app_state.staking.params.max_validators", maxValidators),
		cosmos.NewGenesisKV("app_state.provider.params.max_provider_consensus_validators", maxConsensusValidators),
	)
	s := &ExportSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			UpgradeOnSetup: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
				ChainConfig: ibc.ChainConfig{
					ModifyGenesis: cosmos.ModifyGenesis(genesis),
				},
			},
		}),
	}
	suite.Run(t, s)
}
