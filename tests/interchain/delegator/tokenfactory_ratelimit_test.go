package delegator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryRateLimitSuite struct {
	*TokenFactoryBaseSuite
	ChainB       *chainsuite.Chain
	ChainBWallet ibc.Wallet
}

func (s *TokenFactoryRateLimitSuite) SetupSuite() {
	s.Suite.SetupSuite()

	// Delegate some tokens to have voting power for governance proposals
	node := s.Chain.GetNode()
	stakeAmount := "10000000" + s.Chain.Config().Denom
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(),
		s.Chain.ValidatorWallets[0].ValoperAddress, stakeAmount)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(),
		s.Chain.ValidatorWallets[0].ValoperAddress, stakeAmount)

	// Add a second chain for IBC testing
	chainB, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.ChainB = chainB

	// Create wallet on chain B
	wallet, err := chainB.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.ChainBWallet = wallet

	// Fund wallet on chain B
	err = chainB.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.ChainBWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   chainB.Config().Denom,
	})
	s.Require().NoError(err)
}

// Helper functions

// addRateLimit creates a governance proposal to add rate limit
func (s *TokenFactoryRateLimitSuite) addRateLimit(
	ctx context.Context,
	denom string,
	channel string,
	sendPercent, recvPercent int64,
	durationHours uint64,
) string {
	// Get gov module authority address
	authority, err := s.Chain.GetGovernanceAddress(ctx)
	s.Require().NoError(err)

	// Create MsgAddRateLimit message as JSON
	// Note: field is "channel_or_client_id" not "channel_id" (supports both v1 and v2 IBC)
	rateLimitMessage := fmt.Sprintf(`{
		"@type": "/ratelimit.v1.MsgAddRateLimit",
		"authority": "%s",
		"denom": "%s",
		"channel_or_client_id": "%s",
		"max_percent_send": "%d",
		"max_percent_recv": "%d",
		"duration_hours": "%d"
	}`, authority, denom, channel, sendPercent, recvPercent, durationHours)

	// Create proposal using ProposalJSON struct
	proposal := ProposalJSON{
		Messages:       []json.RawMessage{json.RawMessage(rateLimitMessage)},
		InitialDeposit: fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		Title:          "Add Rate Limit for " + denom,
		Summary:        fmt.Sprintf("Add %d%% send, %d%% recv quota", sendPercent, recvPercent),
		Metadata:       "ipfs://CID",
	}

	// Write proposal to file
	proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, proposalBytes, "ratelimit-proposal.json")
	s.Require().NoError(err)

	proposalPath := s.Chain.GetNode().HomeDir() + "/ratelimit-proposal.json"

	// Submit proposal using ExecTx
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.FormattedAddress(),
		"gov", "submit-proposal", proposalPath)
	s.Require().NoError(err)

	// Query for the last proposal ID
	lastProposal, err := s.Chain.QueryJSON(ctx, "proposals.@reverse.0.id", "gov", "proposals")
	s.Require().NoError(err)
	proposalID := lastProposal.String()

	// Pass the proposal
	err = s.Chain.PassProposal(ctx, proposalID)
	s.Require().NoError(err)

	// Wait for proposal to be executed
	s.Require().Eventually(func() bool {
		proposal, err := s.Chain.GovQueryProposalV1(ctx, mustParseUint(proposalID))
		if err != nil {
			return false
		}
		return proposal.Status == govv1.StatusPassed
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
		"proposal did not pass")

	return proposalID
}

