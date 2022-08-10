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

	s.Run("send_uatom_to_chainB", func() {
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
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				fmt.Print("c.denom:", c.Denom, "c.amount:", c.Amount)
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
			beforeSenderUAtomBalance    sdk.Coin
			beforeRecipientUAtomBalance sdk.Coin
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, sender, "uatom")
				s.Require().NoError(err)

				beforeRecipientUAtomBalance, err = getSpecificBalance(chainAAPIEndpoint, recipient, "uatom")
				s.Require().NoError(err)

				return beforeSenderUAtomBalance.IsValid() && beforeRecipientUAtomBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		s.sendMsgSend(s.chainA, 0, sender, recipient, tokenAmount.String(), fees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterSenderUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, "uatom")
				s.Require().NoError(err)

				afterRecipientUAtomBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, "uatom")
				s.Require().NoError(err)

				decremented := beforeSenderUAtomBalance.Sub(tokenAmount).Sub(fees).IsEqual(afterSenderUAtomBalance)
				incremented := beforeRecipientUAtomBalance.Add(tokenAmount).IsEqual(afterRecipientUAtomBalance)

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

	initialGovBalance, err := getSpecificBalance(chainAAPIEndpoint, govModuleAddress, uatomDenom)
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
		beforeDistUatomBalance, _ := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
		if beforeDistUatomBalance.IsNil() {
			// Set balance to 0 if previous balance does not exist
			beforeDistUatomBalance = sdk.NewInt64Coin(uatomDenom, 0)
		}

		s.execDistributionFundCommunityPool(s.chainA, 0, chainAAPIEndpoint, sender, tokenAmount.String(), fees.String())

		// there are still tokens being added to the community pool through block production rewards but they should be less than 500 tokens
		marginOfErrorForBlockReward := sdk.NewInt64Coin(uatomDenom, 500)

		s.Require().Eventually(
			func() bool {
				afterDistPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, distModuleAddress, tokenAmount.Denom)
				if err != nil {
					s.T().Logf("Error getting balance: %s", afterDistPhotonBalance)
				}
				s.Require().NoError(err)

				return afterDistPhotonBalance.Sub(beforeDistUatomBalance.Add(tokenAmount.Add(fees))).IsLT(marginOfErrorForBlockReward)
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

func (s *IntegrationTestSuite) submitNewGovProposal(chainAAPIEndpoint, sender string, proposalId int, proposalPath string) {
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

func (s *IntegrationTestSuite) depositGovProposal(chainAAPIEndpoint, sender string, fees string, proposalId int) {
	s.Run("deposit_gov_proposal", func() {
		s.execGovDepositProposal(s.chainA, 0, chainAAPIEndpoint, sender, proposalId, depositAmount.String(), fees)

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

func (s *IntegrationTestSuite) voteGovProposal(chainAAPIEndpoint, sender string, fees string, proposalId int, vote string, weighted bool) {
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
	feeInGenesis, err := sdk.ParseDecCoins(initialGlobalFeeAmt + uatomDenom)
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
global fee in genesis is "0.00001uatom", which is the same as min_gas_price.
This initial value setup is for not to fail other e2e tests.
global fee e2e tests:
0. initial globalfee = 0.00001uatom, min_gas_price = 0.00001uatom

test1. gov proposal globalfee = [], min_gas_price=0.00001uatom, query globalfee still get empty
  - tx with fee denom photon, fail
  - tx with zero fee denom photon, fail
  - tx with fee denom uatom, pass
  - tx with fee empty, fail

test2. gov propose globalfee =  0.000001uatom(lower than min_gas_price)
  - tx with fee higher than 0.000001uatom but lower than 0.00001uatom, fail
  - tx with fee higher than/equal to 0.00001uatom, pass
  - tx with fee photon fail

test3. gov propose globalfee = 0.0001uatom (higher than min_gas_price)
  - tx with fee equal to 0.0001uatom, pass
  - tx with fee equal to 0.00001uatom, fail

test4. gov propose globalfee =  0.000001uatom (lower than min_gas_price), 0photon
  - tx with fee 0.0000001photon, fail
  - tx with fee 0.000001photon, pass
  - tx with empty fee, pass
  - tx with fee photon pass
  - tx with fee 0photon, 0.000005uatom fail
  - tx with fee 0photon, 0.00001uatom pass
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
	var globalFees sdk.DecCoins
	s.Require().Eventually(
		func() bool {
			globalFees, err = queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)

			// attention: when global fee is empty, when query, it shows empty rather than default ante.DefaultZeroGlobalFee() = 0uatom.
			return globalFees.IsEqual(emptyGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	paidFeeAmt := sdk.MustNewDecFromStr(minGasPrice).Mul(sdk.NewDec(gas)).String()

	s.T().Logf("test case: empty global fee, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
	s.T().Logf("Tx fee is zero coin with correct denom: uatom, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom, true)
	s.T().Logf("Tx fee is empty, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "", true)
	s.T().Logf("Tx with wrong denom: photon, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "4"+photonDenom, true)
	s.T().Logf("Tx fee is zero coins of wrong denom: photon, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom, true)
	s.T().Logf("Tx fee is higher than min_gas_price, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
	sucessBankSendCount++

	// ------------------ test2: globalfee lower than min_gas_price -----------------------------------
	// prepare gov globalfee proposal
	lowGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(lowGlobalFeesAmt))}
	s.writeGovParamChangeProposalGlobalFees(s.chainA, lowGlobalFee)

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

			return globalFees.IsEqual(lowGlobalFee)
		},
		15*time.Second,
		5*time.Second,
	)

	paidFeeAmt = sdk.MustNewDecFromStr(minGasPrice).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtLowMinGasHighGlobalFee := sdk.MustNewDecFromStr(lowGlobalFeesAmt).Mul(sdk.NewDec(2)).Mul(sdk.NewDec(gas)).String()
	paidFeeAmtLowGLobalFee := sdk.MustNewDecFromStr(lowGlobalFeesAmt).Quo(sdk.NewDec(2)).String()

	s.T().Logf("test case: global fee is lower than min_gas_price, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
	s.T().Logf("Tx fee higher than/equal to min_gas_price and global fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx fee lower than/equal to min_gas_price and global fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowGLobalFee+uatomDenom, true)
	s.T().Logf("Tx fee lower than/equal global fee and lower than min_gas_price, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowMinGasHighGlobalFee+uatomDenom, true)
	s.T().Logf("Tx fee has wrong denom, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+photonDenom, true)

	// ------------------ test3: globalfee higher than min_gas_price ----------------------------------
	// prepare gov globalfee proposal
	highGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(highGlobalFeeAmt))}
	s.writeGovParamChangeProposalGlobalFees(s.chainA, highGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, proposalCounter, "yes", false)

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

	s.T().Logf("test case: global fee is higher than min_gas_price, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
	s.T().Logf("Tx fee is higher than/equal to global fee and min_gas_price, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx fee is higher than/equal to min_gas_price but lower than global fee, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtHigherMinGasLowerGalobalFee+uatomDenom, true)

	// ---------------------------- test4: global fee with two denoms -----------------------------------
	// prepare gov globalfee proposal
	mixGlobalFee := sdk.DecCoins{
		sdk.NewDecCoinFromDec(photonDenom, sdk.NewDec(0)),
		sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(lowGlobalFeesAmt)),
	}.Sort()
	s.writeGovParamChangeProposalGlobalFees(s.chainA, mixGlobalFee)

	// gov proposing new fees
	proposalCounter++
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
	s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, "param-change", proposalCounter, "/root/.gaia/config/proposal_globalfee.json")
	s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, proposalCounter)
	s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, proposalCounter, "yes", false)

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

	s.T().Logf("test case: global fees contain multiple denoms: one zero coin, one non-zero coin, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
	s.T().Logf("Tx with fee higher than/equal to one of denom's amount the global fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx with fee lower than one of denom's amount the global fee, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLow+uatomDenom, true)
	s.T().Logf("Tx with fee empty fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "", false)
	sucessBankSendCount++
	s.T().Logf("Tx with zero coin in the denom of zero coin of global fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx with non-zero coin in the denom of zero coin of global fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "2"+photonDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx with mulitple fee coins, zero coin and low fee, fail")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom+","+paidFeeAmtLow+uatomDenom, true)
	s.T().Logf("Tx with mulitple fee coins, zero coin and high fee, pass")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom+","+paidFeeAmt+uatomDenom, false)
	sucessBankSendCount++
	s.T().Logf("Tx with mulitple fee coins, all higher than global fee and min_gas_price")
	s.sendMsgSend(s.chainA, 0, submitter, recipient, token.String(), "2"+photonDenom+","+paidFeeAmt+uatomDenom, false)
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
	s.T().Logf("Propose to change back to original global fees: %s", initialGlobalFeeAmt+uatomDenom)
	oldfees, err := sdk.ParseDecCoins(initialGlobalFeeAmt + uatomDenom)
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
	s.T().Logf("bypass-msg with fee in the denom of global fee, pass")
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), paidFeeAmt+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin in the denom of global fee, pass")
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), "0"+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin not in the denom of global fee, pass")
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), "0"+photonDenom, false)
	// fail
	s.T().Logf("bypass-msg with non-zero coin not in the denom of global fee, fail")
	s.withdrawReward(s.chainA, 0, chainAAPIEndpoint, payee.String(), paidFeeAmt+photonDenom, true)
}

// todo add fee test with wrong denom order
