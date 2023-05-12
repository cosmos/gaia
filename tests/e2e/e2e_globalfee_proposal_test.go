package e2e

import (
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (s *IntegrationTestSuite) govProposeNewGlobalfee(newGlobalfee sdk.DecCoins, proposalCounter int, submitter string, _ string) {
	s.writeGovParamChangeProposalGlobalFees(s.chainA, newGlobalfee)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	submitGovFlags := []string{"param-change", configFile(proposalGlobalFeeFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing new fees
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fee to %s", newGlobalfee.String())
	s.runGovProcess(chainAAPIEndpoint, submitter, proposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

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

func (s *IntegrationTestSuite) govProposeNewBypassMsgs(newBypassMsgs []string, proposalCounter int, submitter string, fees string) { //nolint:unparam
	s.writeGovParamChangeProposalBypassMsgs(s.chainA, newBypassMsgs)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	submitGovFlags := []string{"param-change", configFile(proposalBypassMsgFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing new fees
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change bypass min fee msg types to %s", newBypassMsgs)
	s.runGovProcess(chainAAPIEndpoint, submitter, proposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

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
			bypassMsgs, err := queryBypassMsgs(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", newBypassMsgs)
			s.Require().NoError(err)

			// attention: if global fee is empty, when query globalfee, it shows empty rather than default ante.DefaultZeroGlobalFee() = 0uatom.
			s.Require().Equal(newBypassMsgs, bypassMsgs)
			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) govProposeNewMaxTotalBypassMinFeeMsgGasUsage(newGas uint64, proposalCounter int, submitter string) {
	s.writeGovParamChangeProposalMaxTotalBypass(s.chainA, newGas)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	submitGovFlags := []string{"param-change", configFile(proposalMaxTotalBypassFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing new max gas usage for bypass msgs
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change maxTotalBypassMinFeeMsgGasUsage to %d", newGas)
	s.runGovProcess(chainAAPIEndpoint, submitter, proposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

	// query the proposal status and max gas usage for bypass msgs
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
			gas, err := queryMaxTotalBypassMinFeeMsgGasUsage(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %d", gas)
			s.Require().NoError(err)

			s.Require().Equal(newGas, gas)
			return true
		},
		15*time.Second,
		5*time.Second,
	)
}
