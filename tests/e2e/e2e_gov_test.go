package e2e

import (
	"fmt"
	"strconv"
	"time"

	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
	"github.com/cosmos/gaia/v25/tests/e2e/msg"
	"github.com/cosmos/gaia/v25/tests/e2e/query"
)

/*
GovSoftwareUpgrade tests passing a gov proposal to upgrade the chain at a given height.
Test Benchmarks:
1. Submission, deposit and vote of message based proposal to upgrade the chain at a height (current height + buffer)
2. Validation that chain halted at upgrade height
3. Teardown & restart chains
4. Reset proposalCounter so subsequent tests have the correct last effective proposal id for chainA
TODO: Perform upgrade in place of chain restart
*/
func (s *IntegrationTestSuite) GovSoftwareUpgrade() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	sender := senderAddress.String()
	height, err := query.GetLatestBlockHeight(chainAAPIEndpoint)
	s.Require().NoError(err)
	proposalHeight := height + common.GovProposalBlockBuffer
	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	s.TestCounters.ProposalCounter++

	err = msg.WriteSoftwareUpgradeProposal(s.Resources.ChainA, int64(proposalHeight), "upgrade-v0")
	s.Require().NoError(err)

	submitGovFlags := []string{configFile(common.ProposalSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05"}
	s.submitGovProposal(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "weighted-vote")

	s.verifyChainHaltedAtUpgradeHeight(chainAAPIEndpoint, proposalHeight)
	s.T().Logf("Successfully halted chain at  height %d", proposalHeight)

	s.TearDownSuite()

	s.T().Logf("Restarting containers")
	s.SetupSuite()

	s.Require().Eventually(
		func() bool {
			h, err := query.GetLatestBlockHeight(chainAAPIEndpoint)
			s.Require().NoError(err)
			return h > 0
		},
		30*time.Second,
		5*time.Second,
	)

	s.TestCounters.ProposalCounter = 0
}

/*
GovCancelSoftwareUpgrade tests passing a gov proposal that cancels a pending upgrade.
Test Benchmarks:
1. Submission, deposit and vote of message based proposal to upgrade the chain at a height (current height + buffer)
2. Submission, deposit and vote of message based proposal to cancel the pending upgrade
3. Validation that the chain produced blocks past the intended upgrade height
*/
func (s *IntegrationTestSuite) GovCancelSoftwareUpgrade() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()

	sender := senderAddress.String()
	height, err := query.GetLatestBlockHeight(chainAAPIEndpoint)
	s.Require().NoError(err)
	proposalHeight := height + 50
	err = msg.WriteSoftwareUpgradeProposal(s.Resources.ChainA, int64(proposalHeight), "upgrade-v1")
	s.Require().NoError(err)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.TestCounters.ProposalCounter++
	err = msg.WriteCancelSoftwareUpgradeProposal(s.Resources.ChainA)
	s.Require().NoError(err)
	submitGovFlags = []string{configFile(common.ProposalCancelSoftwareUpgrade)}
	depositGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, upgradetypes.ProposalTypeCancelSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.verifyChainPassesUpgradeHeight(chainAAPIEndpoint, proposalHeight)
	s.T().Logf("Successfully canceled upgrade at height %d", proposalHeight)
}

/*
GovCommunityPoolSpend tests passing a community spend proposal.
Test Benchmarks:
1. Fund Community Pool
2. Submission, deposit and vote of proposal to spend from the community pool to send atoms to a recipient
3. Validation that the recipient balance has increased by proposal amount
*/
func (s *IntegrationTestSuite) GovCommunityPoolSpend() {
	s.fundCommunityPool()
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	sender := senderAddress.String()
	recipientAddress, _ := s.Resources.ChainA.Validators[1].KeyInfo.GetAddress()
	recipient := recipientAddress.String()
	sendAmount := sdk.NewCoin(common.UAtomDenom, math.NewInt(10000000)) // 10uatom
	err := msg.WriteGovCommunitySpendProposal(s.Resources.ChainA, sendAmount, recipient)
	s.Require().NoError(err)

	beforeRecipientBalance, err := query.SpecificBalance(chainAAPIEndpoint, recipient, common.UAtomDenom)
	s.Require().NoError(err)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalCommunitySpendFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, "CommunityPoolSpend", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			afterRecipientBalance, err := query.SpecificBalance(chainAAPIEndpoint, recipient, common.UAtomDenom)
			s.Require().NoError(err)

			return afterRecipientBalance.Sub(sendAmount).IsEqual(beforeRecipientBalance)
		},
		10*time.Second,
		5*time.Second,
	)
}

// NOTE: in SDK >= v0.47 the submit-proposal does not have a --deposit flag
// Instead, the depoist is added to the "deposit" field of the proposal JSON (usually stored as a file)
// you can use `gaiad tx gov draft-proposal` to create a proposal file that you can use
// min initial deposit of 100uatom is required in e2e tests, otherwise the proposal would be dropped
func (s *IntegrationTestSuite) submitGovProposal(chainAAPIEndpoint, sender string, proposalID int, proposalType string, submitFlags []string, depositFlags []string, voteFlags []string, voteCommand string) {
	s.T().Logf("Submitting Gov Proposal: %s", proposalType)
	sflags := submitFlags
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, "submit-proposal", sflags, govtypesv1beta1.StatusDepositPeriod)
	s.T().Logf("Depositing Gov Proposal: %s", proposalType)
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, "deposit", depositFlags, govtypesv1beta1.StatusVotingPeriod)
	s.T().Logf("Voting Gov Proposal: %s", proposalType)
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, voteCommand, voteFlags, govtypesv1beta1.StatusPassed)
}

