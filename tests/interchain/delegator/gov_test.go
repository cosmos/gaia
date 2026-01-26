package delegator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
)

const (
	govStakeAmount          = "10000000" // 10 ATOM
	govSubmissionDeposit    = "100"
	proposalDepositInt      = chainsuite.GovMinDepositAmount
	govCommunityPoolAmount  = "200000000" // 200 ATOM
	govYesWeight            = 0.6
	govNoWeight             = 0.3
	govAbstainWeight        = 0.06
	govVetoWeight           = 0.04
	govQueryScaleMultiplier = 1000000000000000000 // 18 zeroes
)

type GovSuite struct {
	*delegator.Suite
	Host            *chainsuite.Chain
	icaAddress      string
	srcChannel      *ibc.ChannelOutput
	srcAddress      string
	ibcStakeDenom   string
	contractWasm    []byte
	contractPath    string
	contractAddress string
}

func (s *GovSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Delegate >1 ATOM with delegator account
	node := s.Chain.GetNode()
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)

	// WASM: The delegate-vote contract will be used to test cosmwasm governance votes
	contractWasm, err := os.ReadFile("testdata/delegate_vote.wasm")
	s.Require().NoError(err)
	s.contractWasm = contractWasm

	// WriteFile expects a relative path from the node's home directory
	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), s.contractWasm, "delegate_vote.wasm"))

	// Store the contract path for later use - use full path within the container
	s.contractPath = path.Join(s.Chain.GetNode().HomeDir(), "delegate_vote.wasm")

	// Store the contract using tx wasm store command directly
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "store", s.contractPath, "--gas", "auto",
	)
	s.Require().NoError(err)

	// Wait for blocks to ensure the contract is stored before instantiating
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.Chain))

	// Instantiate the contract
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "instantiate", "1", "{}", "--label", "delegate-vote", "--no-admin", "--gas", "auto",
	)
	s.Require().NoError(err)

	// Obtain the contract address
	contractInfo, err := s.Chain.QueryJSON(s.GetContext(), "contracts", "wasm", "list-contract-by-code", "1")
	s.Require().NoError(err)
	s.contractAddress = contractInfo.Get("0").String()
	fmt.Printf("Delegate-vote contract address: %s\n", s.contractAddress)

	// Fund contract address with 10ATOM
	err = s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   s.Chain.Config().Denom,
		Amount:  sdkmath.NewInt(10000000),
		Address: s.contractAddress,
	})
	s.Require().NoError(err)

}

