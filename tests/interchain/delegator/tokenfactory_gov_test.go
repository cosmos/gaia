package delegator_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryGovSuite struct {
	*TokenFactoryBaseSuite
}

func (s *TokenFactoryGovSuite) SetupSuite() {
	s.Suite.SetupSuite()

	// Delegate some tokens to have voting power
	node := s.Chain.GetNode()
	stakeAmount := "10000000" + s.Chain.Config().Denom
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(),
		s.Chain.ValidatorWallets[0].ValoperAddress, stakeAmount)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(),
		s.Chain.ValidatorWallets[0].ValoperAddress, stakeAmount)
}

// TestParamChangeCreationFee tests changing the denom creation fee via governance
func (s *TokenFactoryGovSuite) TestParamChangeCreationFee() {
	// Query current params
	params, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)

	// Get current creation fee
	currentFee := params.Get("denom_creation_fee.0.amount").String()
	s.Require().NotEmpty(currentFee)

	// Propose new fee (double the current fee)
	currentFeeInt, ok := sdkmath.NewIntFromString(currentFee)
	s.Require().True(ok)
	newFee := currentFeeInt.MulRaw(2)

	// Get gov module authority address
	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	// Create MsgUpdateParams message as JSON
	paramChangeMessage := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgUpdateParams",
		"authority": "%s",
		"params": {
			"denom_creation_fee": [
				{
					"denom": "uatom",
					"amount": "%s"
				}
			]
		}
	}`, authority, newFee.String())

	// Create proposal JSON (workaround for interchaintest decoding issue)
	type ProposalJSON struct {
		Messages       []json.RawMessage `json:"messages"`
		InitialDeposit string            `json:"deposit"`
		Title          string            `json:"title"`
		Summary        string            `json:"summary"`
		Metadata       string            `json:"metadata"`
	}

	proposal := ProposalJSON{
		Messages:       []json.RawMessage{json.RawMessage(paramChangeMessage)},
		InitialDeposit: fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		Title:          "Change TokenFactory Creation Fee",
		Summary:        "Increase denom creation fee to " + newFee.String(),
		Metadata:       "ipfs://CID",
	}

	// Write proposal to file
	proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(s.GetContext(), proposalBytes, "tokenfactory-proposal.json")
	s.Require().NoError(err)

	proposalPath := s.Chain.GetNode().HomeDir() + "/tokenfactory-proposal.json"

	// Submit proposal using ExecTx to avoid interchaintest decoding
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"gov", "submit-proposal", proposalPath)
	s.Require().NoError(err)

	// Query for the last proposal ID
	lastProposal, err := s.Chain.QueryJSON(s.GetContext(), "proposals.@reverse.0.id", "gov", "proposals")
	s.Require().NoError(err)
	proposalID := lastProposal.String()

	// Pass the proposal
	err = s.Chain.PassProposal(s.GetContext(), proposalID)
	s.Require().NoError(err)

	// Wait for proposal to be executed
	s.Require().Eventually(func() bool {
		proposal, err := s.Chain.GovQueryProposalV1(s.GetContext(),
			mustParseUint(proposalID))
		if err != nil {
			return false
		}
		return proposal.Status == govv1.StatusPassed
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
		"proposal did not pass")

	// Query params again and verify change
	updatedParams, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)

	newFeeStr := updatedParams.Get("denom_creation_fee.0.amount").String()
	s.Require().Equal(newFee.String(), newFeeStr)

	// Verify new fee is charged
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet2.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "create-denom", "newfeetest",
	)
	s.Require().NoError(err)

	balanceAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet2.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	// Balance should have decreased by at least the new fee
	s.Require().True(balanceBefore.Sub(balanceAfter).GTE(newFee))
}

// TestMetadataModification tests setting metadata on tokenfactory denoms
func (s *TokenFactoryGovSuite) TestMetadataModification() {
	denom, err := s.CreateDenom(s.DelegatorWallet, "metadata")
	s.Require().NoError(err, "failed to create denom 'metadata'")
	err = s.Mint(s.DelegatorWallet, denom, 1000000)
	s.Require().NoError(err, "failed to mint tokens for metadata test")

	// Set metadata
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "modify-metadata",
		denom,
		"META",                 // ticker-symbol
		"Test metadata token",  // description
		"6",                    // exponent
	)
	s.Require().NoError(err)

	// Query and verify metadata
	metadata, err := s.Chain.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal("META", metadata.Get("symbol").String())
	s.Require().Equal("Test metadata token", metadata.Get("description").String())
}

// TestAdminChange tests transferring admin privileges for tokenfactory denoms
func (s *TokenFactoryGovSuite) TestAdminChange() {
	denom, err := s.CreateDenom(s.DelegatorWallet, "adminchange")
	s.Require().NoError(err, "failed to create denom 'adminchange'")
	err = s.Mint(s.DelegatorWallet, denom, 1000000)
	s.Require().NoError(err, "failed to mint tokens for admin change test")

	// Verify original admin can mint
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("500000%s", denom),
	)
	s.Require().NoError(err)

	// Change admin to DelegatorWallet2
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom, s.DelegatorWallet2.FormattedAddress(),
	)
	s.Require().NoError(err)

	// Verify admin was changed
	admin, err := s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet2.FormattedAddress(), admin.String())

	// Verify new admin can mint
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("300000%s", denom),
	)
	s.Require().NoError(err)

	// Verify old admin can no longer mint
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("100000%s", denom),
	)
	s.Require().Error(err)
}

// TestCreateDenomAfterUpgrade tests that new denoms can be created after upgrade
func (s *TokenFactoryGovSuite) TestCreateDenomAfterUpgrade() {
	// Create new denom
	denom, err := s.CreateDenom(s.DelegatorWallet, "postupgrade")
	s.Require().NoError(err, "failed to create denom 'postupgrade'")

	// Mint tokens
	err = s.Mint(s.DelegatorWallet, denom, 5000000)
	s.Require().NoError(err, "failed to mint tokens for postupgrade test")

	// Verify balance
	balance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(5000000), balance)

	// Verify all operations work
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "burn",
		fmt.Sprintf("1000000%s", denom),
	)
	s.Require().NoError(err)

	balance, err = s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(4000000), balance)
}

// TestParamQueryAfterUpgrade tests that params can be queried after upgrade
func (s *TokenFactoryGovSuite) TestParamQueryAfterUpgrade() {
	// Query params
	paramsAfter, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)
	s.Require().NotNil(paramsAfter)

	// Verify params are still accessible
	creationFee := paramsAfter.Get("denom_creation_fee.0.amount").String()
	s.Require().NotEmpty(creationFee)
}

// TestGovernanceProposalWithTokenFactoryToken tests using tokenfactory tokens in governance
func (s *TokenFactoryGovSuite) TestGovernanceProposalWithTokenFactoryToken() {
	// This test verifies that tokenfactory tokens exist and work with governance
	// even if they can't be used as staking tokens

	// Create tokenfactory denom
	denom, err := s.CreateDenom(s.DelegatorWallet, "govtoken")
	s.Require().NoError(err, "failed to create denom 'govtoken'")
	err = s.Mint(s.DelegatorWallet, denom, 100000000)
	s.Require().NoError(err, "failed to mint tokens for gov token test")

	// Submit a text proposal (normal governance with ATOM)
	prop, err := s.Chain.BuildProposal(
		nil,
		"Test with TokenFactory",
		"Testing governance while tokenfactory tokens exist",
		"ipfs://CID",
		fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		s.DelegatorWallet.FormattedAddress(),
		false,
	)
	s.Require().NoError(err)

	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalID := result.ProposalID

	// Vote on proposal
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"gov", "vote", proposalID, "yes",
	)
	s.Require().NoError(err)

	// Verify vote was recorded
	vote, err := s.Chain.QueryJSON(s.GetContext(),
		"vote", "gov", "vote", proposalID, s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	s.Require().Equal(float64(1.0), actual_yes_weight.Float())

	// Verify tokenfactory token still exists and works
	balance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(100000000), balance)
}

// Helper function to parse uint
func mustParseUint(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return val
}

func TestTokenFactoryGov(t *testing.T) {
	s := &TokenFactoryGovSuite{
		TokenFactoryBaseSuite: &TokenFactoryBaseSuite{
			Suite: &delegator.Suite{
				Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
					UpgradeOnSetup: true, // Upgrade to v26 before running tests
				}),
			},
		},
	}
	suite.Run(t, s)
}