func (s *IntegrationTestSuite) verifyChainHaltedAtUpgradeHeight(endpoint string, upgradeHeight int) {
	s.Require().Eventually(
		func() bool {
			currentHeight, err := query.GetLatestBlockHeight(endpoint)
			s.Require().NoError(err)

			return currentHeight == upgradeHeight
		},
		30*time.Second,
		5*time.Second,
	)

	counter := 0
	s.Require().Eventually(
		func() bool {
			currentHeight, err := query.GetLatestBlockHeight(endpoint)
			s.Require().NoError(err)

			if currentHeight > upgradeHeight {
				return false
			}
			if currentHeight == upgradeHeight {
				counter++
			}
			return counter >= 2
		},
		8*time.Second,
		2*time.Second,
	)
}

func (s *IntegrationTestSuite) verifyChainPassesUpgradeHeight(endpoint string, upgradeHeight int) {
	s.Require().Eventually(
		func() bool {
			currentHeight, err := query.GetLatestBlockHeight(endpoint)
			s.Require().NoError(err)

			return currentHeight > upgradeHeight
		},
		30*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) submitGovCommand(chainAAPIEndpoint, sender string, proposalID int, govCommand string, proposalFlags []string, expectedSuccessStatus govtypesv1beta1.ProposalStatus) {
	s.Run(fmt.Sprintf("Running tx gov %s", govCommand), func() {
		s.RunGovExec(s.Resources.ChainA, 0, sender, govCommand, proposalFlags, common.StandardFees.String(), nil)

		s.Require().Eventually(
			func() bool {
				proposal, err := query.GovProposal(chainAAPIEndpoint, proposalID)
				s.Require().NoError(err)
				return proposal.GetProposal().Status == expectedSuccessStatus
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

// testSetBlocksPerEpoch tests that we can change `BlocksPerEpoch` through a governance proposal
func (s *IntegrationTestSuite) testSetBlocksPerEpoch() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	providerParams := providertypes.DefaultParams()

	// assert that initially, the actual blocks per epoch are the default blocks per epoch
	s.Require().Eventually(
		func() bool {
			blocksPerEpoch, err := query.BlocksPerEpoch(chainEndpoint)
			s.T().Logf("Initial BlocksPerEpoch param: %v", blocksPerEpoch)
			s.Require().NoError(err)

			s.Require().Equal(blocksPerEpoch, providerParams.BlocksPerEpoch)
			return true
		},
		15*time.Second,
		5*time.Second,
	)

	// create a governance proposal to change blocks per epoch to the default blocks per epoch plus one
	expectedBlocksPerEpoch := providerParams.BlocksPerEpoch + 1
	providerParams.BlocksPerEpoch = expectedBlocksPerEpoch
	paramsJSON := common.Cdc.MustMarshalJSON(&providerParams)
	err := msg.WriteGovParamChangeProposalBlocksPerEpoch(s.Resources.ChainA, string(paramsJSON))
	s.Require().NoError(err)

	validatorAAddr, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalBlocksPerEpochFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Change BlocksPerEpoch parameter")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			blocksPerEpoch, err := query.BlocksPerEpoch(chainEndpoint)
			s.Require().NoError(err)

			s.T().Logf("Newly set blocks per epoch: %d", blocksPerEpoch)
			s.Require().Equal(expectedBlocksPerEpoch, blocksPerEpoch)
			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

// GovSoftwareUpgradeExpedited can be expedited but it can only be submitted using "tx gov submit-proposal" command.
// Messages submitted using "tx gov submit-legacy-proposal" command cannot be expedited.// submit but vote no so that the proposal is not passed
func (s *IntegrationTestSuite) GovSoftwareUpgradeExpedited() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	sender := senderAddress.String()

	s.TestCounters.ProposalCounter++
	err := msg.WriteExpeditedSoftwareUpgradeProp(s.Resources.ChainA)
	s.Require().NoError(err)
	submitGovFlags := []string{configFile(common.ProposalExpeditedSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes=0.1,no=0.8,abstain=0.05,no_with_veto=0.05"}

	s.Run(fmt.Sprintf("Running expedited tx gov %s", "submit-proposal"), func() {
		s.submitGovCommand(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, "submit-proposal", submitGovFlags, govtypesv1beta1.StatusDepositPeriod)

		s.Require().Eventually(
			func() bool {
				proposal, err := query.GovProposalV1(chainAAPIEndpoint, s.TestCounters.ProposalCounter)
				s.Require().NoError(err)
				return proposal.Proposal.Expedited && proposal.GetProposal().Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
			},
			15*time.Second,
			5*time.Second,
		)
		s.submitGovCommand(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, "deposit", depositGovFlags, govtypesv1beta1.StatusVotingPeriod)
		s.submitGovCommand(chainAAPIEndpoint, sender, s.TestCounters.ProposalCounter, "weighted-vote", voteGovFlags, govtypesv1beta1.StatusRejected) // voting no on prop

		// confirm that the proposal was moved from expedited
		s.Require().Eventually(
			func() bool {
				proposal, err := query.GovProposalV1(chainAAPIEndpoint, s.TestCounters.ProposalCounter)
				s.Require().NoError(err)
				return proposal.Proposal.Expedited == false
			},
			15*time.Second,
			5*time.Second,
		)
	})
}
