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

// OsmosisIBCSuite tests IBC transfers between Gaia and Osmosis,
// including native tokens and tokenfactory tokens from both chains.
type OsmosisIBCSuite struct {
	*TokenFactoryBaseSuite
	Osmosis       *chainsuite.Chain
	OsmosisWallet ibc.Wallet
}

func (s *OsmosisIBCSuite) SetupSuite() {
	s.Suite.SetupSuite()

	// Add Osmosis chain for IBC testing
	osmosis, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.OsmosisChainSpec())
	s.Require().NoError(err)
	s.Osmosis = osmosis

	// Create wallet on Osmosis
	wallet, err := osmosis.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.OsmosisWallet = wallet

	// Fund wallet on Osmosis
	err = osmosis.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.OsmosisWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   osmosis.Config().Denom,
	})
	s.Require().NoError(err)
}

// TestNativeTransfer_GaiaToOsmosis tests transferring uatom from Gaia to Osmosis
func (s *OsmosisIBCSuite) TestNativeTransfer_GaiaToOsmosis() {
	// Get initial balance on Gaia
	gaiaBalanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Osmosis)
	s.Require().NoError(err)

	// Calculate expected IBC denom on Osmosis
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		chainsuite.Uatom,
	)
	expectedDenomOsmosis := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Get initial balance on Osmosis (should be zero)
	osmosisBalanceBefore, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), expectedDenomOsmosis)
	s.Require().NoError(err)

	// Transfer uatom from Gaia to Osmosis
	transferAmount := int64(5_000_000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.OsmosisWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, chainsuite.Uatom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Osmosis)

	// Wait for transfer to complete and verify balance on Osmosis
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceOsmosis, err := s.Osmosis.GetBalance(s.GetContext(),
			s.OsmosisWallet.FormattedAddress(), expectedDenomOsmosis)
		assert.NoError(c, err)
		assert.True(c, balanceOsmosis.Sub(osmosisBalanceBefore).Equal(sdkmath.NewInt(transferAmount)),
			"expected balance increase of %d, got %s", transferAmount, balanceOsmosis.Sub(osmosisBalanceBefore).String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Osmosis balance did not increase")

	// Verify balance decreased on Gaia
	gaiaBalanceAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	s.Require().True(gaiaBalanceBefore.Sub(gaiaBalanceAfter).GTE(sdkmath.NewInt(transferAmount)),
		"Gaia balance should have decreased by at least %d", transferAmount)
}

// TestNativeTransfer_OsmosisToGaia tests transferring uosmo from Osmosis to Gaia
func (s *OsmosisIBCSuite) TestNativeTransfer_OsmosisToGaia() {
	// Get IBC transfer channel from Osmosis to Gaia
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Osmosis, s.Chain)
	s.Require().NoError(err)

	// Calculate expected IBC denom on Gaia
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		chainsuite.OsmosisDenom,
	)
	expectedDenomGaia := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Get initial balance on Gaia (should be zero)
	gaiaBalanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), expectedDenomGaia)
	s.Require().NoError(err)

	// Transfer uosmo from Osmosis to Gaia
	transferAmount := int64(5_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, chainsuite.OsmosisDenom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Osmosis, s.Chain)

	// Wait for transfer to complete and verify balance on Gaia
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceGaia, err := s.Chain.GetBalance(s.GetContext(),
			s.DelegatorWallet.FormattedAddress(), expectedDenomGaia)
		assert.NoError(c, err)
		assert.True(c, balanceGaia.Sub(gaiaBalanceBefore).Equal(sdkmath.NewInt(transferAmount)),
			"expected balance increase of %d, got %s", transferAmount, balanceGaia.Sub(gaiaBalanceBefore).String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Gaia balance did not increase")
}

