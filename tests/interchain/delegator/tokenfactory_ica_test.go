package delegator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"

	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
)

type ICATokenFactorySuite struct {
	*TokenFactoryBaseSuite
	Host          *chainsuite.Chain
	HostWallet    ibc.Wallet
	icaAddress    string
	srcChannel    *ibc.ChannelOutput
	ibcStakeDenom string
}

func TestICATokenFactory(t *testing.T) {
	s := &ICATokenFactorySuite{
		TokenFactoryBaseSuite: &TokenFactoryBaseSuite{
			Suite: &delegator.Suite{
				Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
					UpgradeOnSetup: true,
					CreateRelayer:  true, // Required for ICA tests
				}),
			},
		},
	}
	suite.Run(t, s)
}

func (s *ICATokenFactorySuite) SetupSuite() {
	s.Suite.SetupSuite()

	ctx := s.GetContext()

	// Add a second chain for ICA testing (host chain)
	// Start at old version, then upgrade to get tokenfactory with proper params
	hostChain, err := s.Chain.AddLinkedChain(ctx, s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.Host = hostChain

	// Upgrade host chain to v26.0.0 so it has tokenfactory with proper params
	err = s.Host.Upgrade(ctx, s.Env.UpgradeName, s.Env.NewGaiaImageVersion)
	s.Require().NoError(err)

	// Increase relayer max_gas to handle high-gas tokenfactory operations
	err = s.Relayer.SetMaxGas(ctx, 3000000)
	s.Require().NoError(err)

	// Create and fund wallet on host chain
	hostWallet, err := hostChain.BuildWallet(ctx, "host-delegator", "")
	s.Require().NoError(err)
	s.HostWallet = hostWallet

	err = hostChain.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: hostWallet.FormattedAddress(),
		Denom:   chainsuite.Uatom,
		Amount:  sdkmath.NewInt(11_000_000_000),
	})
	s.Require().NoError(err)

	// Get transfer channel for IBC operations
	srcChannel, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.Host)
	s.Require().NoError(err)
	s.srcChannel = srcChannel

	// Setup ICA account on host chain
	srcAddress := s.DelegatorWallet.FormattedAddress()
	icaAddress, err := s.Chain.SetupICAAccount(ctx, s.Host, s.Relayer, srcAddress, 0, 10_000_000_000)
	s.Require().NoError(err)
	s.Require().NotEmpty(icaAddress)
	s.icaAddress = icaAddress

	chainsuite.GetLogger(ctx).Sugar().Infof("ICA address on host chain: %s", icaAddress)

	// Send IBC transfer to fund the ICA account
	_, err = s.Chain.SendIBCTransfer(ctx, srcChannel.ChannelID, srcAddress, ibc.WalletAmount{
		Address: icaAddress,
		Amount:  sdkmath.NewInt(10_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}, ibc.TransferOptions{})
	s.Require().NoError(err)

	// Calculate IBC denom for uatom on host chain
	s.ibcStakeDenom = transfertypes.NewDenom(chainsuite.Uatom, transfertypes.NewHop(srcChannel.PortID, srcChannel.ChannelID)).IBCDenom()

	// Wait for IBC denom to appear on host
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceResult, err := s.Host.QueryJSON(ctx, "balance.amount", "bank", "balance", icaAddress, s.ibcStakeDenom)
		assert.NoError(c, err)
		balance, ok := sdkmath.NewIntFromString(balanceResult.String())
		assert.True(c, ok)
		assert.True(c, balance.GT(sdkmath.ZeroInt()), "ICA account should have IBC denom balance")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

// sendICATx sends an ICA transaction from controller to host chain
func (s *ICATokenFactorySuite) sendICATx(ctx context.Context, srcAddress, srcConnection, txJSON string) error {
	// Get connection ID from channel
	connectionHops := s.srcChannel.ConnectionHops
	s.Require().NotEmpty(connectionHops, "no connection hops found")
	srcConnection = connectionHops[0]

	// Generate packet data from transaction JSON
	packetDataJSON, _, err := s.Chain.GetNode().ExecBin(ctx,
		"tx", "ica", "host", "generate-packet-data", txJSON, "--encoding", "proto3")
	if err != nil {
		return fmt.Errorf("failed to generate packet data: %w", err)
	}

	// Write packet data to file in node's home directory
	packetDataFile := "ica_packet_data.json"
	err = s.Chain.GetNode().WriteFile(ctx, packetDataJSON, packetDataFile)
	if err != nil {
		return fmt.Errorf("failed to write packet data file: %w", err)
	}
	packetDataFile = s.Chain.GetNode().HomeDir() + "/" + packetDataFile

	// Send ICA transaction
	_, err = s.Chain.GetNode().ExecTx(ctx,
		s.DelegatorWallet.KeyName(),
		"interchain-accounts", "controller", "send-tx",
		srcConnection,
		packetDataFile,
	)
	if err != nil {
		return fmt.Errorf("failed to send ICA tx: %w", err)
	}

	return nil
}

// TestICACreateDenom tests creating a tokenfactory denom via ICA
func (s *ICATokenFactorySuite) TestICACreateDenom() {
	ctx := s.GetContext()
	subdenom := "icatoken"
	srcConnection := s.srcChannel.ConnectionHops[0]

	// Build MsgCreateDenom JSON
	msgJSON := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom",
		"sender": "%s",
		"subdenom": "%s"
	}`, s.icaAddress, subdenom)

	chainsuite.GetLogger(ctx).Sugar().Infof("Sending ICA create denom tx: %s", msgJSON)

	// Send via ICA
	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, msgJSON))

	// Clear relayer packets to ensure delivery
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify denom exists with correct admin
	expectedDenom := fmt.Sprintf("factory/%s/%s", s.icaAddress, subdenom)

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		adminResult, err := s.Host.QueryJSON(ctx,
			"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", expectedDenom)
		assert.NoError(c, err)
		assert.Equal(c, s.icaAddress, adminResult.String(),
			"ICA address should be the admin of created denom")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	chainsuite.GetLogger(ctx).Sugar().Infof("Successfully created denom %s via ICA", expectedDenom)
}

// TestICAMintTokens tests minting tokens via ICA
func (s *ICATokenFactorySuite) TestICAMintTokens() {
	ctx := s.GetContext()
	subdenom := "minttest"
	mintAmount := int64(1000000)
	srcConnection := s.srcChannel.ConnectionHops[0]

	// First create the denom
	createMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom",
		"sender": "%s",
		"subdenom": "%s"
	}`, s.icaAddress, subdenom)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, createMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)
	time.Sleep(2 * chainsuite.CommitTimeout)

	denom := fmt.Sprintf("factory/%s/%s", s.icaAddress, subdenom)

	// Now mint tokens
	mintMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgMint",
		"sender": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, denom, mintAmount)

	chainsuite.GetLogger(ctx).Sugar().Infof("Sending ICA mint tx")

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, mintMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify tokens were minted to ICA account
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceResult, err := s.Host.QueryJSON(ctx, "balance.amount", "bank", "balance", s.icaAddress, denom)
		assert.NoError(c, err)
		balance, ok := sdkmath.NewIntFromString(balanceResult.String())
		assert.True(c, ok)
		assert.Equal(c, sdkmath.NewInt(mintAmount), balance,
			"ICA account should have minted tokens")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	chainsuite.GetLogger(ctx).Sugar().Infof("Successfully minted %d %s via ICA", mintAmount, denom)
}

