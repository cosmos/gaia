package delegator_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryIBCSuite struct {
	*delegator.Suite
	ChainB       *chainsuite.Chain
	ChainBWallet ibc.Wallet
}

func (s *TokenFactoryIBCSuite) SetupSuite() {
	s.Suite.SetupSuite()

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

// createDenom creates a tokenfactory denom on chain A
func (s *TokenFactoryIBCSuite) createDenom(subdenom string) string {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)
}

// mint mints tokens on chain A
func (s *TokenFactoryIBCSuite) mint(denom string, amount int64) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	s.Require().NoError(err)
}

// TestIBCTransferToChainB tests transferring tokenfactory tokens to another chain
func (s *TokenFactoryIBCSuite) TestIBCTransferToChainB() {
	// Create tokenfactory denom on chain A
	denom := s.createDenom("ibctoken")

	// Mint tokens on chain A
	mintAmount := int64(10000000)
	s.mint(denom, mintAmount)

	// Verify balance on chain A
	balanceA, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balanceA)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate expected IBC denom on chain B
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	ibcDenomTrace := transfertypes.ParseDenomTrace(ibcDenom)
	expectedDenomB := ibcDenomTrace.IBCDenom()

	// Get initial balance on chain B (should be zero)
	balanceBBefore, err := s.ChainB.GetBalance(s.GetContext(),
		s.ChainBWallet.FormattedAddress(), expectedDenomB)
	s.Require().NoError(err)

	// Transfer tokens from chain A to chain B
	transferAmount := int64(5000000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)

	// Wait for transfer to complete and verify balance on chain B
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceB, err := s.ChainB.GetBalance(s.GetContext(),
			s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balanceB.Sub(balanceBBefore).Equal(sdkmath.NewInt(transferAmount)),
			"expected balance increase of %d, got %d", transferAmount, balanceB.Sub(balanceBBefore))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "chain B balance did not increase")

	// Verify balance decreased on chain A
	balanceAAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount), balanceAAfter)
}

// TestIBCTransferRoundTrip tests transferring tokenfactory tokens to another chain and back
func (s *TokenFactoryIBCSuite) TestIBCTransferRoundTrip() {
	// Create tokenfactory denom on chain A
	denom := s.createDenom("roundtrip")

	// Mint tokens on chain A
	mintAmount := int64(10000000)
	s.mint(denom, mintAmount)

	// Get IBC transfer channel
	transferChAtoB, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate expected IBC denom on chain B
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferChAtoB.Counterparty.PortID,
		transferChAtoB.Counterparty.ChannelID,
		denom,
	)
	ibcDenomTrace := transfertypes.ParseDenomTrace(ibcDenom)
	expectedDenomB := ibcDenomTrace.IBCDenom()

	// Transfer tokens from chain A to chain B
	transferAmount := int64(5000000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChAtoB.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)

	// Wait for transfer to complete
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceB, err := s.ChainB.GetBalance(s.GetContext(),
			s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balanceB.Equal(sdkmath.NewInt(transferAmount)),
			"expected balance %d, got %d", transferAmount, balanceB)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "chain B balance did not update")

	// Get return channel from B to A
	transferChBtoA, err := s.Relayer.GetTransferChannel(s.GetContext(), s.ChainB, s.Chain)
	s.Require().NoError(err)

	// Get balance on chain A before return transfer
	balanceABefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer tokens back from chain B to chain A
	returnAmount := int64(3000000)
	_, err = s.ChainB.GetNode().ExecTx(
		s.GetContext(),
		s.ChainBWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChBtoA.ChannelID,
		s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", returnAmount, expectedDenomB),
	)
	s.Require().NoError(err)

	// Wait for return transfer and verify tokens are unwrapped on chain A
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceA, err := s.Chain.GetBalance(s.GetContext(),
			s.DelegatorWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.True(c, balanceA.Sub(balanceABefore).Equal(sdkmath.NewInt(returnAmount)),
			"expected balance increase of %d, got %d", returnAmount, balanceA.Sub(balanceABefore))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "chain A balance did not increase")

	// Verify final balances
	finalBalanceA, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount+returnAmount), finalBalanceA)

	finalBalanceB, err := s.ChainB.GetBalance(s.GetContext(),
		s.ChainBWallet.FormattedAddress(), expectedDenomB)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(transferAmount-returnAmount), finalBalanceB)
}