// TestGaiaTokenFactoryToOsmosis tests transferring a Gaia tokenfactory token to Osmosis
func (s *OsmosisIBCSuite) TestGaiaTokenFactoryToOsmosis() {
	// Create tokenfactory denom on Gaia
	denom, err := s.CreateDenom(s.DelegatorWallet, "gaia2osmo")
	s.Require().NoError(err, "failed to create denom on Gaia")

	// Mint tokens on Gaia
	mintAmount := int64(10_000_000)
	err = s.Mint(s.DelegatorWallet, denom, mintAmount)
	s.Require().NoError(err, "failed to mint tokens on Gaia")

	// Get IBC transfer channel
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Osmosis)
	s.Require().NoError(err)

	// Calculate expected IBC denom on Osmosis
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomOsmosis := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Transfer tokenfactory tokens from Gaia to Osmosis
	transferAmount := int64(5_000_000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.OsmosisWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Osmosis)

	// Wait for transfer and verify balance on Osmosis
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceOsmosis, err := s.Osmosis.GetBalance(s.GetContext(),
			s.OsmosisWallet.FormattedAddress(), expectedDenomOsmosis)
		assert.NoError(c, err)
		assert.True(c, balanceOsmosis.Equal(sdkmath.NewInt(transferAmount)),
			"expected balance %d, got %s", transferAmount, balanceOsmosis.String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Osmosis balance did not update")

	// Verify balance on Gaia decreased
	balanceGaia, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount), balanceGaia)
}

// TestOsmosisTokenFactoryToGaia tests transferring an Osmosis tokenfactory token to Gaia
func (s *OsmosisIBCSuite) TestOsmosisTokenFactoryToGaia() {
	// Create tokenfactory denom on Osmosis
	subdenom := "osmo2gaia"
	// Osmosis tokenfactory has denom_creation_gas_consume of 1M, need explicit high gas
	_, err := s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
		"--gas", "2000000",
	)
	s.Require().NoError(err, "failed to create denom on Osmosis")

	denom := fmt.Sprintf("factory/%s/%s", s.OsmosisWallet.FormattedAddress(), subdenom)

	// Mint tokens on Osmosis (Osmosis mint requires: amount, mint-to-address)
	mintAmount := int64(10_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", mintAmount, denom),
		s.OsmosisWallet.FormattedAddress(),
		"--gas", "2000000",
	)
	s.Require().NoError(err, "failed to mint tokens on Osmosis")

	// Verify balance on Osmosis
	balanceOsmosis, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balanceOsmosis)

	// Get IBC transfer channel from Osmosis to Gaia
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Osmosis, s.Chain)
	s.Require().NoError(err)

	// Calculate expected IBC denom on Gaia
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferCh.Counterparty.PortID,
		transferCh.Counterparty.ChannelID,
		denom,
	)
	expectedDenomGaia := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Transfer tokenfactory tokens from Osmosis to Gaia
	transferAmount := int64(5_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferCh.ChannelID,
		s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Osmosis, s.Chain)

	// Wait for transfer and verify balance on Gaia
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceGaia, err := s.Chain.GetBalance(s.GetContext(),
			s.DelegatorWallet.FormattedAddress(), expectedDenomGaia)
		assert.NoError(c, err)
		assert.True(c, balanceGaia.Equal(sdkmath.NewInt(transferAmount)),
			"expected balance %d, got %s", transferAmount, balanceGaia.String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Gaia balance did not update")

	// Verify balance on Osmosis decreased
	balanceOsmosisAfter, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount), balanceOsmosisAfter)
}

// TestTokenFactoryRoundTrip_GaiaOrigin tests round-trip transfer of Gaia tokenfactory token
func (s *OsmosisIBCSuite) TestTokenFactoryRoundTrip_GaiaOrigin() {
	// Create tokenfactory denom on Gaia
	denom, err := s.CreateDenom(s.DelegatorWallet, "roundtrip")
	s.Require().NoError(err, "failed to create denom on Gaia")

	// Mint tokens on Gaia
	mintAmount := int64(10_000_000)
	err = s.Mint(s.DelegatorWallet, denom, mintAmount)
	s.Require().NoError(err, "failed to mint tokens on Gaia")

	// Get IBC transfer channel Gaia -> Osmosis
	transferChGaiaToOsmo, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Osmosis)
	s.Require().NoError(err)

	// Calculate IBC denom on Osmosis
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferChGaiaToOsmo.Counterparty.PortID,
		transferChGaiaToOsmo.Counterparty.ChannelID,
		denom,
	)
	ibcDenomOnOsmosis := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Transfer Gaia -> Osmosis
	transferAmount := int64(5_000_000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChGaiaToOsmo.ChannelID,
		s.OsmosisWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Osmosis)

	// Wait for transfer to Osmosis
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Osmosis.GetBalance(s.GetContext(),
			s.OsmosisWallet.FormattedAddress(), ibcDenomOnOsmosis)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(sdkmath.NewInt(transferAmount)))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Get return channel Osmosis -> Gaia
	transferChOsmoToGaia, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Osmosis, s.Chain)
	s.Require().NoError(err)

	// Get Gaia balance before return
	gaiaBalanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer back Osmosis -> Gaia (should unwrap to original denom)
	returnAmount := int64(3_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChOsmoToGaia.ChannelID,
		s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", returnAmount, ibcDenomOnOsmosis),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Osmosis, s.Chain)

	// Wait for return transfer and verify tokens are unwrapped on Gaia
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		gaiaBalance, err := s.Chain.GetBalance(s.GetContext(),
			s.DelegatorWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.True(c, gaiaBalance.Sub(gaiaBalanceBefore).Equal(sdkmath.NewInt(returnAmount)),
			"expected balance increase of %d, got %s", returnAmount, gaiaBalance.Sub(gaiaBalanceBefore).String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Gaia balance did not increase")

	// Verify final balances
	finalGaiaBalance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount+returnAmount), finalGaiaBalance)

	finalOsmosisBalance, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), ibcDenomOnOsmosis)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(transferAmount-returnAmount), finalOsmosisBalance)
}

