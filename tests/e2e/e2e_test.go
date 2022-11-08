package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func (s *IntegrationTestSuite) TestGov() {
	s.SendTokensFromNewGovAccount()
	s.GovSoftwareUpgrade()
	s.GovCancelSoftwareUpgrade()
}

// globalfee in genesis is set to be "0.00001uatom"
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
	s.Run("global fees", func() {
		var globalFees sdk.DecCoins

		paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
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
			beforeRecipientPhotonBalance = sdk.NewCoin(photonDenom, math.ZeroInt())
		}

		sendAmt := int64(1000)
		token := sdk.NewInt64Coin(photonDenom, sendAmt) // send 1000photon each time
		sucessBankSendCount := 0

		s.Run("empty global fee", func() {
			// prepare gov globalfee proposal
			emptyGlobalFee := sdk.DecCoins{}
			s.writeGovParamChangeProposalGlobalFees(s.chainA, emptyGlobalFee)

			// gov proposing new fees
			s.proposalCounter++
			s.T().Logf("Proposal number: %d", s.proposalCounter)
			s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
			s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), "param-change", s.proposalCounter, configFile(proposalGlobalFee))
			s.depositGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), s.proposalCounter)
			s.voteGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), s.proposalCounter, "yes", false)

			// query the proposal status and new fee
			s.Require().Eventually(
				func() bool {
					proposal, err := queryGovProposal(chainAAPIEndpoint, s.proposalCounter)
					s.Require().NoError(err)
					return proposal.GetProposal().Status == govv1beta1.StatusPassed
				},
				15*time.Second,
				5*time.Second,
			)

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

			s.T().Logf("test case: empty global fee, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
			s.T().Logf("Tx fee is zero coin with correct denom: uatom, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "0"+uatomDenom, true)
			s.T().Logf("Tx fee is empty, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "", true)
			s.T().Logf("Tx with wrong denom: photon, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "4"+photonDenom, true)
			s.T().Logf("Tx fee is zero coins of wrong denom: photon, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom, true)
			s.T().Logf("Tx fee is higher than min_gas_price, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
		})

		s.Run("global fee lower than min_gas_price", func() {
			// prepare gov globalfee proposal
			lowGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(lowGlobalFeesAmt))}
			s.writeGovParamChangeProposalGlobalFees(s.chainA, lowGlobalFee)

			// gov proposing new fees
			s.proposalCounter++
			s.T().Logf("Proposal number: %d", s.proposalCounter)
			s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
			s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), "param-change", s.proposalCounter, configFile(proposalGlobalFee))
			s.depositGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), s.proposalCounter)
			s.voteGovProposal(chainAAPIEndpoint, submitter, standardFees.String(), s.proposalCounter, "yes", false)

			// query the proposal status and new fee
			s.Require().Eventually(
				func() bool {
					proposal, err := queryGovProposal(chainAAPIEndpoint, s.proposalCounter)
					s.Require().NoError(err)
					return proposal.GetProposal().Status == govv1beta1.StatusPassed
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

			paidFeeAmtLowMinGasHighGlobalFee := math.LegacyMustNewDecFromStr(lowGlobalFeesAmt).
				Mul(math.LegacyNewDec(2)).
				Mul(math.LegacyNewDec(gas)).
				String()
			paidFeeAmtLowGlobalFee := math.LegacyMustNewDecFromStr(lowGlobalFeesAmt).Quo(math.LegacyNewDec(2)).String()

			s.T().Logf("test case: global fee is lower than min_gas_price, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
			s.T().Logf("Tx fee higher than/equal to min_gas_price and global fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx fee lower than/equal to min_gas_price and global fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowGlobalFee+uatomDenom, true)
			s.T().Logf("Tx fee lower than/equal global fee and lower than min_gas_price, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLowMinGasHighGlobalFee+uatomDenom, true)
			s.T().Logf("Tx fee has wrong denom, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+photonDenom, true)
		})

		s.Run("global fee higher than min_gas_price", func() {
			// prepare gov globalfee proposal
			highGlobalFee := sdk.DecCoins{sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(highGlobalFeeAmt))}
			s.writeGovParamChangeProposalGlobalFees(s.chainA, highGlobalFee)

			// gov proposing new fees
			s.proposalCounter++
			s.T().Logf("Proposal number: %d", s.proposalCounter)
			s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
			s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, "param-change", s.proposalCounter, configFile(proposalGlobalFee))
			s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, s.proposalCounter)
			s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, s.proposalCounter, "yes", false)

			// query the proposal status and new fee
			s.Require().Eventually(
				func() bool {
					proposal, err := queryGovProposal(chainAAPIEndpoint, s.proposalCounter)
					s.Require().NoError(err)
					return proposal.GetProposal().Status == govv1beta1.StatusPassed
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

			paidFeeAmt := math.LegacyMustNewDecFromStr(highGlobalFeeAmt).Mul(math.LegacyNewDec(gas)).String()
			paidFeeAmtHigherMinGasLowerGalobalFee := math.LegacyMustNewDecFromStr(minGasPrice).
				Quo(math.LegacyNewDec(2)).String()

			s.T().Logf("test case: global fee is higher than min_gas_price, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
			s.T().Logf("Tx fee is higher than/equal to global fee and min_gas_price, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx fee is higher than/equal to min_gas_price but lower than global fee, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtHigherMinGasLowerGalobalFee+uatomDenom, true)
		})

		s.Run("global fees with two denoms", func() {
			// prepare gov globalfee proposal
			mixGlobalFee := sdk.DecCoins{
				sdk.NewDecCoinFromDec(photonDenom, sdk.NewDec(0)),
				sdk.NewDecCoinFromDec(uatomDenom, sdk.MustNewDecFromStr(lowGlobalFeesAmt)),
			}.Sort()
			s.writeGovParamChangeProposalGlobalFees(s.chainA, mixGlobalFee)

			// gov proposing new fees
			s.proposalCounter++
			s.T().Logf("Proposal number: %d", s.proposalCounter)
			s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change global fees empty")
			s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, "param-change", s.proposalCounter, configFile(proposalGlobalFee))
			s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, s.proposalCounter)
			s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+uatomDenom, s.proposalCounter, "yes", false)

			// query the proposal status and new fee
			s.Require().Eventually(
				func() bool {
					proposal, err := queryGovProposal(chainAAPIEndpoint, s.proposalCounter)
					s.Require().NoError(err)
					return proposal.GetProposal().Status == govv1beta1.StatusPassed
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
			paidFeeAmtLow := math.LegacyMustNewDecFromStr(lowGlobalFeesAmt).
				Quo(math.LegacyNewDec(2)).
				Mul(math.LegacyNewDec(gas)).
				String()

			s.T().Logf("test case: global fees contain multiple denoms: one zero coin, one non-zero coin, globalfee=%s, min_gas_price=%s", globalFees.String(), minGasPrice+uatomDenom)
			s.T().Logf("Tx with fee higher than/equal to one of denom's amount the global fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx with fee lower than one of denom's amount the global fee, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), paidFeeAmtLow+uatomDenom, true)
			s.T().Logf("Tx with fee empty fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "", false)
			sucessBankSendCount++
			s.T().Logf("Tx with zero coin in the denom of zero coin of global fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx with non-zero coin in the denom of zero coin of global fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "2"+photonDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx with mulitple fee coins, zero coin and low fee, fail")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom+","+paidFeeAmtLow+uatomDenom, true)
			s.T().Logf("Tx with mulitple fee coins, zero coin and high fee, pass")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "0"+photonDenom+","+paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
			s.T().Logf("Tx with mulitple fee coins, all higher than global fee and min_gas_price")
			s.execBankSend(s.chainA, 0, submitter, recipient, token.String(), "2"+photonDenom+","+paidFeeAmt+uatomDenom, false)
			sucessBankSendCount++
		})

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

		s.proposalCounter++
		s.T().Logf("Proposal number: %d", s.proposalCounter)
		s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: change back global fees")
		// fee is 0uatom
		s.submitLegacyGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, "param-change", s.proposalCounter, configFile(proposalGlobalFee))
		s.depositGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, s.proposalCounter)
		s.voteGovProposal(chainAAPIEndpoint, submitter, paidFeeAmt+photonDenom, s.proposalCounter, "yes", false)

		// query the proposal status and fee
		s.Require().Eventually(
			func() bool {
				proposal, err := queryGovProposal(chainAAPIEndpoint, s.proposalCounter)
				s.Require().NoError(err)
				return proposal.GetProposal().Status == govv1beta1.StatusPassed
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
	})
}

func (s *IntegrationTestSuite) TestByPassMinFeeWithdrawReward() {

	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	// pass
	s.T().Logf("bypass-msg with fee in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), paidFeeAmt+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), "0"+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin not in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), "0"+photonDenom, false)
	// fail
	s.T().Logf("bypass-msg with non-zero coin not in the denom of global fee, fail")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), paidFeeAmt+photonDenom, true)
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStaking() {

	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorB := s.chainA.validators[1]
	validatorAAddr, err := validatorA.keyInfo.GetAddress()
	s.Require().NoError(err)
	validatorBAddr, err := validatorB.keyInfo.GetAddress()
	s.Require().NoError(err)

	valOperA := sdk.ValAddress(validatorAAddr)
	valOperB := sdk.ValAddress(validatorBAddr)

	alice, err := s.chainA.genesisAccounts[2].keyInfo.GetAddress()
	s.Require().NoError(err)
	bob, err := s.chainA.genesisAccounts[3].keyInfo.GetAddress()
	s.Require().NoError(err)

	delegationFees := sdk.NewCoin(uatomDenom, math.NewInt(10))

	s.testStaking(chainAAPIEndpoint, alice.String(), valOperA.String(), valOperB.String(), delegationFees, gaiaHomePath)
	s.testDistribution(chainAAPIEndpoint, alice.String(), bob.String(), valOperB.String(), gaiaHomePath)
}

func (s *IntegrationTestSuite) TestGroups() {
	s.GroupsSendMsgTest()
}

func (s *IntegrationTestSuite) TestVesting() {
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	s.testPermanentLockedAccount(chainAAPI)
	s.testPeriodicVestingAccount(chainAAPI)
}

func (s *IntegrationTestSuite) TestSlashing() {
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
}
