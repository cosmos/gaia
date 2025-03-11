package validator_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/mod/semver"
)

const (
	maxValidators                  = 5
	maxProviderConsensusValidators = 4
)

type InactiveValidatorsSuite struct {
	*chainsuite.Suite
	Consumer *chainsuite.Chain
}

func (s *InactiveValidatorsSuite) SetupSuite() {
	s.Suite.SetupSuite()
	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	stakingProposal := fmt.Sprintf(`{
		"@type": "/cosmos.staking.v1beta1.MsgUpdateParams",
		"authority": "%s",
		"params": {
			"unbonding_time": "1814400s",
			"max_validators": 5,
			"max_entries": 7,
			"historical_entries": 10000,
			"bond_denom": "%s",
			"min_commission_rate": "0.050000000000000000",
			"validator_bond_factor": "250.000000000000000000",
			"global_liquid_staking_cap": "0.250000000000000000",
			"validator_liquid_staking_cap": "0.500000000000000000"
		}	
	}`, authority, s.Chain.Config().Denom)

	prop, err := s.Chain.BuildProposal(nil, "update staking params", "update staking params", "", chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(stakingProposal)}
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), result.ProposalID))

	s.UpgradeChain()

	stakingParams, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "staking", "params")
	s.Require().NoError(err)

	providerParams, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "provider", "params")
	s.Require().NoError(err)

	if semver.Compare(s.Env.OldGaiaImageVersion, "v20.0.0") < 0 {
		// These are set by the v20 upgrade handler
		s.Require().Equal(uint64(200), gjson.GetBytes(stakingParams, "params.max_validators").Uint(), string(stakingParams))
		s.Require().Equal(uint64(180), gjson.GetBytes(providerParams, "max_provider_consensus_validators").Uint(), string(providerParams))
	}

	providerParams, err = sjson.SetBytes(providerParams, "max_provider_consensus_validators", maxProviderConsensusValidators)
	s.Require().NoError(err)
	providerProposal, err := sjson.SetRaw(fmt.Sprintf(`{
		"@type": "/interchain_security.ccv.provider.v1.MsgUpdateParams",
		"authority": "%s"
	}`, authority), "params", string(providerParams))
	s.Require().NoError(err)

	stakingProposal, err = sjson.Set(stakingProposal, "params.max_validators", maxValidators)
	s.Require().NoError(err)
	prop, err = s.Chain.BuildProposal(nil, "update staking params", "update staking params", "", chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	prop.Messages = append(prop.Messages, json.RawMessage(stakingProposal))
	result, err = s.Chain.SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), result.ProposalID))

	prop, err = s.Chain.BuildProposal(nil, "update provider params", "update provider params", "", chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(providerProposal)}
	result, err = s.Chain.SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), result.ProposalID))

	cfg := chainsuite.ConsumerConfig{
		ChainName:             "ics-consumer",
		Version:               "v6.2.1",
		ShouldCopyProviderKey: []bool{true, true, true, true, true, true},
		Denom:                 chainsuite.Ucon,
		TopN:                  100,
		AllowInactiveVals:     true,
		MinStake:              1_000_000,
		Spec: &interchaintest.ChainSpec{
			ChainConfig: ibc.ChainConfig{
				Images: []ibc.DockerImage{
					{
						Repository: chainsuite.HyphaICSRepo,
						Version:    "v6.2.1",
						UIDGID:     chainsuite.ICSUidGuid,
					},
				},
			},
		},
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	s.Consumer = consumer
}

// This is called 0ValidatorSets because it should run first; if the validator sets are wrong, obviously the other tests will fail
func (s *InactiveValidatorsSuite) Test0ValidatorSets() {
	vals, err := s.Chain.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(maxProviderConsensusValidators, len(vals.Array()), vals)
	for i := 0; i < maxProviderConsensusValidators; i++ {
		valCons := vals.Array()[i].Get("address").String()
		s.Require().NoError(err)
		s.Require().Equal(s.Chain.ValidatorWallets[i].ValConsAddress, valCons)
	}

	vals, err = s.Consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(maxProviderConsensusValidators, len(vals.Array()), vals)
	for i := 0; i < maxProviderConsensusValidators; i++ {
		valCons := vals.Array()[i].Get("address").String()
		s.Require().NoError(err)
		s.Require().Equal(s.Consumer.ValidatorWallets[i].ValConsAddress, valCons)
	}
}

