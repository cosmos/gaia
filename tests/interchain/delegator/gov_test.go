package delegator_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

const (
	stakeAmount          = "10000000" // 10 ATOM
	submissionDeposit    = "100"
	proposalDepositInt   = chainsuite.GovMinDepositAmount
	communityPoolAmount  = "200000000" // 200 ATOM
	yesWeight            = 0.6
	noWeight             = 0.3
	abstainWeight        = 0.06
	vetoWeight           = 0.04
	queryScaleMultiplier = 1000000000000000000 // 18 zeroes
)

type GovSuite struct {
	*delegator.Suite
}

func (s *GovSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Delegate >1 ATOM with delegator account
	node := s.Chain.GetNode()
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(stakeAmount)+s.Chain.Config().Denom)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(stakeAmount)+s.Chain.Config().Denom)
}

func (s *GovSuite) TestProposal() {
	// Test:
	// 1. Proposal submission
	// 2. Proposal deposit
	// 3. Vote
	// 4. Weighted vote

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Test Proposal", "Test Proposal", "ipfs://CID", submissionDeposit+"uatom", s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID

	// Get status
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)

	// Test submission
	currentStatus := proposal.Status.String()
	s.Require().Equal("PROPOSAL_STATUS_DEPOSIT_PERIOD", currentStatus)

	// Submit deposit to proposal
	proposalDeposit := strconv.Itoa(proposalDepositInt)
	node := s.Chain.GetNode()
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet.KeyName(), "gov", "deposit", proposalId, proposalDeposit+"uatom", "--gas", "auto")
	s.Require().NoError(err)
	submissionDepositUint, err := strconv.ParseUint(submissionDeposit, 10, 64)
	s.Require().NoError(err)
	depositTotal := proposalDepositInt + submissionDepositUint

	// Get status
	proposalIDuint, err = strconv.ParseUint(proposalId, 10, 64)
	proposal, err = s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)
	currentStatus = proposal.Status.String()

	// Test deposit
	deposit, err := s.Chain.QueryJSON(s.GetContext(), "deposit", "gov", "deposit", proposalId, s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)
	depositAmount := deposit.Get("amount.#(denom==\"uatom\").amount").String()
	depositAmountUint, err := strconv.ParseUint(depositAmount, 10, 64)
	s.Require().NoError(err)
	s.Require().Equal("PROPOSAL_STATUS_VOTING_PERIOD", currentStatus)
	s.Require().Equal(depositTotal, depositAmountUint)

	// Submit yes vote
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet.KeyName(), "gov", "vote", proposalId, "yes", "--gas", "auto")
	s.Require().NoError(err)

	// Test vote
	vote, err := s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_yes_weight.String())
	s.Require().Equal(float64(1.0), actual_yes_weight.Float())

	// Submit weighted vote
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(), "gov", "weighted-vote", proposalId, fmt.Sprintf("yes=%0.2f,no=%0.2f,abstain=%0.2f,no_with_veto=%0.2f", yesWeight, noWeight, abstainWeight, vetoWeight), "--gas", "auto")
	s.Require().NoError(err)

	// Test weighted vote
	vote, err = s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err)
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", vote)
	actual_yes_weight = vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_yes_weight.String())
	s.Require().Equal(float64(yesWeight), actual_yes_weight.Float())
	actual_no_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_no_weight.String())
	s.Require().Equal(float64(noWeight), actual_no_weight.Float())
	actual_abstain_weight := vote.Get("options.#(option==\"VOTE_OPTION_ABSTAIN\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_abstain_weight.String())
	s.Require().Equal(float64(abstainWeight), actual_abstain_weight.Float())
	actual_veto_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO_WITH_VETO\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_veto_weight.String())
	s.Require().Equal(float64(vetoWeight), actual_veto_weight.Float())
}

func (s *GovSuite) TestParamChange() {
	govParams, err := s.Chain.QueryJSON(s.GetContext(), "params", "gov", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", govParams)
	currentThreshold := govParams.Get("threshold").Float()
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Current threshold: %f", currentThreshold)
	newThreshold := 0.6

	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	updatedParams, err := sjson.Set(govParams.String(), "threshold", fmt.Sprintf("%f", newThreshold))
	s.Require().NoError(err)
	updatedParams, err = sjson.Set(updatedParams, "expedited_voting_period", "10s")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Updated params: %s", updatedParams)

	paramChangeMessage := fmt.Sprintf(`{
    	"@type": "/cosmos.gov.v1.MsgUpdateParams",
		"authority": "%s",
    	"params": %s
	}`, authority, updatedParams)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Message: %s", paramChangeMessage)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Gov Param Change Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.KeyName(), false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(paramChangeMessage)}
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID

	json, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "gov", "proposal", proposalId)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(json))

	// Pass proposal
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), proposalId))

	// Test
	govParams, err = s.Chain.QueryJSON(s.GetContext(), "params", "gov", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", govParams)
	currentThreshold = govParams.Get("threshold").Float()
	s.Require().Equal(newThreshold, currentThreshold)
}

func (s *GovSuite) TestGovFundCommunityPool() {
	// Fund gov account
	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)
	err = s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   chainsuite.Uatom,
		Amount:  sdkmath.NewInt(250000000),
		Address: authority,
	})
	s.Require().NoError(err)
	balance, err := s.Chain.GetBalance(s.GetContext(), authority, chainsuite.Uatom)
	fmt.Println("Gov module balance:", balance)

	jsonMsg, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "distribution", "community-pool")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(jsonMsg))
	poolBalanceJson := gjson.Get(string(jsonMsg), "pool.#(%\"*uatom\")")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Community pool balance: %s", poolBalanceJson.String())
	startingPoolBalance, err := chainsuite.StrToSDKInt(poolBalanceJson.String())
	s.Require().NoError(err)

	// Build proposal
	fundMessage := fmt.Sprintf(`{
        "@type": "/cosmos.distribution.v1beta1.MsgFundCommunityPool",
        "amount": [
          {
            "denom": "uatom",
            "amount": "%s"
          }
        ],
        "depositor": "%s"
	}`, communityPoolAmount, authority)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Community Pool Funding Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(fundMessage)}
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID

	// Pass proposal
	err = s.Chain.PassProposal(s.GetContext(), proposalId)

	jsonMsg, _, err = s.Chain.GetNode().ExecQuery(s.GetContext(), "distribution", "community-pool")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(jsonMsg))

	// Test
	poolBalanceJson = gjson.Get(string(jsonMsg), "pool.#(%\"*uatom\")")
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Community pool balance: %s", poolBalanceJson.String())
	endingPoolBalance, err := chainsuite.StrToSDKInt(poolBalanceJson.String())
	s.Require().NoError(err)
	balanceDifference := endingPoolBalance.Uint64() - startingPoolBalance.Uint64()
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("The community pool balance increased by %d", balanceDifference)
	fundAmount, err := strconv.ParseUint(communityPoolAmount, 10, 64)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(balanceDifference, fundAmount)
}

func TestGov(t *testing.T) {
	s := &GovSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
