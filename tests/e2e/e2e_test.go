package e2e

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
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

		s.sendMsgSend(s.chainA, 0, sender, recipient, tokenAmount.String(), fees.String(), false)

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
	s.depositGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter)
	s.T().Logf("Voting Legacy Gov Proposal: Community Spend Funding Gov Module")
	s.voteGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter, "yes", false)

	initialGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, photonDenom)
	s.Require().NoError(err)
	proposalCounter++

	s.T().Logf("Submitting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_2.json")
	s.T().Logf("Depositing Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.depositGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter)
	s.T().Logf("Voting Gov Proposal: Sending Tokens from Gov Module to Recipient")
	s.voteGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter, "yes", false)
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
	s.depositGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter)
	s.T().Logf("Weighted Voting Gov Proposal: Software Upgrade")
	s.voteGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter, "yes=0.8,no=0.1,abstain=0.05,no_with_veto=0.05", true)

	s.verifyChainHaltedAtUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully halted chain at height %d", proposalHeight)

	s.TearDownSuite()

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
	s.T().Skip()

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
	s.depositGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter, "yes", false)

	proposalCounter++

	s.T().Logf("Submitting Gov Proposal: Cancel Software Upgrade")
	s.submitNewGovProposal(chainAAPIEndpoint, sender, proposalCounter, "/root/.gaia/config/proposal_4.json")
	s.depositGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, sender, fees.String(), proposalCounter, "yes", false)

	s.verifyChainPassesUpgradeHeight(s.chainA, 0, proposalHeight)
	s.T().Logf("Successfully canceled upgrade at height %d", proposalHeight)
}