// updateRateLimit creates a governance proposal to update an existing rate limit
func (s *TokenFactoryRateLimitSuite) updateRateLimit(
	ctx context.Context,
	denom string,
	channel string,
	sendPercent, recvPercent int64,
	durationHours uint64,
) string {
	// Get gov module authority address
	authority, err := s.Chain.GetGovernanceAddress(ctx)
	s.Require().NoError(err)

	// Create MsgUpdateRateLimit message as JSON
	rateLimitMessage := fmt.Sprintf(`{
		"@type": "/ratelimit.v1.MsgUpdateRateLimit",
		"authority": "%s",
		"denom": "%s",
		"channel_or_client_id": "%s",
		"max_percent_send": "%d",
		"max_percent_recv": "%d",
		"duration_hours": "%d"
	}`, authority, denom, channel, sendPercent, recvPercent, durationHours)

	// Create proposal using ProposalJSON struct
	proposal := ProposalJSON{
		Messages:       []json.RawMessage{json.RawMessage(rateLimitMessage)},
		InitialDeposit: fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		Title:          "Update Rate Limit for " + denom,
		Summary:        fmt.Sprintf("Update to %d%% send, %d%% recv quota", sendPercent, recvPercent),
		Metadata:       "ipfs://CID",
	}

	// Write proposal to file
	proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, proposalBytes, "ratelimit-update-proposal.json")
	s.Require().NoError(err)

	proposalPath := s.Chain.GetNode().HomeDir() + "/ratelimit-update-proposal.json"

	// Submit proposal using ExecTx
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.FormattedAddress(),
		"gov", "submit-proposal", proposalPath)
	s.Require().NoError(err)

	// Query for the last proposal ID
	lastProposal, err := s.Chain.QueryJSON(ctx, "proposals.@reverse.0.id", "gov", "proposals")
	s.Require().NoError(err)
	proposalID := lastProposal.String()

	// Pass the proposal
	err = s.Chain.PassProposal(ctx, proposalID)
	s.Require().NoError(err)

	// Wait for proposal to be executed
	s.Require().Eventually(func() bool {
		proposal, err := s.Chain.GovQueryProposalV1(ctx, mustParseUint(proposalID))
		if err != nil {
			return false
		}
		return proposal.Status == govv1.StatusPassed
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
		"proposal did not pass")

	return proposalID
}

// removeRateLimit creates a governance proposal to remove rate limit
func (s *TokenFactoryRateLimitSuite) removeRateLimit(
	ctx context.Context,
	denom string,
	channel string,
) string {
	// Get gov module authority address
	authority, err := s.Chain.GetGovernanceAddress(ctx)
	s.Require().NoError(err)

	// Create MsgRemoveRateLimit message as JSON
	rateLimitMessage := fmt.Sprintf(`{
		"@type": "/ratelimit.v1.MsgRemoveRateLimit",
		"authority": "%s",
		"denom": "%s",
		"channel_or_client_id": "%s"
	}`, authority, denom, channel)

	// Create proposal using ProposalJSON struct
	proposal := ProposalJSON{
		Messages:       []json.RawMessage{json.RawMessage(rateLimitMessage)},
		InitialDeposit: fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		Title:          "Remove Rate Limit for " + denom,
		Summary:        "Remove rate limit quota",
		Metadata:       "ipfs://CID",
	}

	// Write proposal to file
	proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, proposalBytes, "ratelimit-removal-proposal.json")
	s.Require().NoError(err)

	proposalPath := s.Chain.GetNode().HomeDir() + "/ratelimit-removal-proposal.json"

	// Submit proposal using ExecTx
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.FormattedAddress(),
		"gov", "submit-proposal", proposalPath)
	s.Require().NoError(err)

	// Query for the last proposal ID
	lastProposal, err := s.Chain.QueryJSON(ctx, "proposals.@reverse.0.id", "gov", "proposals")
	s.Require().NoError(err)
	proposalID := lastProposal.String()

	// Pass the proposal
	err = s.Chain.PassProposal(ctx, proposalID)
	s.Require().NoError(err)

	// Wait for proposal to be executed
	s.Require().Eventually(func() bool {
		proposal, err := s.Chain.GovQueryProposalV1(ctx, mustParseUint(proposalID))
		if err != nil {
			return false
		}
		return proposal.Status == govv1.StatusPassed
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
		"proposal did not pass")

	return proposalID
}

// calculateQuota computes max transfer amount based on percentage
func (s *TokenFactoryRateLimitSuite) calculateQuota(
	totalSupply sdkmath.Int,
	percent int64,
) sdkmath.Int {
	return totalSupply.Mul(sdkmath.NewInt(percent)).Quo(sdkmath.NewInt(100))
}

// attemptTransfer attempts an IBC transfer and returns any error
func (s *TokenFactoryRateLimitSuite) attemptTransfer(
	ctx context.Context,
	from *chainsuite.Chain,
	sender ibc.Wallet,
	recipient string,
	amount sdkmath.Int,
	denom string,
	channelID string,
) error {
	_, err := from.GetNode().ExecTx(
		ctx,
		sender.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		channelID,
		recipient,
		fmt.Sprintf("%s%s", amount.String(), denom),
	)
	return err
}

// verifyBalance checks that a balance matches the expected value
func (s *TokenFactoryRateLimitSuite) verifyBalance(
	ctx context.Context,
	chain *chainsuite.Chain,
	address string,
	denom string,
	expected sdkmath.Int,
	message string,
) error {
	balance, err := chain.GetBalance(ctx, address, denom)
	if err != nil {
		return err
	}
	if !balance.Equal(expected) {
		return fmt.Errorf("%s: expected %s, got %s", message, expected, balance)
	}
	return nil
}

