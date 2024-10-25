package interchain_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v21/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

const (
	delegatorName      = "delegator1"
	stakeAmount        = "10000000" // 10 ATOM
	submissionDeposit  = "100"
	proposalDepositInt = chainsuite.GovMinDepositAmount
	yesWeight          = 0.6
	noWeight           = 0.3
	abstainWeight      = 0.06
	vetoWeight         = 0.04
)

var proposalId string = "0"

type DelegatorGovernanceSuite struct {
	*chainsuite.Suite
}

func (s *DelegatorGovernanceSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Create delegator account
	err := s.Chain.CreateKey(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddress, err := s.Chain.GetAddress(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddressString := types.MustBech32ifyAddressBytes("cosmos", delegatorAddress)
	fmt.Println("delegator address:", delegatorAddressString)

	// Fund delegator account
	err = s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   chainsuite.Uatom,
		Amount:  sdkmath.NewInt(chainsuite.ValidatorFunds),
		Address: delegatorAddressString,
	})
	s.Require().NoError(err)
	balance, err := s.Chain.GetBalance(s.GetContext(), delegatorAddressString, chainsuite.Uatom)
	fmt.Println("Balance:", balance)

	// Delegate >1 ATOM with delegator account
	node := s.Chain.GetNode()
	node.StakingDelegate(s.GetContext(), delegatorName, s.Chain.ValidatorWallets[0].ValoperAddress, string(stakeAmount)+s.Chain.Config().Denom)
}

func (s *DelegatorGovernanceSuite) Test1SubmitProposal() {
	delegatorAddress, err := s.Chain.GetAddress(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddressString := types.MustBech32ifyAddressBytes("cosmos", delegatorAddress)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Test Proposal", "Test Proposal", "ipfs://CID", submissionDeposit+"uatom", delegatorAddressString, false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), delegatorName, prop)
	s.Require().NoError(err)
	proposalId = result.ProposalID

	// Get status
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)
	currentStatus := proposal.Status.String()

	// Test
	s.Require().NotEqual("0", proposalId)
	s.Require().Equal("PROPOSAL_STATUS_DEPOSIT_PERIOD", currentStatus)
}

func (s *DelegatorGovernanceSuite) Test2ProposalDeposit() {
	node := s.Chain.GetNode()
	delegatorAddress, err := s.Chain.GetAddress(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddressString := types.MustBech32ifyAddressBytes("cosmos", delegatorAddress)
	proposalDeposit := strconv.Itoa(proposalDepositInt)

	// Submit deposit to proposal
	_, err = node.ExecTx(s.GetContext(), delegatorName, "gov", "deposit", proposalId, proposalDeposit+"uatom", "--gas", "auto")
	s.Require().NoError(err)
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)
	currentStatus := proposal.Status.String()

	submissionDepositUint, err := strconv.ParseUint(submissionDeposit, 10, 64)
	s.Require().NoError(err)
	depositTotal := chainsuite.GovMinDepositAmount + submissionDepositUint
	fmt.Println("Deposit total: ", depositTotal)

	deposit, err := s.Chain.QueryJSON(s.GetContext(), "deposit", "gov", "deposit", proposalId, delegatorAddressString)
	s.Require().NoError(err)
	depositAmount := deposit.Get("amount.#(denom==\"uatom\").amount").String()
	depositAmountUint, err := strconv.ParseUint(depositAmount, 10, 64)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Query amount: %d", depositAmountUint)
	// Test
	s.Require().Equal("PROPOSAL_STATUS_VOTING_PERIOD", currentStatus)
	s.Require().Equal(depositTotal, depositAmountUint)
}

func (s *DelegatorGovernanceSuite) Test3ProposalVote() {
	node := s.Chain.GetNode()
	delegatorAddress, err := s.Chain.GetAddress(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddressString := types.MustBech32ifyAddressBytes("cosmos", delegatorAddress)
	// Vote yes on proposal
	_, err = node.ExecTx(s.GetContext(), delegatorName, "gov", "vote", proposalId, "yes", "--gas", "auto")
	s.Require().NoError(err)

	// Test
	vote, err := s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, delegatorAddressString)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", vote)
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_yes_weight.String())
	s.Require().Equal(float64(1.0), actual_yes_weight.Float())
}

func (s *DelegatorGovernanceSuite) Test4ProposalWeightedVote() {
	node := s.Chain.GetNode()
	delegatorAddress, err := s.Chain.GetAddress(s.GetContext(), delegatorName)
	s.Require().NoError(err)
	delegatorAddressString := types.MustBech32ifyAddressBytes("cosmos", delegatorAddress)
	// Submit weighted vote to proposal
	_, err = node.ExecTx(s.GetContext(), delegatorName, "gov", "weighted-vote", proposalId, fmt.Sprintf("yes=%0.2f,no=%0.2f,abstain=%0.2f,no_with_veto=%0.2f", yesWeight, noWeight, abstainWeight, vetoWeight), "--gas", "auto")
	s.Require().NoError(err)

	// Test
	vote, err := s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, delegatorAddressString)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", vote)
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_yes_weight.String())
	s.Require().Equal(float64(yesWeight), actual_yes_weight.Float())
	actual_no_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_no_weight.String())
	s.Require().Equal(float64(noWeight), actual_no_weight.Float())
	actual_abstain_weight := vote.Get("options.#(option==\"VOTE_OPTION_ABSTAIN\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_abstain_weight.String())
	s.Require().Equal(float64(abstainWeight), actual_abstain_weight.Float())
	actual_veto_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO_WITH_VETO\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_veto_weight.String())
	s.Require().Equal(float64(vetoWeight), actual_veto_weight.Float())
}

func TestDelegatorGovernance(t *testing.T) {
	delegatorGovSuite := DelegatorGovernanceSuite{chainsuite.NewSuite(chainsuite.SuiteConfig{UpgradeOnSetup: true})}
	suite.Run(t, &delegatorGovSuite)
}
