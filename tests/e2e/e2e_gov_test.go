package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ccvtypes "github.com/cosmos/interchain-security/x/ccv/provider/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
	senderAddress := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + govProposalBlockBuffer
	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	submitGovFlags := []string{"software-upgrade", "Upgrade-0", "--title='Upgrade V1'", "--description='Software Upgrade'", fmt.Sprintf("--upgrade-height=%d", proposalHeight)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "weighted-vote", true)

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
	senderAddress := s.chainA.validators[0].keyInfo.GetAddress()

	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + 50
	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	submitGovFlags := []string{"software-upgrade", "Upgrade-1", "--title='Upgrade V1'", "--description='Software Upgrade'", fmt.Sprintf("--upgrade-height=%d", proposalHeight)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote", true)

	proposalCounter++
	submitGovFlags = []string{"cancel-software-upgrade", "--title='Upgrade V1'", "--description='Software Upgrade'"}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, upgradetypes.ProposalTypeCancelSoftwareUpgrade, submitGovFlags, depositGovFlags, voteGovFlags, "vote", true)

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
	senderAddress := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()
	recipientAddress := s.chainA.validators[1].keyInfo.GetAddress()
	recipient := recipientAddress.String()
	sendAmount := sdk.NewCoin(uatomDenom, sdk.NewInt(10000000)) // 10uatom
	s.writeGovCommunitySpendProposal(s.chainA, sendAmount.String(), recipient)

	beforeRecipientBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, uatomDenom)
	s.Require().NoError(err)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	submitGovFlags := []string{"community-pool-spend", configFile(proposalCommunitySpendFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, distrtypes.ProposalTypeCommunityPoolSpend, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

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

/*
AddRemoveConsumerChain tests adding and subsequently removing a new consumer chain to Gaia.
Test Benchmarks:
1. Submit and pass proposal to add consumer chain
2. Validation that consumer chain was added
3. Submit and pass proposal to remove consumer chain
4. Validation that consumer chain was removed
*/
func (s *IntegrationTestSuite) AddRemoveConsumerChain() {
	s.fundCommunityPool()
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	proposerAddress := s.chainA.validators[0].keyInfo.GetAddress()
	sender := proposerAddress.String()
	consumerChainID := "consumer"
	s.writeAddRemoveConsumerProposals(s.chainA, consumerChainID)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	// Add Consumer Chain
	proposalCounter++
	submitGovFlags := []string{"consumer-addition", configFile(proposalAddConsumerChainFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, ccvtypes.ProposalTypeConsumerAddition, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

	// Query and assert consumer has been added
	s.execQueryConsumerChains(s.chainA, 0, gaiaHomePath, validateConsumerAddition, consumerChainID)

	// Remove Consumer Chain
	proposalCounter++
	submitGovFlags = []string{"consumer-removal", configFile(proposalRemoveConsumerChainFilename)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}
	s.runGovProcess(chainAAPIEndpoint, sender, proposalCounter, ccvtypes.ProposalTypeConsumerRemoval, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)
	// Query and assert consumer has been removed
	s.execQueryConsumerChains(s.chainA, 0, gaiaHomePath, validateConsumerRemoval, consumerChainID)
}

func validateConsumerAddition(res ccvtypes.QueryConsumerChainsResponse, consumerChainID string) bool {
	if res.Size() == 0 {
		return false
	}
	for _, chain := range res.GetChains() {
		return strings.Compare(chain.ChainId, consumerChainID) == 0
	}
	return false
}

func validateConsumerRemoval(res ccvtypes.QueryConsumerChainsResponse, consumerChainID string) bool {
	if res.Size() > 0 {
		for _, chain := range res.GetChains() {
			if strings.Compare(chain.ChainId, consumerChainID) == 0 {
				return false
			}
		}
	}
	return true
}

func (s *IntegrationTestSuite) runGovProcess(chainAAPIEndpoint, sender string, proposalID int, proposalType string, submitFlags []string, depositFlags []string, voteFlags []string, voteCommand string, withDeposit bool) {
	s.T().Logf("Submitting Gov Proposal: %s", proposalType)
	// min deposit of 1000uatom is required in e2e tests, otherwise the gov antehandler causes the proposal to be dropped
	sflags := submitFlags
	if withDeposit {
		sflags = append(sflags, "--deposit=1000uatom")
	}
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, "submit-proposal", sflags, govtypes.StatusDepositPeriod)
	s.T().Logf("Depositing Gov Proposal: %s", proposalType)
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, "deposit", depositFlags, govtypes.StatusVotingPeriod)
	s.T().Logf("Voting Gov Proposal: %s", proposalType)
	s.submitGovCommand(chainAAPIEndpoint, sender, proposalID, voteCommand, voteFlags, govtypes.StatusPassed)
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

func (s *IntegrationTestSuite) submitGovCommand(chainAAPIEndpoint, sender string, proposalID int, govCommand string, proposalFlags []string, expectedSuccessStatus govtypes.ProposalStatus) {
	s.Run(fmt.Sprintf("Running tx gov %s", govCommand), func() {
		s.runGovExec(s.chainA, 0, sender, govCommand, proposalFlags, standardFees.String())

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