// TestTokenFactoryRoundTrip_OsmosisOrigin tests round-trip transfer of Osmosis tokenfactory token
func (s *OsmosisIBCSuite) TestTokenFactoryRoundTrip_OsmosisOrigin() {
	// Create tokenfactory denom on Osmosis
	subdenom := "osmoroundtrip"
	// Osmosis tokenfactory has denom_creation_gas_consume of 1M, need explicit high gas
	_, err := s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
		"--gas", "2000000",
	)
	s.Require().NoError(err, "failed to create denom on Osmosis")

	denom := fmt.Sprintf("factory/%s/%s", s.OsmosisWallet.FormattedAddress(), subdenom)

	// Mint tokens on Osmosis (Osmosis mint requires: amount, mint-to-address)
	mintAmount := int64(10_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", mintAmount, denom),
		s.OsmosisWallet.FormattedAddress(),
		"--gas", "2000000",
	)
	s.Require().NoError(err, "failed to mint tokens on Osmosis")

	// Get IBC transfer channel Osmosis -> Gaia
	transferChOsmoToGaia, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Osmosis, s.Chain)
	s.Require().NoError(err)

	// Calculate IBC denom on Gaia
	ibcDenom := transfertypes.GetPrefixedDenom(
		transferChOsmoToGaia.Counterparty.PortID,
		transferChOsmoToGaia.Counterparty.ChannelID,
		denom,
	)
	ibcDenomOnGaia := transfertypes.ParseDenomTrace(ibcDenom).IBCDenom()

	// Transfer Osmosis -> Gaia
	transferAmount := int64(5_000_000)
	_, err = s.Osmosis.GetNode().ExecTx(
		s.GetContext(),
		s.OsmosisWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChOsmoToGaia.ChannelID,
		s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", transferAmount, denom),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Osmosis, s.Chain)

	// Wait for transfer to Gaia
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(s.GetContext(),
			s.DelegatorWallet.FormattedAddress(), ibcDenomOnGaia)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(sdkmath.NewInt(transferAmount)))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Get return channel Gaia -> Osmosis
	transferChGaiaToOsmo, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Osmosis)
	s.Require().NoError(err)

	// Get Osmosis balance before return
	osmosisBalanceBefore, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Transfer back Gaia -> Osmosis (should unwrap to original denom)
	returnAmount := int64(3_000_000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer",
		transferChGaiaToOsmo.ChannelID,
		s.OsmosisWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", returnAmount, ibcDenomOnGaia),
	)
	s.Require().NoError(err)
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Osmosis)

	// Wait for return transfer and verify tokens are unwrapped on Osmosis
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		osmosisBalance, err := s.Osmosis.GetBalance(s.GetContext(),
			s.OsmosisWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.True(c, osmosisBalance.Sub(osmosisBalanceBefore).Equal(sdkmath.NewInt(returnAmount)),
			"expected balance increase of %d, got %s", returnAmount, osmosisBalance.Sub(osmosisBalanceBefore).String())
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "Osmosis balance did not increase")

	// Verify final balances
	finalOsmosisBalance, err := s.Osmosis.GetBalance(s.GetContext(),
		s.OsmosisWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount+returnAmount), finalOsmosisBalance)

	finalGaiaBalance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), ibcDenomOnGaia)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(transferAmount-returnAmount), finalGaiaBalance)
}

func TestOsmosisIBC(t *testing.T) {
	s := &OsmosisIBCSuite{
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