func (s *IntegrationTestSuite) fundCommunityPool(chainAAPIEndpoint, sender string) {
	s.Run("fund_community_pool", func() {
		beforeDistPhotonBalance, _ := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
		if beforeDistPhotonBalance.IsNil() {
			// Set balance to 0 if previous balance does not exist
			beforeDistPhotonBalance = sdk.NewInt64Coin("photon", 0)
		}

		s.execDistributionFundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())

		// there are still tokens being added to the community pool through block production rewards but they should be less than 500 tokens
		marginOfErrorForBlockReward := sdk.NewInt64Coin("photon", 500)

		s.Require().Eventually(
			func() bool {
				afterDistPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
				if err != nil {
					s.T().Logf("Error getting balance: %s", afterDistPhotonBalance)
				}
				s.Require().NoError(err)

				return afterDistPhotonBalance.Sub(beforeDistPhotonBalance.Add(tokenAmount.Add(fees))).IsLT(marginOfErrorForBlockReward)
			},
			15*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) submitLegacyProposalFundGovAccount(chainAAPIEndpoint, sender string, proposalId int) {
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

func (s *IntegrationTestSuite) submitLegacyGovProposal(chainAAPIEndpoint string, sender string, fees string, proposalTypeSubCmd string, proposalId int, proposalPath string) {
	s.Run("submit_legacy_gov_proposal", func() {
		s.execGovSubmitLegacyGovProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalPath, fees, proposalTypeSubCmd)

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

func (s *IntegrationTestSuite) depositGovProposal(chainAAPIEndpoint string, sender string, fees string, proposalId int) {
	s.Run("deposit_gov_proposal", func() {
		s.execGovDepositProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, depositAmount, fees)

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

func (s *IntegrationTestSuite) voteGovProposal(chainAAPIEndpoint string, sender string, fees string, proposalId int, vote string, weighted bool) {
	s.Run("vote_gov_proposal", func() {
		if weighted {
			s.execGovWeightedVoteProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, vote, fees)
		} else {
			s.execGovVoteProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, vote, fees)
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

// globalfee in genesis is set to be "0.00001photon"
func (s *IntegrationTestSuite) TestQueryGlobalFeesInGenesis() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	feeInGenesis, err := sdk.ParseDecCoins(initialGlobalFeeAmt + photonDenom)
	s.Require().NoError(err)
	s.Require().Eventually(
		func() bool {
			fees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("Global Fees in Genesis: %s", fees.String())
			s.Require().NoError(err)

			return fees.IsEqual(feeInGenesis)
		},
		15*time.Second,
		5*time.Second,
	)
}

/*
global fee in genesis is "0.00001photon", which is the same as min_gas_price.
This initial value setup is for not to fail other e2e tests.
global fee e2e tests:
0. initial globalfee = 0.00001photon, min_gas_price = 0.00001photon

test1. gov proposal globalfee = [], query globalfee still get empty
  - tx with fee denom photon, fail
  - tx with zero fee denom photon, pass
  - tx with fee denom uatom, pass
  - tx with fee empty, pass

test2. gov propose globalfee =  0.000001photon(lower than min_gas_price)
  - tx with fee higher than 0.000001photon but lower than 0.00001photon, fail
  - tx with fee higher than/equal to 0.00001photon, pass
  - tx with fee uatom fail

test3. gov propose globalfee = 0.0001photon (higher than min_gas_price)
  - tx with fee equal to 0.0001photon, pass
  - tx with fee equal to 0.00001photon, fail

test4. gov propose globalfee =  0.000001photon (lower than min_gas_price), 0uatom
  - tx with fee 0.000001photon, fail
  - tx with fee 0.000001photon, pass
  - tx with empty fee, pass
  - tx with fee uatom pass
  - tx with fee 0uatom, 0.000005photon fail
  - tx with fee 0uatom, 0.00001photon pass
5. check balance correct: all the sucessful tx sent token amt is received
6. gov propose change back to initial globalfee = 0.00001photon, This is for not influence other e2e tests.
*/
func (s *IntegrationTestSuite) TestGlobalFees() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	submitterAddr, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	submitter := submitterAddr.String()
	recipientAddress, err := s.chainA.validators[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	recipient := recipientAddress.String()
	var beforeRecipientPhotonBalance sdk.Coin
	s.Require().Eventually(
		func() bool {
			beforeRecipientPhotonBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, photonDenom)
			s.Require().NoError(err)

			return beforeRecipientPhotonBalance.IsValid()
		},
		10*time.Second,
		5*time.Second,
	)
	if beforeRecipientPhotonBalance.Equal(sdk.Coin{}) {
		beforeRecipientPhotonBalance = sdk.NewCoin(photonDenom, sdk.ZeroInt())
	}

	sendAmt := int64(1000)
	token := sdk.NewInt64Coin(photonDenom, sendAmt) // send 1000photon each time
	sucessBankSendCount := 0
	// ---------------------------- test1: globalfee empty --------------------------------------------
	// prepare gov globalfee proposal
	emptyGlobalFee := sdk.DecCoins{}
	s.writeGovParamChangeProposalGlobalFees(s.chainA, emptyGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, fees.String(), "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, fees.String(), proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, fees.String(), proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)

			// attention: when global fee is empty, when query, it shows empty rather than default ante.DefaultZeroGlobalFee() = 0uatom.
			return globalFees.IsEqual(emptyGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	// test bank send with fees
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom, false)
	sucessBankSendCount++
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "", false)
	sucessBankSendCount++
	// wrong denom
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "4"+photonDenom, true)
	// zerofee wrong denom, pass
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom, false)
	sucessBankSendCount++

	// ------------------ test2: globalfee lower than min_gas_price -----------------------------------
	// prepare gov globalfee proposal
	lowGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec("photon", sdk.MustNewDecFromStr(lowGlobalFeesAmt))}
	s.writeGovParamChangeProposalGlobalFees(s.chainA, lowGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, "", "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, "", proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, "", proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)

			return globalFees.IsEqual(lowGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	paidFeeAmt := sdk.MustNewDecFromStr(minGasPrice).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtLowMinGasHighGlobalFee := sdk.MustNewDecFromStr(lowGlobalFeesAmt).Mul(sdk.NewDec(2)).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtLowGLobalFee := sdk.MustNewDecFromStr(lowGlobalFeesAmt).Quo(sdk.NewDec(2)).String()

	// paid fee higher than/equal to min_gas_price and global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+photonDenom, false)
	sucessBankSendCount++
	// paid fee lower than global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowGLobalFee+photonDenom, true)
	// paid fee higher/equal than global fee lower than min_gas_pirce
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowMinGasHighGlobalFee+photonDenom, true)
	// wrong denom
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, true)

	// ------------------ test3: globalfee higher than min_gas_price ----------------------------------
	// prepare gov globalfee proposal
	highGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec("photon", sdk.MustNewDecFromStr(highGlobalFeeAmt))}
	s.writeGovParamChangeProposalGlobalFees(s.chainA, highGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)

			return globalFees.IsEqual(highGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	paidFeeAmt = sdk.MustNewDecFromStr(highGlobalFeeAmt).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtHigherMinGasLowerGalobalFee := sdk.MustNewDecFromStr(minGasPrice).Quo(sdk.NewDec(2)).String()
	// paid fee higher than the global fee and min_gas_price
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+photonDenom, false)
	sucessBankSendCount++
	// paid fee higher than/equal to min_gas_price but lower than global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtHigherMinGasLowerGalobalFee+photonDenom, true)

	// ---------------------------- test4: global fee with two denoms -----------------------------------
	// prepare gov globalfee proposal
	mixGlobalFee := sdk.DecCoins{
		sdk.NewDecCoinFromDec(photonDenom, sdk.MustNewDecFromStr(lowGlobalFeesAmt)),
		sdk.NewDecCoinFromDec(uatomDenom, sdk.NewDec(0)),
	}.Sort()
	s.writeGovParamChangeProposalGlobalFees(s.chainA, mixGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter, "yes", false)

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)
			return globalFees.IsEqual(mixGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	// equal to min_gas_price
	paidFeeAmt = sdk.MustNewDecFromStr(minGasPrice).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtLow := sdk.MustNewDecFromStr(lowGlobalFeesAmt).Quo(sdk.NewDec(2)).Mul(sdk.NewDec(gas)).String()
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+photonDenom, false)
	sucessBankSendCount++
	// paid fee lower than global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLow+photonDenom, true)
	// empty paid fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "", false)
	sucessBankSendCount++
	// paid fee in uatom
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom, false)
	sucessBankSendCount++
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "2"+uatomDenom, false)
	sucessBankSendCount++
	// paid fee 0uatom and photon lower than global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom+","+paidFeeAmtLow+photonDenom, true)
	// paid fee 0uatom and photon higher than/equal to global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom+","+paidFeeAmt+photonDenom, false)
	sucessBankSendCount++
	// paid fee 2uatom and photon higher than/equal to global fee
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "2"+uatomDenom+","+paidFeeAmt+photonDenom, false)
	sucessBankSendCount++
	// ---------------------------------------------------------------------------

	// check the balance is correct after previous txs
	s.Require().Eventually(
		func() bool {
			afterRecipientPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, photonDenom)
			s.Require().NoError(err)
			IncrementedPhoton := afterRecipientPhotonBalance.Sub(beforeRecipientPhotonBalance)
			photonSent := sdk.NewInt64Coin(photonDenom, sendAmt*int64(sucessBankSendCount))
			return IncrementedPhoton.IsEqual(photonSent)
		},
		time.Minute,
		5*time.Second,
	)

	// gov proposing to change back to original global fee
	s.T().Logf("Propose to change back to original global fees: %s", initialGlobalFeeAmt+photonDenom)
	oldfees, err := sdk.ParseDecCoins(initialGlobalFeeAmt + photonDenom)
	s.Require().NoError(err)
	s.writeGovParamChangeProposalGlobalFees(s.chainA, oldfees)

	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change back global fees")
	// fee is 0uatom
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, proposalCounter, "yes", false)

	// query the proposal status and fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainAAPIEndpoint, proposalCounter)
			s.Require().NoError(err)
			return (proposal.GetProposal().Status == govv1beta1.StatusPassed)
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			fees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov proposal to change back global fees: %s", oldfees.String())
			s.Require().NoError(err)

			return fees.IsEqual(oldfees)
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) TestByPassMinFeeWithdrawReward() {
	// time.Sleep(10)
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	paidFeeAmt := sdk.MustNewDecFromStr(minGasPrice).Mul(sdk.NewDec(gas)).String()

	payee, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	// pass
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), paidFeeAmt+photonDenom, false)
	// pass
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), "0"+photonDenom, false)
	// pass
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), "0"+uatomDenom, false)
	// fail
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), paidFeeAmt+uatomDenom, true)
}
