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
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + govProposalBlockBuffer
	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++

	s.writeSoftwareUpgradeProposal(s.chainA, int64(proposalHeight), "upgrade-v0")

	submitGovFlags := []string{configFile(proposalSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05"}
	s.submitGovProposal(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "weighted-vote")

	s.verifyChainHaltedAtUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully halted chain at  height %d", proposalHeight)

	s.TearDownSuite()

	s.T().Logf("Restarting containers")
	s.SetupSuite()

	s.Require().Eventually(
		func() bool {
			h := s.getLatestBlockHeight(s.chainA, 0)
			return h > 0
		},
		30*time.Second,
		5*time.Second,
	)

	proposalCounter = 0
}

/*
GovCancelSoftwareUpgrade tests passing a gov proposal that cancels a pending upgrade.
Test Benchmarks:
1. Submission, deposit and vote of message based proposal to upgrade the chain at a height (current height + buffer)
2. Submission, deposit and vote of message based proposal to cancel the pending upgrade
3. Validation that the chain produced blocks past the intended upgrade height
*/
func (s *IntegrationTestSuite) GovCancelSoftwareUpgrade() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()

	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + 50
	s.writeSoftwareUpgradeProposal(s.chainA, int64(proposalHeight), "upgrade-v1")

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	submitGovFlags := []string{configFile(proposalSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	proposalCounter++
	s.writeCancelSoftwareUpgradeProposal(s.chainA)
	submitGovFlags = []string{configFile(proposalCancelSoftwareUpgrade)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeCancelSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.verifyChainPassesUpgradeHeight(s.chainA, 0, proposalHeight)
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
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()
	recipientAddress, _ := s.chainA.validators[1].keyInfo.GetAddress()
	recipient := recipientAddress.String()
	sendAmount := sdk.NewCoin(uatomDenom, math.NewInt(10000000)) // 10uatom
	s.writeGovCommunitySpendProposal(s.chainA, sendAmount, recipient)

	beforeRecipientBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, uatomDenom)
	s.Require().NoError(err)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	submitGovFlags := []string{configFile(proposalCommunitySpendFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.submitGovProposal(chainAAPIEndpoint, sender, proposalCounter, "CommunityPoolSpend", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			afterRecipientBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, uatomDenom)
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

func (s *IntegrationTestSuite) verifyChainHaltedAtUpgradeHeight(c *chain, valIdx, upgradeHeight int) {
	s.Require().Eventually(
		func() bool {
			currentHeight := s.getLatestBlockHeight(c, valIdx)

			return currentHeight == upgradeHeight
		},
		30*time.Second,
		5*time.Second,
	)

	counter := 0
	s.Require().Eventually(
		func() bool {
			currentHeight := s.getLatestBlockHeight(c, valIdx)

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

func (s *IntegrationTestSuite) verifyChainPassesUpgradeHeight(c *chain, valIdx, upgradeHeight int) {
	s.Require().Eventually(
		func() bool {
			currentHeight := s.getLatestBlockHeight(c, valIdx)

			return currentHeight > upgradeHeight
		},
		30*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) submitGovCommand(chainAAPIEndpoint, sender string, proposalID int, govCommand string, proposalFlags []string, expectedSuccessStatus govtypesv1beta1.ProposalStatus) {
	s.Run(fmt.Sprintf("Running tx gov %s", govCommand), func() {
		s.runGovExec(s.chainA, 0, sender, govCommand, proposalFlags, standardFees.String(), nil)

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalID)
				s.Require().NoError(err)
				return proposal.GetProposal().Status == expectedSuccessStatus
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitGovCommandExpectingFailure(sender string, govCommand string, proposalFlags []string) {
	s.Run(fmt.Sprintf("Running failing expedited tx gov %s -- expecting error", govCommand), func() {
		// should return an error -- the Tx fails at the ante handler
		s.runGovExec(s.chainA, 0, sender, govCommand, proposalFlags, standardFees.String(), s.expectTxSubmitError("unsupported expedited proposal type"))
	})
}

// testSetBlocksPerEpoch tests that we can change `BlocksPerEpoch` through a governance proposal
func (s *IntegrationTestSuite) testSetBlocksPerEpoch() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	providerParams := providertypes.DefaultParams()

	// assert that initially, the actual blocks per epoch are the default blocks per epoch
	s.Require().Eventually(
		func() bool {
			blocksPerEpoch, err := queryBlocksPerEpoch(chainEndpoint)
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
	paramsJSON := cdc.MustMarshalJSON(&providerParams)
	s.writeGovParamChangeProposalBlocksPerEpoch(s.chainA, string(paramsJSON))

	validatorAAddr, _ := s.chainA.validators[0].keyInfo.GetAddress()
	proposalCounter++
	submitGovFlags := []string{configFile(proposalBlocksPerEpochFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Change BlocksPerEpoch parameter")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, paramtypes.ProposalTypeChange, submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			blocksPerEpoch, err := queryBlocksPerEpoch(chainEndpoint)
			s.Require().NoError(err)

			s.T().Logf("Newly set blocks per epoch: %d", blocksPerEpoch)
			s.Require().Equal(expectedBlocksPerEpoch, blocksPerEpoch)
			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

// ExpeditedProposalRejected tests that expediting a ParamChange proposal fails.
func (s *IntegrationTestSuite) ExpeditedProposalRejected() {
	defaultBlocksPerEpoch := providertypes.DefaultParams().BlocksPerEpoch

	// attempt to change but nothing should happen -> proposal fails at ante handler
	expectedBlocksPerEpoch := defaultBlocksPerEpoch + 100
	s.writeFailingExpeditedProposal(s.chainA, expectedBlocksPerEpoch)

	validatorAAddr, _ := s.chainA.validators[0].keyInfo.GetAddress()
	submitGovFlags := []string{configFile(proposalFailExpedited)}

	s.T().Logf("Submitting, deposit and vote Gov Proposal: Change BlocksPerEpoch parameter - expecting to fail")
	s.submitGovCommandExpectingFailure(validatorAAddr.String(), "submit-proposal", submitGovFlags)
}

// MsgSoftwareUpgrade can be expedited but it can only be submitted using "tx gov submit-proposal" command.
// Messages submitted using "tx gov submit-legacy-proposal" command cannot be expedited.// submit but vote no so that the proposal is not passed
func (s *IntegrationTestSuite) GovSoftwareUpgradeExpedited() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()

	proposalCounter++
	s.writeExpeditedSoftwareUpgradeProp(s.chainA)
	submitGovFlags := []string{configFile(proposalExpeditedSoftwareUpgrade)}

	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes=0.1,no=0.8,abstain=0.05,no_with_veto=0.05"}

	s.Run(fmt.Sprintf("Running expedited tx gov %s", "submit-proposal"), func() {
		s.submitGovCommand(chainAAPIEndpoint, sender, proposalCounter, "submit-proposal", submitGovFlags, govtypesv1beta1.StatusDepositPeriod)

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposalV1(chainAAPIEndpoint, proposalCounter)
				s.Require().NoError(err)
				return proposal.Proposal.Expedited && proposal.GetProposal().Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
			},
			15*time.Second,
			5*time.Second,
		)
		s.submitGovCommand(chainAAPIEndpoint, sender, proposalCounter, "deposit", depositGovFlags, govtypesv1beta1.StatusVotingPeriod)
		s.submitGovCommand(chainAAPIEndpoint, sender, proposalCounter, "weighted-vote", voteGovFlags, govtypesv1beta1.StatusRejected) // voting no on prop

		// confirm that the proposal was moved from expedited
		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposalV1(chainAAPIEndpoint, proposalCounter)
				s.Require().NoError(err)
				return proposal.Proposal.Expedited == false
			},
			15*time.Second,
			5*time.Second,
		)
	})
}