// Test 1: TestRateLimitBasicTransfer
func (s *TokenFactoryRateLimitSuite) TestRateLimitBasicTransfer() {
	ctx := s.GetContext()

	// Create tokenfactory denom on chain A
	denom, err := s.CreateDenom(s.DelegatorWallet, "ratelimitcoin")
	s.Require().NoError(err, "failed to create denom 'ratelimitcoin'")

	// Mint 1,000,000 tokens
	totalSupply := int64(1_000_000)
	err = s.Mint(s.DelegatorWallet, denom, totalSupply)
	s.Require().NoError(err, "failed to mint %d tokens for denom %s", totalSupply, denom)

	// Verify initial balance
	err = s.verifyBalance(ctx, s.Chain, s.DelegatorWallet.FormattedAddress(), denom,
		sdkmath.NewInt(totalSupply), "initial balance should equal minted amount")
	s.Require().NoError(err)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Set rate limit via governance: 10% send quota over 24 hours
	s.addRateLimit(ctx, denom, transferCh.ChannelID, 10, 10, 24)

	// Transfer 50,000 tokens (5%) to ChainB → should PASS
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferCh.ChannelID)
	s.Require().NoError(err, "first transfer (5%%) should succeed")

	// Wait for transfer to complete
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(ctx, s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(amount1), "chain B should receive first transfer")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Transfer another 50,000 tokens (5%) → should PASS (10% total)
	amount2 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount2, denom, transferCh.ChannelID)
	s.Require().NoError(err, "second transfer (5%%) should succeed, total 10%%")

	// Wait for second transfer
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(ctx, s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(amount1.Add(amount2)), "chain B should receive both transfers")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Get balance before failed transfer attempt
	balanceBefore, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer 10,000 more tokens → should FAIL (would exceed 10%)
	amount3 := s.calculateQuota(sdkmath.NewInt(totalSupply), 1)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount3, denom, transferCh.ChannelID)
	s.Require().Error(err, "third transfer should fail (would exceed 10%% quota)")
	s.Require().Contains(err.Error(), "quota", "error should mention quota/rate limit")

	// Verify balance unchanged after failed transfer
	balanceAfter, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore, balanceAfter, "balance should be unchanged after failed transfer")
}

// Test 2: TestRateLimitBidirectional
func (s *TokenFactoryRateLimitSuite) TestRateLimitBidirectional() {
	ctx := s.GetContext()

	// Create tokenfactory denom on chain A, mint 1M tokens
	denom, err := s.CreateDenom(s.DelegatorWallet, "bidir")
	s.Require().NoError(err, "failed to create denom 'bidir'")
	totalSupply := int64(1_000_000)
	err = s.Mint(s.DelegatorWallet, denom, totalSupply)
	s.Require().NoError(err, "failed to mint %d tokens for denom %s", totalSupply, denom)

	// Get IBC transfer channel
	transferChAtoB, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate expected IBC denom on chain B
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferChAtoB.Counterparty.PortID,
		transferChAtoB.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Set rate limit: 10% send, 5% receive over 24 hours
	s.addRateLimit(ctx, denom, transferChAtoB.ChannelID, 10, 5, 24)

	// Transfer 100k from A→B (10% outflow) → should PASS
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 10)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferChAtoB.ChannelID)
	s.Require().NoError(err, "first transfer A→B (10%%) should succeed")

	// Wait for transfer
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(ctx, s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(amount1), "chain B should receive transfer")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Transfer 100k more from A→B → should FAIL (exceed send quota)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferChAtoB.ChannelID)
	s.Require().Error(err, "second transfer A→B should fail (exceed send quota)")

	// Get return channel from B to A
	transferChBtoA, err := s.Relayer.GetTransferChannel(ctx, s.ChainB, s.Chain)
	s.Require().NoError(err)

	// Get balance on chain A before return transfer
	balanceABefore, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer 80k from B→A (8% inflow on A, net inflow = 80k - 100k = -20k) → should PASS
	// Note: Rate limits track NET flow (inflow - outflow), not absolute flow
	amount2 := s.calculateQuota(sdkmath.NewInt(totalSupply), 8)
	err = s.attemptTransfer(ctx, s.ChainB, s.ChainBWallet,
		s.DelegatorWallet.FormattedAddress(), amount2, expectedDenomB, transferChBtoA.ChannelID)
	s.Require().NoError(err, "return transfer B→A (8%%) should succeed (net inflow -20%%)")

	// Wait for return transfer
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.True(c, balance.Sub(balanceABefore).Equal(amount2), "chain A should receive return transfer")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Transfer 80k more from B→A (total inflow 160k, net inflow = 160k - 100k = 60k)
	// Net inflow 60k > 5% quota (50k) → should FAIL
	amount3 := s.calculateQuota(sdkmath.NewInt(totalSupply), 8)
	err = s.attemptTransfer(ctx, s.ChainB, s.ChainBWallet,
		s.DelegatorWallet.FormattedAddress(), amount3, expectedDenomB, transferChBtoA.ChannelID)
	s.Require().Error(err, "second return transfer should fail (net inflow would exceed 5%% receive quota)")
}

