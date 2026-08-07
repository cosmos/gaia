package integrator_test

import (
	"fmt"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
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
		ChainName:     "gaia-reimport",
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

	// The max_validators param will be set to the max_provider_consensus_validators value during the v28.0.0 upgrade,
	// so we check that the correct number of validators are bonded and that the next validator is unbonded (i.e. not in the active set).
	maxValidatorsJSON, err := newChain.QueryJSON(s.GetContext(), "params.max_validators", "staking", "params")
	s.Require().NoError(err)
	maxValidatorsParam := int(maxValidatorsJSON.Int())

	for i := range maxValidatorsParam {
		s.Require().Equal(stakingtypes.Bonded, validators[i].Status)
	}
	s.Require().NotEqual(stakingtypes.Bonded, validators[maxValidatorsParam].Status)

	vals, err := newChain.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(maxValidatorsParam, len(vals.Array()), vals)
	for i := range maxValidatorsParam {
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
