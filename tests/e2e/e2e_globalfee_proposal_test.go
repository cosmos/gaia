package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func (s *IntegrationTestSuite) govProposeNewGlobalfee(newGlobalfee sdk.DecCoins, proposalCounter int, submitter string, fees string) {
	s.writeGovParamChangeProposalGlobalFees(s.chainA, newGlobalfee)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	// gov proposing new fees
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fee to %s", newGlobalfee.String())
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, fees, "param-change", proposalCounter, configFile(proposalGlobalFee))
	s.depositGovProposal(chainAAPIEndpoint, submitter, fees, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, fees, proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == gov.StatusPassed
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
