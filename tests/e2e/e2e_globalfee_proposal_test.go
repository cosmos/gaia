package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *IntegrationTestSuite) govProposeNewGlobalfee(newGlobalfee sdk.DecCoins, proposalCounter int, submitter string, fees string) {
	s.writeGovParamChangeProposalGlobalFees(s.chainA, newGlobalfee)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	// gov proposing new fees
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fee to %s", newGlobalfee.String())
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, fees, "param-change", proposalCounter, configFile("proposal_globalfee.json"))
	s.depositGovProposal(chainAAPIEndpoint, submitter, fees, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, fees, proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)

			// attention: if global fee is empty, when query globalfee, it shows empty rather than default ante.DefaultZeroGlobalFee() = 0uatom.
			return globalFees.IsEqual(newGlobalfee)
		},
		15*time.Second,
		5*time.Second,
	)
}