func (s *GovSuite) TestProposal() {
	// Test:
	// 1. Proposal submission
	// 2. Proposal deposit
	// 3. Vote
	// 4. Weighted vote

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Test Proposal", "Test Proposal", "ipfs://CID", govSubmissionDeposit+"uatom", s.DelegatorWallet.FormattedAddress(), false)
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
	submissionDepositUint, err := strconv.ParseUint(govSubmissionDeposit, 10, 64)
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
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(), "gov", "weighted-vote", proposalId, fmt.Sprintf("yes=%0.2f,no=%0.2f,abstain=%0.2f,no_with_veto=%0.2f", govYesWeight, govNoWeight, govAbstainWeight, govVetoWeight), "--gas", "auto")
	s.Require().NoError(err)

	// Test weighted vote
	vote, err = s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err)
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", vote)
	actual_yes_weight = vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_yes_weight.String())
	s.Require().Equal(float64(govYesWeight), actual_yes_weight.Float())
	actual_no_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_no_weight.String())
	s.Require().Equal(float64(govNoWeight), actual_no_weight.Float())
	actual_abstain_weight := vote.Get("options.#(option==\"VOTE_OPTION_ABSTAIN\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_abstain_weight.String())
	s.Require().Equal(float64(govAbstainWeight), actual_abstain_weight.Float())
	actual_veto_weight := vote.Get("options.#(option==\"VOTE_OPTION_NO_WITH_VETO\").weight")
	// chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", actual_veto_weight.String())
	s.Require().Equal(float64(govVetoWeight), actual_veto_weight.Float())
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
	}`, govCommunityPoolAmount, authority)

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
	fundAmount, err := strconv.ParseUint(govCommunityPoolAmount, 10, 64)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(balanceDifference, fundAmount)
}

func (s *GovSuite) TestVoteStakeValidation() {
	// Test that votes without sufficient stake are rejected

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Stake Validation Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID
	// print proposalId
	fmt.Println("Proposal ID:", proposalId)

	// Get status
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)

	currentStatus := proposal.Status.String()
	s.Require().Equal("PROPOSAL_STATUS_VOTING_PERIOD", currentStatus)

	// Submit vote from delegator with insufficient stake
	node := s.Chain.GetNode()
	// Delegate 1uatom from delegator to ensure insufficient stake
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet3.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, "2uatom")
	// Print delegator 3 delegations
	response, err := s.Chain.StakingQueryDelegations(s.GetContext(), s.DelegatorWallet3.FormattedAddress())
	s.Require().NoError(err)
	// Verify the delegation amount
	delegationBalance := response[0].Balance
	// Verify it's exactly 2uatom
	s.Require().Equal("2", delegationBalance.Amount.String())
	s.Require().Equal("uatom", delegationBalance.Denom)

	// Attempt to submit vote with insufficient stake
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(), "gov", "vote", proposalId, "yes", "--gas", "auto")
	s.Require().Error(err)

	// Attempt to submit weighted vote with insufficient stake
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(), "gov", "weighted-vote", proposalId, "yes=0.5,no=0.5", "--gas", "auto")
	s.Require().Error(err)

	err = node.StakingDelegate(s.GetContext(), s.DelegatorWallet3.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
	s.Require().NoError(err)

	// Attempt to submit vote with required stake
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(), "gov", "vote", proposalId, "yes", "--gas", "auto")
	s.Require().NoError(err)

	// Attempt to submit weighted vote with required stake
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(), "gov", "weighted-vote", proposalId, "yes=0.5,no=0.5", "--gas", "auto")
	s.Require().NoError(err)

}

func (s *GovSuite) TestAuthzVoteStakeValidation() {
	// Test that votes submitted through authz without sufficient stake are rejected

	// Unbond from delegator2 to ensure insufficient stake
	node := s.Chain.GetNode()
	err := node.StakingUnbond(s.GetContext(), s.DelegatorWallet2.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
	s.Require().NoError(err)

	// First grant delegation and vote authorization from delegator 2 to delegator 3

	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(), "authz", "grant", s.DelegatorWallet3.FormattedAddress(),
		"delegate", "--allowed-validators", s.Chain.ValidatorWallets[0].ValoperAddress, "--gas", "auto")
	s.Require().NoError(err)
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(), "authz", "grant", s.DelegatorWallet3.FormattedAddress(),
		"generic", "--msg-type", "/cosmos.gov.v1.MsgVote", "--gas", "auto")
	s.Require().NoError(err)
	_, err = node.ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(), "authz", "grant", s.DelegatorWallet3.FormattedAddress(),
		"generic", "--msg-type", "/cosmos.gov.v1.MsgVoteWeighted", "--gas", "auto")
	s.Require().NoError(err)

	// Verify grants to grantee
	grants, err := s.Chain.AuthzQueryGrants(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), s.DelegatorWallet3.FormattedAddress(), "")
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(grants), 3)
	fmt.Println("Authz grants from delegator 2 to delegator 3:", grants)

	// Delegate 2uatom from delegator2 via authz to ensure insufficient stake
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, "2uatom", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)
	// _, err = node.ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(), "authz", "exec", s.DelegatorWallet2.FormattedAddress(),
	// "staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, "2uatom", "--gas", "auto")
	// s.Require().NoError(err)
	// Print delegator2 delegations
	response, err := s.Chain.StakingQueryDelegations(s.GetContext(), s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err)
	// Verify the delegation amount
	delegationBalance := response[0].Balance
	// Verify it's exactly 2uatom
	s.Require().Equal("2", delegationBalance.Amount.String())
	s.Require().Equal("uatom", delegationBalance.Denom)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Stake Validation Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID
	// print proposalId
	fmt.Println("Proposal ID:", proposalId)

	// Get status
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)

	currentStatus := proposal.Status.String()
	s.Require().Equal("PROPOSAL_STATUS_VOTING_PERIOD", currentStatus)

	// Submit vote from delegator with insufficient stake via authz
	// Submit authz vote tx
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "gov", "vote", proposalId, "yes", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().Error(err)
	// Submit authz weighted-vote tx
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "gov", "weighted-vote", proposalId, "yes=0.5,no=0.5", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().Error(err)

	// Delegate 1 atom from delegator2 via authz to ensure insufficient stake
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, "1000000uatom", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)

	// Submit vote from delegator with required stake via authz
	// Submit authz vote tx
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "gov", "vote", proposalId, "yes", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)
	// Submit authz weighted-vote tx
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet3, "gov", "weighted-vote", proposalId, "yes=0.5,no=0.5", "--from", string(s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)
}

func (s *GovSuite) TestWasmVoteStakeValidation() {
	// Test that votes submitted through cosmwasm contracts without sufficient stake are rejected

	// Delegate 2uatom from contract address
	// Excecute 'delegate' on contract
	jsonDelegate := fmt.Sprintf(`{
		"delegate": {
			"validator": "%s",
			"amount": {
				"denom": "%s",
				"amount": "2"
			}
		}
	}`, s.Chain.ValidatorWallets[0].ValoperAddress, s.Chain.Config().Denom)

	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "execute", s.contractAddress, jsonDelegate, "--gas", "auto",
	)
	s.Require().NoError(err)

	// Verify delegation
	response, err := s.Chain.StakingQueryDelegations(s.GetContext(), s.contractAddress)
	s.Require().NoError(err)
	// Verify the delegation amount
	delegationBalance := response[0].Balance
	// Verify it's exactly 2uatom
	s.Require().Equal("2", delegationBalance.Amount.String())
	s.Require().Equal("uatom", delegationBalance.Denom)
	// Print delegations
	fmt.Println("Contract Delegations:", response)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Stake Validation Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID
	// print proposalId
	fmt.Println("Proposal ID:", proposalId)

	// Get status
	proposalIDuint, err := strconv.ParseUint(proposalId, 10, 64)
	proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), proposalIDuint)
	s.Require().NoError(err)

	currentStatus := proposal.Status.String()
	s.Require().Equal("PROPOSAL_STATUS_VOTING_PERIOD", currentStatus)

	// Submit vote from contract with insufficient stake
	jsonVote := fmt.Sprintf(`{
		"vote": {
			"proposal_id": %s,
			"option": "yes"
		}
	}`, proposalId)

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "execute", s.contractAddress, jsonVote, "--gas", "auto",
	)
	s.Require().Error(err)

	// Print tally after the vote
	tallyAfter, err := s.Chain.QueryJSON(s.GetContext(), "tally", "gov", "tally", proposalId)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Tally after contract vote with insufficient stake: %s", tallyAfter.String())

	// Delegate more tokens from contract to reach required stake
	jsonDelegate = fmt.Sprintf(`{
		"delegate": {
			"validator": "%s",
			"amount": {
				"denom": "%s",
				"amount": "%s"
			}
		}
	}`, s.Chain.ValidatorWallets[0].ValoperAddress, s.Chain.Config().Denom, "1000000")

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "execute", s.contractAddress, jsonDelegate, "--gas", "auto",
	)
	s.Require().NoError(err)

	// Submit vote from contract with sufficient stake
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet3.KeyName(),
		"wasm", "execute", s.contractAddress, jsonVote, "--gas", "auto",
	)
	s.Require().NoError(err)

	// Print tally after the vote
	tallyAfter, err = s.Chain.QueryJSON(s.GetContext(), "tally", "gov", "tally", proposalId)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Tally after contract vote with required stake: %s", tallyAfter.String())

	// Test vote was counted
	vote, err := s.Chain.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.contractAddress)
	s.Require().NoError(err)
	// Print vote
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Contract Vote: %s", vote.String())
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	s.Require().Equal(float64(1.0), actual_yes_weight.Float())
}

func TestGovModule(t *testing.T) {
	// Use permissionless wasm params
	wasmGenesis := append(chainsuite.DefaultGenesis(),
		cosmos.NewGenesisKV("app_state.wasm.params.code_upload_access.permission", "Everybody"),
		cosmos.NewGenesisKV("app_state.wasm.params.instantiate_default_permission", "Everybody"),
	)
	// Create custom chain spec
	chainSpec := chainsuite.DefaultChainSpec(chainsuite.GetEnvironment())
	chainSpec.ChainConfig.ModifyGenesis = cosmos.ModifyGenesis(wasmGenesis)

	s := &GovSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		ChainSpec:      chainSpec,
		UpgradeOnSetup: false,
		CreateRelayer:  true,
	})}}
	suite.Run(t, s)
}

func (s GovSuite) authzGenExec(ctx context.Context, grantee ibc.Wallet, command ...string) error {
	txjson, err := s.Chain.GenerateTx(ctx, 0, command...)
	s.Require().NoError(err)

	// WriteFile expects a relative path from the node's home directory
	err = s.Chain.GetNode().WriteFile(ctx, []byte(txjson), "tx.json")
	s.Require().NoError(err)

	// ExecTx needs the full path within the container
	txFilePath := path.Join(s.Chain.GetNode().HomeDir(), "tx.json")
	_, err = s.Chain.GetNode().ExecTx(
		ctx,
		grantee.KeyName(),
		"authz", "exec", txFilePath,
	)
	return err
}