// Test 3: TestRateLimitMultipleUsers
func (s *TokenFactoryRateLimitSuite) TestRateLimitMultipleUsers() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	denom, err := s.CreateDenom(s.DelegatorWallet, "multiuser")
	s.Require().NoError(err, "failed to create denom 'multiuser'")
	totalSupply := int64(1_000_000)
	err = s.Mint(s.DelegatorWallet, denom, totalSupply)
	s.Require().NoError(err, "failed to mint %d tokens for denom %s", totalSupply, denom)

	// Distribute tokens to DelegatorWallet2 and create a third wallet
	wallet3, err := s.Chain.BuildWallet(ctx, "user3", "")
	s.Require().NoError(err)

	// Fund wallet3 with native tokens for gas
	err = s.Chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: wallet3.FormattedAddress(),
		Amount:  sdkmath.NewInt(10_000_000),
		Denom:   s.Chain.Config().Denom,
	})
	s.Require().NoError(err)

	// Distribute tokenfactory tokens to all wallets
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 30)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(), fmt.Sprintf("%s%s", amount1.String(), denom))
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		wallet3.FormattedAddress(), fmt.Sprintf("%s%s", amount1.String(), denom))
	s.Require().NoError(err)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Set 10% rate limit
	s.addRateLimit(ctx, denom, transferCh.ChannelID, 10, 10, 24)

	// User1 transfers 6% → should PASS
	user1Amount := s.calculateQuota(sdkmath.NewInt(totalSupply), 6)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), user1Amount, denom, transferCh.ChannelID)
	s.Require().NoError(err, "user1 transfer (6%%) should succeed")

	// User2 transfers 3% → should PASS (9% cumulative)
	user2Amount := s.calculateQuota(sdkmath.NewInt(totalSupply), 3)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet2,
		s.ChainBWallet.FormattedAddress(), user2Amount, denom, transferCh.ChannelID)
	s.Require().NoError(err, "user2 transfer (3%%) should succeed (9%% cumulative)")

	// User3 transfers 2% → should FAIL (would be 11% cumulative)
	user3Amount := s.calculateQuota(sdkmath.NewInt(totalSupply), 2)
	err = s.attemptTransfer(ctx, s.Chain, wallet3,
		s.ChainBWallet.FormattedAddress(), user3Amount, denom, transferCh.ChannelID)
	s.Require().Error(err, "user3 transfer should fail (would be 11%% cumulative)")
}

// Test 4: TestRateLimitGovernanceUpdate
func (s *TokenFactoryRateLimitSuite) TestRateLimitGovernanceUpdate() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	denom, err := s.CreateDenom(s.DelegatorWallet, "govupdate")
	s.Require().NoError(err, "failed to create denom 'govupdate'")
	totalSupply := int64(1_000_000)
	err = s.Mint(s.DelegatorWallet, denom, totalSupply)
	s.Require().NoError(err, "failed to mint %d tokens for denom %s", totalSupply, denom)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Set initial 5% rate limit
	s.addRateLimit(ctx, denom, transferCh.ChannelID, 5, 5, 24)

	// Transfer 4% → should PASS
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 4)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferCh.ChannelID)
	s.Require().NoError(err, "transfer (4%%) should succeed")

	// Transfer 2% → should FAIL (exceeds 5%)
	amount2 := s.calculateQuota(sdkmath.NewInt(totalSupply), 2)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount2, denom, transferCh.ChannelID)
	s.Require().Error(err, "transfer should fail (exceeds 5%% quota)")

	// Submit governance proposal to update rate limit from 5% to 10%
	s.updateRateLimit(ctx, denom, transferCh.ChannelID, 10, 10, 24)

	// Transfer 5% more → should PASS (9% total, within new 10% limit)
	amount3 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount3, denom, transferCh.ChannelID)
	s.Require().NoError(err, "transfer (5%%) should succeed after quota increase (9%% total)")
}

