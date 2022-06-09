package e2e

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *IntegrationTestSuite) TestAIBCTokenTransfer() {
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

func (s *IntegrationTestSuite) TestBBankTokenTransfer() {
	s.Run("send_photon_between_accounts", func() {
		var (
			err error
		)

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
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()

	s.fundCommunityPool(chainAAPIEndpoint, sender)
	s.submitLegacyProposalFundGovAccount(chainAAPIEndpoint, sender)
	s.depositLegacyProposalFundGovAccount(chainAAPIEndpoint, sender)
	s.voteLegacyProposalFundGovAccount(chainAAPIEndpoint, sender, proposal1Id, "yes")

	initialGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
	s.Require().NoError(err)

	s.T().Logf("Submitting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposal2Id, "/root/.gaia/config/proposal_2.json")
	s.depositNewGovProposal(chainAAPIEndpoint, sender, proposal2Id)
	s.voteGovProposal(chainAAPIEndpoint, sender, proposal2Id, "yes")

	s.Run("new_msg_send_from_gov_proposal_successfully_funds_recipient_account", func() {
		s.Require().Eventually(
			func() bool {
				sendGovAmount := sdk.NewInt64Coin(photonDenom, 10)

				newGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
				s.Require().NoError(err)

				recipientBalance, err := getSpecificBalance(chainAAPIEndpoint, govSendMsgRecipientAddress, photonDenom)
				s.Require().NoError(err)

				return newGovBalance.IsEqual(initialGovBalance.Sub(sendGovAmount)) && recipientBalance.Equal(initialGovBalance.Sub(newGovBalance))
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) TestGovUpgrade() {

}

func (s *IntegrationTestSuite) fundCommunityPool(chainAAPIEndpoint string, sender string) {
	s.Run("fund_community_pool", func() {
		s.execDistributionFundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())

		s.Require().Eventually(
			func() bool {
				afterDistPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
				if err != nil {
					s.T().Logf("Error getting balance: %s", afterDistPhotonBalance)
				}
				s.Require().NoError(err)

				return afterDistPhotonBalance.IsValid()
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitLegacyProposalFundGovAccount(chainAAPIEndpoint string, sender string) {
	s.Run("submit_legacy_community_spend_proposal_to_fund_gov_acct", func() {
		s.execGovSubmitLegacyGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, "/root/.gaia/config/proposal.json", fees.String(), "community-pool-spend")

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, 1)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusDepositPeriod)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) depositLegacyProposalFundGovAccount(chainAAPIEndpoint string, sender string) {
	s.Run("submit_legacy_community_spend_proposal_to_fund_gov_acct", func() {
		s.execGovDepositProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposal1Id, depositAmount, fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposal1Id)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusVotingPeriod)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) voteLegacyProposalFundGovAccount(chainAAPIEndpoint string, sender string, proposalId uint64, vote string) {
	s.voteGovProposal(chainAAPIEndpoint, sender, proposalId, vote)
	s.Run("verify_legacy_community_spend_funds_gov_successfully", func() {
		s.Require().Eventually(
			func() bool {
				govBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
				s.Require().NoError(err)

				return (govBalance.IsEqual(sdk.NewInt64Coin(photonDenom, 1000)))
			},
			25*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitNewGovProposal(chainAAPIEndpoint string, sender string, proposalId uint64, proposalPath string) {
	s.Run("submit_new_gov_proposal", func() {
		s.execGovSubmitProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalPath, fees.String())

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

func (s *IntegrationTestSuite) depositNewGovProposal(chainAAPIEndpoint string, sender string, proposalId uint64) {
	s.Run("deposit_new_gov_proposal", func() {
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

func (s *IntegrationTestSuite) voteGovProposal(chainAAPIEndpoint string, sender string, proposalId uint64, vote string) {
	s.Run("vote_new_msg_send_from_gov_proposal", func() {
		s.execGovVoteProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, vote, fees.String())

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