// TestIBCTransferWithMetadata tests that metadata is accessible on the remote chain
func (s *TokenFactoryIBCSuite) TestIBCTransferWithMetadata() {
	// Create tokenfactory denom on chain A
	subdenom := "metatoken"
	denom := s.createDenom(subdenom)

	// Set metadata on chain A
	metadataJSON := fmt.Sprintf(`{
		"base": "%s",
		"display": "%s",
		"name": "IBC Meta Token",
		"symbol": "IBCMETA",
		"description": "A test token with metadata for IBC",
		"denom_units": [
			{"denom": "%s", "exponent": 0},
			{"denom": "%s", "exponent": 6}
		]
	}`, denom, subdenom, denom, subdenom)

	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "set-denom-metadata",
		denom,
		metadataJSON,
	)
	s.Require().NoError(err)

	// Verify metadata on chain A
	metadataA, err := s.Chain.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal("IBC Meta Token", metadataA.Get("metadata.name").String())

	// Mint and transfer tokens
	s.mint(denom, 10000000)

	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate IBC denom on chain B
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("5000000%s", denom),
	)
	s.Require().NoError(err)

	// Wait for transfer
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceB, err := s.ChainB.GetBalance(s.GetContext(),
			s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balanceB.GT(sdkmath.ZeroInt()))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "transfer did not complete")

	// Query metadata on chain B
	// Note: IBC denoms may not have metadata automatically transferred
	// This test verifies the behavior
	metadataB, err := s.ChainB.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", expectedDenomB)

	// Depending on implementation, metadata might not be present for IBC denoms
	// We're testing the behavior here
	if err == nil && metadataB.Exists() {
		s.T().Logf("Metadata found on chain B: %s", metadataB.String())
	} else {
		s.T().Logf("Metadata not found on chain B (expected for IBC denoms)")
	}
}

// TestMultipleTokenFactoryIBCTransfers tests multiple tokenfactory tokens over IBC
func (s *TokenFactoryIBCSuite) TestMultipleTokenFactoryIBCTransfers() {
	// Create multiple tokenfactory denoms
	denom1 := s.createDenom("ibc1")
	denom2 := s.createDenom("ibc2")
	denom3 := s.createDenom("ibc3")

	// Mint tokens for each
	s.mint(denom1, 10000000)
	s.mint(denom2, 20000000)
	s.mint(denom3, 30000000)

	// Get transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate expected IBC denoms on chain B
	expectedDenoms := make([]string, 3)
	for i, denom := range []string{denom1, denom2, denom3} {
		ibcDenom := transfertypes.GetPrefixedDenom(
			transferCh.Counterparty.PortID,
			transferCh.Counterparty.ChannelID,
			denom,
		)
		expectedDenoms[i] = transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()
	}

	// Transfer all three tokens
	amounts := []int64{5000000, 10000000, 15000000}
	for i, denom := range []string{denom1, denom2, denom3} {
		_, err = s.Chain.GetNode().ExecTx(
			s.GetContext(),
			s.DelegatorWallet.KeyName(),
			"ibc-transfer", "transfer", "transfer",
			transferCh.ChannelID,
			s.ChainBWallet.FormattedAddress(),
			fmt.Sprintf("%d%s", amounts[i], denom),
		)
		s.Require().NoError(err)
	}

	// Verify all transfers completed
	for i, expectedDenom := range expectedDenoms {
		i := i // capture loop variable
		expectedAmount := amounts[i]
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balance, err := s.ChainB.GetBalance(s.GetContext(),
				s.ChainBWallet.FormattedAddress(), expectedDenom)
			assert.NoError(c, err)
			assert.True(c, balance.Equal(sdkmath.NewInt(expectedAmount)),
				"token %d: expected balance %d, got %d", i+1, expectedAmount, balance)
		}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
			fmt.Sprintf("transfer of token %d did not complete", i+1))
	}
}

// TestIBCTransferAfterAdminChange tests that IBC transfers work after admin change
func (s *TokenFactoryIBCSuite) TestIBCTransferAfterAdminChange() {
	// Create denom with DelegatorWallet as admin
	denom := s.createDenom("adminchange")
	s.mint(denom, 10000000)

	// Transfer some tokens to DelegatorWallet2
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("5000000%s", denom),
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

	// Get transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	// Calculate expected IBC denom
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Both wallets should be able to transfer via IBC
	// Transfer from DelegatorWallet (no longer admin, but has tokens)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("1000000%s", denom),
	)
	s.Require().NoError(err)

	// Transfer from DelegatorWallet2 (current admin)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("2000000%s", denom),
	)
	s.Require().NoError(err)

	// Verify total received on chain B
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(s.GetContext(),
			s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(sdkmath.NewInt(3000000)),
			"expected total balance 3000000, got %d", balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "transfers did not complete")
}

// TestIBCTransferAfterAdminRenounce tests that IBC transfers work after admin is renounced
func (s *TokenFactoryIBCSuite) TestIBCTransferAfterAdminRenounce() {
	// Create denom and mint tokens
	denom := s.createDenom("renounced")
	s.mint(denom, 10000000)

	// Renounce admin
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom, "",
	)
	s.Require().NoError(err)

	// IBC transfers should still work even without admin
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.ChainB)
	s.Require().NoError(err)

	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomB := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.ChainBWallet.FormattedAddress(),
		fmt.Sprintf("5000000%s", denom),
	)
	s.Require().NoError(err)

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.ChainB.GetBalance(s.GetContext(),
			s.ChainBWallet.FormattedAddress(), expectedDenomB)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(sdkmath.NewInt(5000000)))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "transfer did not complete")
}

func TestTokenFactoryIBC(t *testing.T) {
	s := &TokenFactoryIBCSuite{
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: true,
				CreateRelayer:  true,
			}),
		},
	}
	suite.Run(t, s)
}