// Test 5: TestRateLimitRemoval
func (s *TokenFactoryRateLimitSuite) TestRateLimitRemoval() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	denom, err := s.CreateDenom(s.DelegatorWallet, "removal")
	s.Require().NoError(err, "failed to create denom 'removal'")
	totalSupply := int64(1_000_000)
	err = s.Mint(s.DelegatorWallet, denom, totalSupply)
	s.Require().NoError(err, "failed to mint %d tokens for denom %s", totalSupply, denom)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Set 5% rate limit
	s.addRateLimit(ctx, denom, transferCh.ChannelID, 5, 5, 24)

	// Transfer 5% → should PASS
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferCh.ChannelID)
	s.Require().NoError(err, "transfer (5%%) should succeed")

	// Transfer 1% → should FAIL
	amount2 := s.calculateQuota(sdkmath.NewInt(totalSupply), 1)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount2, denom, transferCh.ChannelID)
	s.Require().Error(err, "transfer should fail (exceeds quota)")

	// Submit governance proposal to remove rate limit
	s.removeRateLimit(ctx, denom, transferCh.ChannelID)

	// Transfer 50% → should PASS (no limit)
	amount3 := s.calculateQuota(sdkmath.NewInt(totalSupply), 50)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet,
		s.ChainBWallet.FormattedAddress(), amount3, denom, transferCh.ChannelID)
	s.Require().NoError(err, "transfer (50%%) should succeed after rate limit removal")
}

// Test 6: TestRateLimitWithMintTo tests that tokens minted via mint-to are subject to the same rate limits
func (s *TokenFactoryRateLimitSuite) TestRateLimitWithMintTo() {
	ctx := s.GetContext()

	// Create tokenfactory denom on chain A
	denom, err := s.CreateDenom(s.DelegatorWallet, "minttoratelimit")
	s.Require().NoError(err, "failed to create denom 'minttoratelimit'")

	// Use mint-to to mint tokens directly to DelegatorWallet2 (not to the admin)
	totalSupply := int64(1_000_000)
	err = s.MintTo(s.DelegatorWallet, denom, totalSupply, s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err, "mint-to should succeed")

	// Verify DelegatorWallet2 has the tokens
	balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(totalSupply), balance,
		"DelegatorWallet2 should have all minted tokens via mint-to")

	// Verify admin (DelegatorWallet) has no tokens
	adminBalance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().True(adminBalance.IsZero(), "admin should have zero tokens when using mint-to")

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Set rate limit via governance: 10% send quota over 24 hours
	s.addRateLimit(ctx, denom, transferCh.ChannelID, 10, 10, 24)

	// Calculate expected IBC denom on chain B
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// DelegatorWallet2 (who received tokens via mint-to) transfers 5% → should PASS
	amount1 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet2,
		s.ChainBWallet.FormattedAddress(), amount1, denom, transferCh.ChannelID)
	s.Require().NoError(err, "first transfer (5%%) should succeed")

	// Wait for transfer to complete
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(ctx, s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(amount1), "chain B should receive first transfer")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Transfer another 5% → should PASS (10% total)
	amount2 := s.calculateQuota(sdkmath.NewInt(totalSupply), 5)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet2,
		s.ChainBWallet.FormattedAddress(), amount2, denom, transferCh.ChannelID)
	s.Require().NoError(err, "second transfer (5%%) should succeed, total 10%%")

	// Wait for second transfer
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(ctx, s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(amount1.Add(amount2)), "chain B should receive both transfers")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Get balance before failed transfer attempt
	balanceBefore, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer 1% more → should FAIL (would exceed 10% quota)
	// This verifies that mint-to'd tokens are subject to the same rate limits
	amount3 := s.calculateQuota(sdkmath.NewInt(totalSupply), 1)
	err = s.attemptTransfer(ctx, s.Chain, s.DelegatorWallet2,
		s.ChainBWallet.FormattedAddress(), amount3, denom, transferCh.ChannelID)
	s.Require().Error(err, "third transfer should fail (would exceed 10%% quota)")
	s.Require().Contains(err.Error(), "quota", "error should mention quota/rate limit")

	// Verify balance unchanged after failed transfer
	balanceAfter, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore, balanceAfter, "balance should be unchanged after failed transfer")
}

func TestTokenFactoryRateLimit(t *testing.T) {
	s := &TokenFactoryRateLimitSuite{
		TokenFactoryBaseSuite: &TokenFactoryBaseSuite{
			Suite: &delegator.Suite{
				Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
					UpgradeOnSetup: true,
					CreateRelayer:  true,
				}),
			},
		},
	}
	suite.Run(t, s)
}
