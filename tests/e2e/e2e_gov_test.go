package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	// govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

/*
SendTokensFromNewGovAccount tests passing a gov proposal that sends tokens on behalf of the gov module to a recipient.
Test Benchmarks:
1. Subtest to fund the community pool via the distribution module
2. Submission, deposit and vote of legacy proposal to fund the gov account from the community pool
3. Validation that funds have been deposited to gov account
4. Submission, deposit and vote of message based proposal to send funds from the gov account to a recipient account
5. Validation that funds have been deposited to recipient account
*/
func (s *IntegrationTestSuite) SendTokensFromNewGovAccount() {
	s.writeGovProposals(s.chainA)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()
	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)

	s.fundCommunityPool(chainAAPIEndpoint, sender)

	s.T().Logf("Submitting Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.submitLegacyGovProposal(chainAAPIEndpoint, sender, standardFees.String(), "community-pool-spend", proposalCounter, configFile("proposal.json"))
	s.T().Logf("Depositing Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.depositGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter)
	s.T().Logf("Voting Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.voteGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter, "yes", false)

	initialGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, uatomDenom)
	s.Require().NoError(err)
	proposalCounter++

	s.T().Logf("Submitting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, configFile("proposal_2.json"))
	s.T().Logf("Depositing Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.depositGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter)
	s.T().Logf("Voting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.voteGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter, "yes", false)
	s.Require().Eventually(
		func() bool {
			newGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, uatomDenom)
			s.Require().NoError(err)

			recipientBalance, err := getSpecificBalance(chainAAPIEndpoint, govSendMsgRecipientAddress, uatomDenom)
			s.Require().NoError(err)
			return newGovBalance.IsEqual(initialGovBalance.Sub(sendGovAmount)) && recipientBalance.Equal(initialGovBalance.Sub(newGovBalance))
		},
		15*time.Second,
		5*time.Second,
	)
}

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

	s.T().Logf("Writing proposal %d on chain %s", proposalCounter, s.chainA.id)
	s.writeGovUpgradeSoftwareProposal(s.chainA, proposalHeight)

	s.T().Logf("Submitting Gov Proposal: Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, configFile("proposal_3.json"))
	s.T().Logf("Depositing Gov Proposal: Software Upgrade")
	s.depositGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter)
	s.T().Logf("Weighted Voting Gov Proposal: Software Upgrade")
	s.voteGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter, "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05", true)

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

	s.T().Logf("Writing proposal %d on chain %s", proposalCounter, s.chainA.id)
	s.writeGovUpgradeSoftwareProposal(s.chainA, proposalHeight)

	s.T().Logf("Submitting Gov Proposal: Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, configFile("proposal_3.json"))
	s.depositGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter, "yes", false)

	proposalCounter++

	s.T().Logf("Writing proposal %d on chain %s", proposalCounter, s.chainA.id)
	s.writeGovCancelUpgradeSoftwareProposal(s.chainA)

	s.T().Logf("Submitting Gov Proposal: Cancel Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, configFile("proposal_4.json"))
	s.depositGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, standardFees.String(), proposalCounter, "yes", false)

	s.verifyChainPassesUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully canceled upgrade at height %d", proposalHeight)
}

/*
fundCommunityPool tests the funding of the community pool on behalf of the distribution module.
Test Benchmarks:
1. Validation that balance of the distribution module account before funding
2. Execution funding the community pool
3. Verification that correct funds have been deposited to distribution module account
*/
func (s *IntegrationTestSuite) fundCommunityPool(chainAAPIEndpoint, sender string) {
	s.Run("fund_community_pool", func() {
		beforeDistUatomBalance, _ := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
		if beforeDistUatomBalance.IsNil() {
			// Set balance to 0 if previous balance does not exist
			beforeDistUatomBalance = sdk.NewInt64Coin(uatomDenom, 0)
		}

		s.execDistributionFundCommunityPool(s.chainA, 0, sender, tokenAmount.String(), standardFees.String())

		// there are still tokens being added to the community pool through block production rewards but they should be less than 500 tokens
		marginOfErrorForBlockReward := sdk.NewInt64Coin(uatomDenom, 500)

		s.Require().Eventually(
			func() bool {
				afterDistPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
				s.Require().NoErrorf(err, "Error getting balance: %s", afterDistPhotonBalance)

				return afterDistPhotonBalance.Sub(beforeDistUatomBalance.Add(tokenAmount.Add(standardFees))).IsLT(marginOfErrorForBlockReward)
			},
			15*time.Second,
			5*time.Second,
		)
	})
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

func (s *IntegrationTestSuite) submitLegacyGovProposal(chainAAPIEndpoint string, sender string, fees string, proposalTypeSubCmd string, proposalId int, proposalPath string) {
	s.Run("submit_legacy_gov_proposal", func() {
		s.execGovSubmitLegacyGovProposal(s.chainA, 0, sender, proposalPath, fees, proposalTypeSubCmd)

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)
				return proposal.GetProposal().Status == gov.StatusDepositPeriod
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitNewGovProposal(chainAAPIEndpoint, sender string, proposalId int, proposalPath string) {
	s.Run("submit_new_gov_proposal", func() {
		s.execGovSubmitProposal(s.chainA, 0, sender, proposalPath, standardFees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.T().Logf("Proposal: %s", proposal.String())
				s.Require().NoError(err)

				return proposal.GetProposal().Status == gov.StatusDepositPeriod
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) depositGovProposal(chainAAPIEndpoint, sender string, fees string, proposalId int) {
	s.Run("deposit_gov_proposal", func() {
		s.execGovDepositProposal(s.chainA, 0, sender, proposalId, depositAmount.String(), fees)

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)

				return proposal.GetProposal().Status == gov.StatusVotingPeriod
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) voteGovProposal(chainAAPIEndpoint, sender string, fees string, proposalId int, vote string, weighted bool) {
	s.Run("vote_gov_proposal", func() {
		if weighted {
			s.execGovWeightedVoteProposal(s.chainA, 0, sender, proposalId, vote, fees)
		} else {
			s.execGovVoteProposal(s.chainA, 0, sender, proposalId, vote, fees)
		}

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)

				return proposal.GetProposal().Status == gov.StatusPassed
			},
			15*time.Second,
			5*time.Second,
		)
	})
}
