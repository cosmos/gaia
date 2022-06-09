package e2e

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	proposal1Id = uint64(1)
	proposal2Id = uint64(2)
)

var (
	tokenAmount       = sdk.NewInt64Coin(photonDenom, 3300000000) // 3,300photon
	fees              = sdk.NewInt64Coin(photonDenom, 330000)
	depositAmount     = sdk.NewInt64Coin(photonDenom, 10000000).String()
	distModuleAddress = authtypes.NewModuleAddress(distrtypes.ModuleName).String()
	govModuleAddress  = authtypes.NewModuleAddress(govtypes.ModuleName).String()
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

func (s *IntegrationTestSuite) TestCDistributionFundCommunityPool() {
	s.Run("fund_community_pool", func() {
		s.T().Skip()
		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		s.fundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())

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

func (s *IntegrationTestSuite) TestDSubmitLegacyCommunitySpendFundGovProposal() {
	s.Run("submit_deposit_legacy_community_spend_to_fund_gov_proposal", func() {
		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		s.fundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())
		s.submitLegacyGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, "/root/.gaia/config/proposal.json", fees.String(), "community-pool-spend")

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, 1)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusDepositPeriod)
			},
			15*time.Second,
			5*time.Second,
		)

		s.depositGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposal1Id, depositAmount, fees.String())

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

func (s *IntegrationTestSuite) TestEVoteLegacyCommunitySpendProposal() {
	s.Run("vote_legacy_community_spend_to_fund_gov_proposal", func() {
		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		s.voteGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposal1Id, "yes", fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposal1Id)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
			},
			25*time.Second,
			5*time.Second,
		)

		s.Require().Eventually(
			func() bool {
				govBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, "photon")
				s.Require().NoError(err)

				return (govBalance.IsEqual(sdk.NewInt64Coin(photonDenom, 1000)))
			},
			25*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) TestFSubmitMsgSendFromGovProposal() {
	s.Run("submit_new_msg_send_from_gov_proposal", func() {
		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		s.submitGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, "/root/.gaia/config/proposal_2.json", fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposal2Id)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusDepositPeriod)
			},
			15*time.Second,
			5*time.Second,
		)

		s.depositGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposal2Id, depositAmount, fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposal2Id)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusVotingPeriod)
			},
			15*time.Second,
			5*time.Second,
		)

	})
}

func (s *IntegrationTestSuite) TestGVoteMsgSendFromGovProposal() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()
	s.Run("vote_new_msg_send_from_gov_proposal", func() {
		s.voteGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, uint64(2), "yes", fees.String())

		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, proposal2Id)
				s.Require().NoError(err)

				return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
			},
			15*time.Second,
			5*time.Second,
		)

		s.Require().Eventually(
			func() bool {
				remainingGovBalance := sdk.NewInt64Coin(photonDenom, 990)
				newRecipientBalance := sdk.NewInt64Coin(photonDenom, 10)

				govBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
				s.Require().NoError(err)

				recipientBalance, err := getSpecificBalance(chainAAPIEndpoint, govSendMsgRecipientAddress, photonDenom)
				s.Require().NoError(err)

				return govBalance.IsEqual(remainingGovBalance) && recipientBalance.Equal(newRecipientBalance)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}
