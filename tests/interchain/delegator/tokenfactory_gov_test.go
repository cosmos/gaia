package delegator_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryGovSuite struct {
	*delegator.Suite
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

// createDenom creates a tokenfactory denom
func (s *TokenFactoryGovSuite) createDenom(subdenom string) string {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)
}

// mint mints tokens
func (s *TokenFactoryGovSuite) mint(denom string, amount int64) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	s.Require().NoError(err)
}

// TestParamChangeCreationFee tests changing the denom creation fee via governance
func (s *TokenFactoryGovSuite) TestParamChangeCreationFee() {
	// Query current params
	params, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)

	// Get current creation fee
	currentFee := params.Get("params.denom_creation_fee.0.amount").String()
	s.Require().NotEmpty(currentFee)

	// Propose new fee (double the current fee)
	currentFeeInt, ok := sdkmath.NewIntFromString(currentFee)
	s.Require().True(ok)
	newFee := currentFeeInt.MulRaw(2)

	// Submit proposal
	// Note: Actual param change implementation depends on tokenfactory governance integration
	prop, err := s.Chain.BuildProposal(
		[]cosmos.ProtoMessage{},
		"Change TokenFactory Creation Fee",
		"Increase denom creation fee to "+newFee.String(),
		"ipfs://CID",
		fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		s.DelegatorWallet.FormattedAddress(),
		false,
	)
	s.Require().NoError(err)

	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalID := result.ProposalID

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

	newFeeStr := updatedParams.Get("params.denom_creation_fee.0.amount").String()
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

// TestUpgradePreservesState tests that tokenfactory state is preserved through upgrades
func (s *TokenFactoryGovSuite) TestUpgradePreservesState() {
	// Create multiple denoms with various states before upgrade
	denom1 := s.createDenom("upgrade1")
	denom2 := s.createDenom("upgrade2")
	denom3 := s.createDenom("upgrade3")

	// Mint different amounts
	s.mint(denom1, 1000000)
	s.mint(denom2, 2000000)
	s.mint(denom3, 3000000)

	// Set metadata for denom1
	metadataJSON := fmt.Sprintf(`{
		"base": "%s",
		"display": "upgrade1",
		"name": "Upgrade Test Token",
		"symbol": "UPG1",
		"description": "Token to test upgrade persistence"
	}`, denom1)

	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "set-denom-metadata",
		denom1,
		metadataJSON,
	)
	s.Require().NoError(err)

	// Change admin for denom2
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom2, s.DelegatorWallet2.FormattedAddress(),
	)
	s.Require().NoError(err)

	// Renounce admin for denom3
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom3, "",
	)
	s.Require().NoError(err)

	// Record state before upgrade
	balance1Before, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom1)
	s.Require().NoError(err)

	balance2Before, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom2)
	s.Require().NoError(err)

	balance3Before, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom3)
	s.Require().NoError(err)

	metadata1Before, err := s.Chain.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", denom1)
	s.Require().NoError(err)

	admin1Before, err := s.Chain.QueryJSON(s.GetContext(),
		"admin", "tokenfactory", "denom-authority-metadata", denom1)
	s.Require().NoError(err)

	admin2Before, err := s.Chain.QueryJSON(s.GetContext(),
		"admin", "tokenfactory", "denom-authority-metadata", denom2)
	s.Require().NoError(err)

	// Perform upgrade
	s.UpgradeChain()

	// Verify balances preserved
	balance1After, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom1)
	s.Require().NoError(err)
	s.Require().Equal(balance1Before, balance1After)

	balance2After, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom2)
	s.Require().NoError(err)
	s.Require().Equal(balance2Before, balance2After)

	balance3After, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom3)
	s.Require().NoError(err)
	s.Require().Equal(balance3Before, balance3After)

	// Verify metadata preserved
	metadata1After, err := s.Chain.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", denom1)
	s.Require().NoError(err)
	s.Require().Equal(
		metadata1Before.Get("metadata.name").String(),
		metadata1After.Get("metadata.name").String())
	s.Require().Equal(
		metadata1Before.Get("metadata.symbol").String(),
		metadata1After.Get("metadata.symbol").String())

	// Verify admin assignments preserved
	admin1After, err := s.Chain.QueryJSON(s.GetContext(),
		"admin", "tokenfactory", "denom-authority-metadata", denom1)
	s.Require().NoError(err)
	s.Require().Equal(
		admin1Before.Get("authority_metadata.admin").String(),
		admin1After.Get("authority_metadata.admin").String())

	admin2After, err := s.Chain.QueryJSON(s.GetContext(),
		"admin", "tokenfactory", "denom-authority-metadata", denom2)
	s.Require().NoError(err)
	s.Require().Equal(
		admin2Before.Get("authority_metadata.admin").String(),
		admin2After.Get("authority_metadata.admin").String())
	s.Require().Equal(s.DelegatorWallet2.FormattedAddress(),
		admin2After.Get("authority_metadata.admin").String())

	admin3After, err := s.Chain.QueryJSON(s.GetContext(),
		"admin", "tokenfactory", "denom-authority-metadata", denom3)
	s.Require().NoError(err)
	s.Require().Empty(admin3After.Get("authority_metadata.admin").String())

	// Verify operations still work post-upgrade

	// Mint with denom1 (original admin still works)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("500000%s", denom1),
	)
	s.Require().NoError(err)

	balance1Final, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom1)
	s.Require().NoError(err)
	s.Require().Equal(balance1Before.Add(sdkmath.NewInt(500000)), balance1Final)

	// Mint with denom2 (new admin works)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("500000%s", denom2),
	)
	s.Require().NoError(err)

	// Verify denom3 still cannot be minted (renounced admin)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("500000%s", denom3),
	)
	s.Require().Error(err)

	// Verify denom3 can still be transferred
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("1000000%s", denom3),
	)
	s.Require().NoError(err)
}

// TestCreateDenomAfterUpgrade tests that new denoms can be created after upgrade
func (s *TokenFactoryGovSuite) TestCreateDenomAfterUpgrade() {
	// Perform upgrade first
	s.UpgradeChain()

	// Create new denom after upgrade
	denom := s.createDenom("postupgrade")

	// Mint tokens
	s.mint(denom, 5000000)

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
	// Query params before upgrade
	paramsBefore, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)
	s.Require().NotNil(paramsBefore)

	// Perform upgrade
	s.UpgradeChain()

	// Query params after upgrade
	paramsAfter, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)
	s.Require().NotNil(paramsAfter)

	// Verify params are still accessible
	creationFee := paramsAfter.Get("params.denom_creation_fee.0.amount").String()
	s.Require().NotEmpty(creationFee)
}

// TestGovernanceProposalWithTokenFactoryToken tests using tokenfactory tokens in governance
func (s *TokenFactoryGovSuite) TestGovernanceProposalWithTokenFactoryToken() {
	// This test verifies that tokenfactory tokens exist and work with governance
	// even if they can't be used as staking tokens

	// Create tokenfactory denom
	denom := s.createDenom("govtoken")
	s.mint(denom, 100000000)

	// Submit a text proposal (normal governance with ATOM)
	prop, err := s.Chain.BuildProposal(
		[]cosmos.ProtoMessage{},
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
	s.Require().Equal("VOTE_OPTION_YES",
		vote.Get("vote.options.0.option").String())

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
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: false, // We'll upgrade manually in tests
			}),
		},
	}
	suite.Run(t, s)
}