func (s *InactiveValidatorsSuite) TestProviderJailing() {
	for i := 1; i < maxProviderConsensusValidators; i++ {
		jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Chain, i)
		s.Require().NoError(err)
		s.Assert().True(jailed, "validator %d should be jailed", i)
	}
	for i := maxProviderConsensusValidators; i < len(s.Chain.Validators); i++ {
		jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Chain, i)
		s.Require().NoError(err)
		s.Assert().False(jailed, "validator %d should not be jailed", i)
	}
}

func (s *InactiveValidatorsSuite) TestConsumerJailing() {
	for i := 1; i < maxProviderConsensusValidators; i++ {
		jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Consumer, i)
		s.Require().NoError(err)
		s.Assert().True(jailed, "validator %d should be jailed", i)
	}
	// Validator 4 will have been opted in automatically when the other ones went down
	_, err := s.Chain.Validators[maxProviderConsensusValidators].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[maxProviderConsensusValidators].Moniker, "provider", "opt-out", s.getConsumerID())
	s.Require().NoError(err)
	for i := maxProviderConsensusValidators; i < len(s.Chain.Validators); i++ {
		jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Consumer, i)
		s.Require().NoError(err)
		s.Assert().False(jailed, "validator %d should not be jailed", i)
	}
}

func (s *InactiveValidatorsSuite) TestOptInInactive() {
	consumerID := s.getConsumerID()
	// Validator 4 will have been opted in automatically when the other ones went down
	_, err := s.Chain.Validators[maxProviderConsensusValidators].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[maxProviderConsensusValidators].Moniker, "provider", "opt-out", s.getConsumerID())
	s.Require().NoError(err)

	_, err = s.Chain.Validators[4].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[4].Moniker, "provider", "opt-in", consumerID)
	s.Require().NoError(err)
	defer func() {
		_, err := s.Chain.Validators[4].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[4].Moniker, "provider", "opt-out", consumerID)
		s.Require().NoError(err)
		s.Relayer.ClearCCVChannel(s.GetContext(), s.Chain, s.Consumer)
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			vals, err := s.Consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
			assert.NoError(c, err)
			assert.Equal(c, maxProviderConsensusValidators, len(vals.Array()), vals)
		}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
		jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Consumer, 4)
		s.Require().NoError(err)
		s.Assert().False(jailed, "validator 4 should not be jailed")
	}()
	s.Require().NoError(s.Relayer.ClearCCVChannel(s.GetContext(), s.Chain, s.Consumer))
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		vals, err := s.Consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
		assert.NoError(c, err)
		assert.Equal(c, maxProviderConsensusValidators+1, len(vals.Array()), vals)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Consumer, 4)
	s.Require().NoError(err)
	s.Assert().True(jailed, "validator 4 should be jailed")

	_, err = s.Chain.Validators[5].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[5].Moniker, "provider", "opt-in", consumerID)
	s.Require().NoError(err)
	s.Require().NoError(s.Relayer.ClearCCVChannel(s.GetContext(), s.Chain, s.Consumer))
	vals, err := s.Consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(maxProviderConsensusValidators+1, len(vals.Array()), vals)
	jailed, err = s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, s.Consumer, 5)
	s.Require().NoError(err)
	s.Assert().False(jailed, "validator 5 should not be jailed")
}

func (s *InactiveValidatorsSuite) getConsumerID() string {
	consumerIDJSON, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q).consumer_id", s.Consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	consumerID := consumerIDJSON.String()
	return consumerID
}

func TestInactiveValidators(t *testing.T) {
	s := &InactiveValidatorsSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
	}
	suite.Run(t, s)
}
