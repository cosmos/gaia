package delegator_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/sjson"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
)

const (
	distributionStakeAmount          = "10000000" // 10 ATOM
	distributiongovSubmissionDeposit = "100"
	distributionproposalDepositInt   = chainsuite.GovMinDepositAmount
)

type DistributionSuite struct {
	*delegator.Suite
}

func (s *DistributionSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Delegate >1 ATOM with delegator account
	node := s.Chain.GetNode()
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
}

func (s *DistributionSuite) TestParamChange() {
	distributionParams, err := s.Chain.QueryJSON(s.GetContext(), "params", "distribution", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", distributionParams)
	currentCommunityTax := distributionParams.Get("community_tax").Float()
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Current community tax: %f", currentCommunityTax)
	newCommunityTax := currentCommunityTax - 0.01

	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	updatedParams, err := sjson.Set(distributionParams.String(), "community_tax", fmt.Sprintf("%f", newCommunityTax))
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Updated params: %s", updatedParams)

	paramChangeMessage := fmt.Sprintf(`{
    	"@type": "/cosmos.distribution.v1beta1.MsgUpdateParams",
		"authority": "%s",
    	"params": %s
	}`, authority, updatedParams)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Message: %s", paramChangeMessage)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Distribution Param Change Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.KeyName(), false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(paramChangeMessage)}
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID

	json, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "gov", "proposal", proposalId)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(json))

	// Pass proposal
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), proposalId))

	// Test
	distributionParams, err = s.Chain.QueryJSON(s.GetContext(), "params", "distribution", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", distributionParams)
	currentCommunityTax = distributionParams.Get("community_tax").Float()
	s.Require().Equal(newCommunityTax, currentCommunityTax)
}

func TestDistributionModule(t *testing.T) {
	s := &DistributionSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