// TestICABurnTokens tests burning tokens via ICA
func (s *ICATokenFactorySuite) TestICABurnTokens() {
	ctx := s.GetContext()
	subdenom := "burntest"
	mintAmount := int64(1000000)
	burnAmount := int64(500000)
	srcConnection := s.srcChannel.ConnectionHops[0]

	// Create denom
	createMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom",
		"sender": "%s",
		"subdenom": "%s"
	}`, s.icaAddress, subdenom)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, createMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	denom := fmt.Sprintf("factory/%s/%s", s.icaAddress, subdenom)

	// Verify denom was created
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		adminResult, err := s.Host.QueryJSON(ctx,
			"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
		assert.NoError(c, err)
		assert.Equal(c, s.icaAddress, adminResult.String(),
			"denom should be created with ICA as admin")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Mint tokens
	mintMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgMint",
		"sender": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, denom, mintAmount)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, mintMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify tokens were minted before burning
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceResult, err := s.Host.QueryJSON(ctx, "balance.amount", "bank", "balance", s.icaAddress, denom)
		assert.NoError(c, err)
		balance, ok := sdkmath.NewIntFromString(balanceResult.String())
		assert.True(c, ok)
		assert.Equal(c, sdkmath.NewInt(mintAmount), balance,
			"tokens should be minted before burning")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Burn tokens
	burnMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgBurn",
		"sender": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, denom, burnAmount)

	chainsuite.GetLogger(ctx).Sugar().Infof("Sending ICA burn tx")

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, burnMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify tokens were burned
	expectedBalance := mintAmount - burnAmount

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balanceResult, err := s.Host.QueryJSON(ctx, "balance.amount", "bank", "balance", s.icaAddress, denom)
		assert.NoError(c, err)
		balance, ok := sdkmath.NewIntFromString(balanceResult.String())
		assert.True(c, ok)
		assert.Equal(c, sdkmath.NewInt(expectedBalance), balance,
			"balance should reflect burned tokens")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	chainsuite.GetLogger(ctx).Sugar().Infof("Successfully burned %d %s via ICA", burnAmount, denom)
}

// TestICAChangeAdmin tests changing denom admin via ICA
func (s *ICATokenFactorySuite) TestICAChangeAdmin() {
	ctx := s.GetContext()
	subdenom := "admintest"
	srcConnection := s.srcChannel.ConnectionHops[0]

	// Create denom with ICA as admin
	createMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom",
		"sender": "%s",
		"subdenom": "%s"
	}`, s.icaAddress, subdenom)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, createMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)
	time.Sleep(2 * chainsuite.CommitTimeout)

	denom := fmt.Sprintf("factory/%s/%s", s.icaAddress, subdenom)

	// Change admin to host wallet
	newAdmin := s.HostWallet.FormattedAddress()
	changeAdminMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgChangeAdmin",
		"sender": "%s",
		"denom": "%s",
		"new_admin": "%s"
	}`, s.icaAddress, denom, newAdmin)

	chainsuite.GetLogger(ctx).Sugar().Infof("Sending ICA change admin tx to %s", newAdmin)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, changeAdminMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify admin changed
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		adminResult, err := s.Host.QueryJSON(ctx,
			"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
		assert.NoError(c, err)
		assert.Equal(c, newAdmin, adminResult.String(),
			"admin should be changed to new address")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	chainsuite.GetLogger(ctx).Sugar().Infof("Successfully changed admin to %s via ICA", newAdmin)
}

// TestICASetMetadata tests setting denom metadata via ICA
func (s *ICATokenFactorySuite) TestICASetMetadata() {
	ctx := s.GetContext()
	subdenom := "metatest"
	srcConnection := s.srcChannel.ConnectionHops[0]

	// Create denom
	createMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom",
		"sender": "%s",
		"subdenom": "%s"
	}`, s.icaAddress, subdenom)

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, createMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)
	time.Sleep(2 * chainsuite.CommitTimeout)

	denom := fmt.Sprintf("factory/%s/%s", s.icaAddress, subdenom)

	// Set metadata (following standard tokenfactory pattern where name = denom)
	metadata := map[string]interface{}{
		"description": "ICA Test Token",
		"denom_units": []map[string]interface{}{
			{
				"denom":    denom,
				"exponent": 0,
				"aliases":  []string{"ICAT"},
			},
			{
				"denom":    "ICAT",
				"exponent": 6,
				"aliases":  []string{denom},
			},
		},
		"base":    denom,
		"display": "ICAT",
		"name":    denom, // Standard: name should be the full denom
		"symbol":  "ICAT",
	}

	metadataJSON, err := json.Marshal(metadata)
	s.Require().NoError(err)

	setMetadataMsg := fmt.Sprintf(`{
		"@type": "/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata",
		"sender": "%s",
		"metadata": %s
	}`, s.icaAddress, string(metadataJSON))

	chainsuite.GetLogger(ctx).Sugar().Infof("Sending ICA set metadata tx")

	s.Require().NoError(s.sendICATx(ctx, s.DelegatorWallet.FormattedAddress(), srcConnection, setMetadataMsg))
	s.Relayer.ClearTransferChannel(ctx, s.Chain, s.Host)

	// Verify metadata was set
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		metadata, err := s.Host.QueryJSON(ctx,
			"metadata", "bank", "denom-metadata", denom)
		assert.NoError(c, err)
		assert.Equal(c, denom, metadata.Get("name").String(),
			"denom name should equal full denom")
		assert.Equal(c, "ICAT", metadata.Get("symbol").String(),
			"denom symbol should be set")
		assert.Equal(c, "ICAT", metadata.Get("display").String(),
			"denom display should be set")
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	chainsuite.GetLogger(ctx).Sugar().Infof("Successfully set metadata for %s via ICA", denom)
}
