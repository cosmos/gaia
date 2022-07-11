package e2e

import (
	"fmt"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
	// TODO: Remove skip once IBC is reintegrated
	s.T().Skip()
	var ibcStakeDenom string

	s.Run("send_photon_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances sdk.Coins
			err      error
		)

		address, err := s.chainB.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := address.String()
		s.sendIBC(s.chainA.id, s.chainB.id, recipient, tokenAmount)

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = queryGaiaAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)

				return balances.Len() == 3
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal(tokenAmount.Amount.Int64(), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

func (s *IntegrationTestSuite) TestBankTokenTransfer() {
	s.Run("send_photon_between_accounts", func() {
		var err error

		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		recipientAddress, err := s.chainA.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := recipientAddress.String()

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderPhotonBalance    sdk.Coin
			beforeRecipientPhotonBalance sdk.Coin
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderPhotonBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, "photon")
				s.Require().NoError(err)

				beforeRecipientPhotonBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, "photon")
				s.Require().NoError(err)

				return beforeSenderPhotonBalance.IsValid() && beforeRecipientPhotonBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		s.sendMsgSend(s.chainA, 0, sender, recipient, tokenAmount.String(), fees.String())

		s.Require().Eventually(
			func() bool {
				afterSenderPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, "photon")
				s.Require().NoError(err)

				afterRecipientPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, "photon")
				s.Require().NoError(err)

				decremented := beforeSenderPhotonBalance.Sub(tokenAmount).Sub(fees).IsEqual(afterSenderPhotonBalance)
				incremented := beforeRecipientPhotonBalance.Add(tokenAmount).IsEqual(afterRecipientPhotonBalance)

				return decremented && incremented
			},
			time.Minute,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) TestSendTokensFromNewGovAccount() {
	s.writeGovProposals((s.chainA))
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)

	s.fundCommunityPool(chainAAPIEndpoint, sender)

	s.T().Logf("Submitting Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.submitLegacyProposalFundGovAccount(chainAAPIEndpoint, sender, proposalCounter)
	s.T().Logf("Depositing Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.depositGovProposal(chainAAPIEndpoint, sender, proposalCounter)
	s.T().Logf("Voting Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalCounter, "yes", false)

	initialGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
	s.Require().NoError(err)
	proposalCounter++

	s.T().Logf("Submitting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_2.json")
	s.T().Logf("Depositing Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.depositGovProposal(chainAAPIEndpoint, sender, proposalCounter)
	s.T().Logf("Voting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalCounter, "yes", false)
	s.Require().Eventually(
		func() bool {
			newGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
			s.Require().NoError(err)

			recipientBalance, err := getSpecificBalance(chainAAPIEndpoint, govSendMsgRecipientAddress, photonDenom)
			s.Require().NoError(err)
			return newGovBalance.IsEqual(initialGovBalance.Sub(sendGovAmount)) && recipientBalance.Equal(initialGovBalance.Sub(newGovBalance))
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) TestGovSoftwareUpgrade() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + govProposalBlockBuffer
	proposalCounter++

	s.T().Logf("Writing proposal %d on chain %s", proposalCounter, s.chainA.id)
	s.writeGovUpgradeSoftwareProposal(s.chainA, proposalHeight)

	s.T().Logf("Submitting Gov Proposal: Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_3.json")
	s.T().Logf("Depositing Gov Proposal: Software Upgrade")
	s.depositGovProposal(chainAAPIEndpoint, sender, proposalCounter)
	s.T().Logf("Weighted Voting Gov Proposal: Software Upgrade")
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalCounter, "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05", true)

	s.verifyChainHaltedAtUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully halted chain at height %d", proposalHeight)

	currentChain := s.chainA

	for valIdx := range currentChain.validators {
		var opts docker.RemoveContainerOptions
		opts.ID = s.valResources[currentChain.id][valIdx].Container.ID
		opts.Force = true
		s.dkrPool.Client.RemoveContainer(opts)
		s.T().Logf("Removed Container: %s", s.valResources[currentChain.id][valIdx].Container.Name[1:])
	}

	s.T().Logf("Restarting containers")
	s.SetupSuite()

	s.Require().Eventually(
		func() bool {
			h := s.getLatestBlockHeight(s.chainA, 0)
			s.Require().NoError(err)

			return (h > 0)
		},
		30*time.Second,
		5*time.Second,
	)

	proposalCounter = 0
}

func (s *IntegrationTestSuite) TestGovCancelSoftwareUpgrade() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()
	height := s.getLatestBlockHeight(s.chainA, 0)
	proposalHeight := height + 50
	proposalCounter++

	s.T().Logf("Writing proposal %d on chain %s", proposalCounter, s.chainA.id)
	s.writeGovUpgradeSoftwareProposal(s.chainA, proposalHeight)

	s.T().Logf("Submitting Gov Proposal: Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_3.json")
	s.depositGovProposal(chainAAPIEndpoint, sender, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalCounter, "yes", false)

	proposalCounter++

	s.T().Logf("Submitting Gov Proposal: Cancel Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_4.json")
	s.depositGovProposal(chainAAPIEndpoint, sender, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalCounter, "yes", false)

	s.verifyChainPassesUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully canceled upgrade at height %d", proposalHeight)
}

func (s *IntegrationTestSuite) fundCommunityPool(chainAAPIEndpoint string, sender string) {
	s.Run("fund_community_pool", func() {
		beforeDistPhotonBalance, _ := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
		if beforeDistPhotonBalance.IsNil() {
			// Set balance to 0 if previous balance does not exist
			beforeDistPhotonBalance = sdk.NewInt64Coin("photon", 0)
		}

		s.execDistributionFundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())

		s.Require().Eventually(
			func() bool {
				afterDistPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
				if err != nil {
					s.T().Logf("Error getting balance: %s", afterDistPhotonBalance)
				}
				s.Require().NoError(err)

				return afterDistPhotonBalance.IsEqual(beforeDistPhotonBalance.Add(tokenAmount.Add(fees)))
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitLegacyProposalFundGovAccount(chainAAPIEndpoint string, sender string, proposalId int) {
	s.Run("submit_legacy_community_spend_proposal_to_fund_gov_acct", func() {
		s.execGovSubmitLegacyGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, "/root/.gaia/config/proposal.json", fees.String(), "community-pool-spend")

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusDepositPeriod)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitNewGovProposal(chainAAPIEndpoint string, sender string, proposalId int, proposalPath string) {
	s.Run("submit_new_gov_proposal", func() {
		s.execGovSubmitProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalPath, fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.T().Logf("Proposal: %s", proposal.String())
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusDepositPeriod)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) depositGovProposal(chainAAPIEndpoint string, sender string, proposalId int) {
	s.Run("deposit_gov_proposal", func() {
		s.execGovDepositProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, depositAmount, fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusVotingPeriod)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) voteGovProposal(chainAAPIEndpoint string, sender string, proposalId int, vote string, weighted bool) {
	s.Run("vote_gov_proposal", func() {
		if weighted {
			s.execGovWeightedVoteProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, vote, fees.String())
		} else {
			s.execGovVoteProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, vote, fees.String())
		}

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposalId)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) verifyChainHaltedAtUpgradeHeight(c *chain, valIdx int, upgradeHeight int) {
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

func (s *IntegrationTestSuite) verifyChainPassesUpgradeHeight(c *chain, valIdx int, upgradeHeight int) {
	s.Require().Eventually(
		func() bool {
			currentHeight := s.getLatestBlockHeight(c, valIdx)

			return currentHeight > upgradeHeight
		},
		30*time.Second,
		5*time.Second,
	)
}
