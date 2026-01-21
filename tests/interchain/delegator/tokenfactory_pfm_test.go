package delegator_test

import (
	"encoding/json"
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

type TokenFactoryPFMSuite struct {
	*TokenFactoryBaseSuite
	Chains           []*chainsuite.Chain
	ADelegatorWallet ibc.Wallet
	CDelegatorWallet ibc.Wallet
	DDelegatorWallet ibc.Wallet
}

func (s *TokenFactoryPFMSuite) SetupSuite() {
	s.Suite.SetupSuite()

	// Create 4-chain topology: A -> B -> C -> D
	chainB, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	chainC, err := chainB.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	chainD, err := chainC.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)

	s.Chains = []*chainsuite.Chain{s.Chain, chainB, chainC, chainD}
	s.ADelegatorWallet = s.DelegatorWallet

	// Create and fund wallet on chain C
	walletC, err := chainC.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.CDelegatorWallet = walletC
	err = chainC.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.CDelegatorWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   chainC.Config().Denom,
	})
	s.Require().NoError(err)

	// Create and fund wallet on chain D
	walletD, err := chainD.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.DDelegatorWallet = walletD
	err = chainD.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DDelegatorWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   chainD.Config().Denom,
	})
	s.Require().NoError(err)
}

// TestTokenFactoryPFMForward tests forwarding a tokenfactory token from chain A to chain D via PFM (3 hops)
func (s *TokenFactoryPFMSuite) TestTokenFactoryPFMForward() {
	// Create tokenfactory denom on chain A
	denom, err := s.CreateDenom(s.ADelegatorWallet, "pfmtoken")
	s.Require().NoError(err, "failed to create denom 'pfmtoken'")

	// Mint tokens on chain A
	mintAmount := int64(10000000)
	err = s.Mint(s.ADelegatorWallet, denom, mintAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Build forward channels array and calculate expected IBC denom on chain D
	var forwardChannels []*ibc.ChannelOutput
	targetDenom := denom
	for i := 0; i < len(s.Chains)-1; i++ {
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i+1])
		s.Require().NoError(err)
		forwardChannels = append(forwardChannels, transferCh)
		targetDenom = transfertypes.GetPrefixedDenom(transferCh.PortID, transferCh.Counterparty.ChannelID, targetDenom)
	}
	expectedDenomD := transfertypes.ParseDenomTrace(targetDenom).IBCDenom()

	// Get initial balance on chain D
	dStartBalance, err := s.Chains[3].GetBalance(s.GetContext(), s.DDelegatorWallet.FormattedAddress(), expectedDenomD)
	s.Require().NoError(err)

	// Construct PFM memo for A -> B -> C -> D
	timeout := "10m"
	memo := map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": "pfm",
			"port":     "transfer",
			"channel":  forwardChannels[1].ChannelID,
			"timeout":  timeout,
			"next": map[string]interface{}{
				"forward": map[string]interface{}{
					"receiver": s.DDelegatorWallet.FormattedAddress(),
					"port":     "transfer",
					"channel":  forwardChannels[2].ChannelID,
					"timeout":  timeout,
				},
			},
		},
	}
	memoBytes, err := json.Marshal(memo)
	s.Require().NoError(err)

	// Execute IBC transfer with PFM memo
	transferAmount := int64(5000000)
	_, err = s.Chains[0].GetNode().ExecTx(s.GetContext(), s.ADelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer", forwardChannels[0].ChannelID, "pfm",
		fmt.Sprintf("%d%s", transferAmount, denom),
		"--memo", string(memoBytes))
	s.Require().NoError(err)

	// Verify tokens arrive on chain D
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		dEndBalance, err := s.Chains[3].GetBalance(s.GetContext(), s.DDelegatorWallet.FormattedAddress(), expectedDenomD)
		assert.NoError(c, err)
		assert.Truef(c, dEndBalance.Sub(dStartBalance).Equal(sdkmath.NewInt(transferAmount)),
			"expected balance increase of %d, got %d", transferAmount, dEndBalance.Sub(dStartBalance))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "tokenfactory token did not arrive on chain D via PFM")

	// Verify balance decreased on chain A
	aEndBalance, err := s.Chains[0].GetBalance(s.GetContext(), s.ADelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount), aEndBalance)
}

// TestTokenFactoryPFMRoundTrip tests forwarding a tokenfactory token A -> D and back D -> A
func (s *TokenFactoryPFMSuite) TestTokenFactoryPFMRoundTrip() {
	// Create tokenfactory denom on chain A
	denom, err := s.CreateDenom(s.ADelegatorWallet, "roundtriptoken")
	s.Require().NoError(err, "failed to create denom 'roundtriptoken'")

	// Mint tokens on chain A
	mintAmount := int64(10000000)
	err = s.Mint(s.ADelegatorWallet, denom, mintAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Build forward channels and calculate expected IBC denom on chain D
	var forwardChannels []*ibc.ChannelOutput
	targetDenomAD := denom
	for i := 0; i < len(s.Chains)-1; i++ {
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i+1])
		s.Require().NoError(err)
		forwardChannels = append(forwardChannels, transferCh)
		targetDenomAD = transfertypes.GetPrefixedDenom(transferCh.PortID, transferCh.Counterparty.ChannelID, targetDenomAD)
	}
	expectedDenomD := transfertypes.ParseDenomTrace(targetDenomAD).IBCDenom()

	// Forward transfer A -> D
	timeout := "10m"
	forwardMemo := map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": "pfm",
			"port":     "transfer",
			"channel":  forwardChannels[1].ChannelID,
			"timeout":  timeout,
			"next": map[string]interface{}{
				"forward": map[string]interface{}{
					"receiver": s.DDelegatorWallet.FormattedAddress(),
					"port":     "transfer",
					"channel":  forwardChannels[2].ChannelID,
					"timeout":  timeout,
				},
			},
		},
	}
	forwardMemoBytes, err := json.Marshal(forwardMemo)
	s.Require().NoError(err)

	transferAmount := int64(5000000)
	_, err = s.Chains[0].GetNode().ExecTx(s.GetContext(), s.ADelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer", forwardChannels[0].ChannelID, "pfm",
		fmt.Sprintf("%d%s", transferAmount, denom),
		"--memo", string(forwardMemoBytes))
	s.Require().NoError(err)

	// Wait for tokens to arrive on chain D
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		dBalance, err := s.Chains[3].GetBalance(s.GetContext(), s.DDelegatorWallet.FormattedAddress(), expectedDenomD)
		assert.NoError(c, err)
		assert.Truef(c, dBalance.Equal(sdkmath.NewInt(transferAmount)),
			"expected balance %d on chain D, got %d", transferAmount, dBalance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "forward transfer did not complete")

	// Build backward channels D -> C -> B -> A
	backwardChannels := make([]*ibc.ChannelOutput, len(forwardChannels))
	for i := len(s.Chains) - 1; i > 0; i-- {
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i-1])
		s.Require().NoError(err)
		backwardChannels[i-1] = transferCh
	}

	// Get balance on chain A before return transfer
	aStartBalance, err := s.Chains[0].GetBalance(s.GetContext(), s.ADelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Return transfer D -> A
	returnAmount := int64(3000000)
	backwardMemo := map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": "pfm",
			"port":     "transfer",
			"channel":  backwardChannels[1].ChannelID,
			"timeout":  timeout,
			"next": map[string]interface{}{
				"forward": map[string]interface{}{
					"receiver": s.ADelegatorWallet.FormattedAddress(),
					"port":     "transfer",
					"channel":  backwardChannels[0].ChannelID,
					"timeout":  timeout,
				},
			},
		},
	}
	backwardMemoBytes, err := json.Marshal(backwardMemo)
	s.Require().NoError(err)

	_, err = s.Chains[3].GetNode().ExecTx(s.GetContext(), s.DDelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer", backwardChannels[2].ChannelID, "pfm",
		fmt.Sprintf("%d%s", returnAmount, expectedDenomD),
		"--memo", string(backwardMemoBytes))
	s.Require().NoError(err)

	// Verify tokens arrive back on chain A (should unwrap to original denom)
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		aEndBalance, err := s.Chains[0].GetBalance(s.GetContext(), s.ADelegatorWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Truef(c, aEndBalance.Sub(aStartBalance).Equal(sdkmath.NewInt(returnAmount)),
			"expected balance increase of %d, got %d", returnAmount, aEndBalance.Sub(aStartBalance))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "return transfer did not complete - tokens not unwrapped on chain A")
}

// TestTokenFactoryPFMPartialHop tests forwarding a tokenfactory token from chain A to chain C (2 hops)
func (s *TokenFactoryPFMSuite) TestTokenFactoryPFMPartialHop() {
	// Create tokenfactory denom on chain A
	denom, err := s.CreateDenom(s.ADelegatorWallet, "partialtoken")
	s.Require().NoError(err, "failed to create denom 'partialtoken'")

	// Mint tokens on chain A
	mintAmount := int64(10000000)
	err = s.Mint(s.ADelegatorWallet, denom, mintAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Build forward channels for A -> B -> C only (2 hops)
	var forwardChannels []*ibc.ChannelOutput
	targetDenom := denom
	for i := 0; i < 2; i++ { // Only 2 hops
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i+1])
		s.Require().NoError(err)
		forwardChannels = append(forwardChannels, transferCh)
		targetDenom = transfertypes.GetPrefixedDenom(transferCh.PortID, transferCh.Counterparty.ChannelID, targetDenom)
	}
	expectedDenomC := transfertypes.ParseDenomTrace(targetDenom).IBCDenom()

	// Get initial balance on chain C
	cStartBalance, err := s.Chains[2].GetBalance(s.GetContext(), s.CDelegatorWallet.FormattedAddress(), expectedDenomC)
	s.Require().NoError(err)

	// Construct PFM memo for A -> B -> C (single forward, no nesting)
	timeout := "10m"
	memo := map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": s.CDelegatorWallet.FormattedAddress(),
			"port":     "transfer",
			"channel":  forwardChannels[1].ChannelID,
			"timeout":  timeout,
		},
	}
	memoBytes, err := json.Marshal(memo)
	s.Require().NoError(err)

	// Execute IBC transfer with PFM memo
	transferAmount := int64(5000000)
	_, err = s.Chains[0].GetNode().ExecTx(s.GetContext(), s.ADelegatorWallet.KeyName(),
		"ibc-transfer", "transfer", "transfer", forwardChannels[0].ChannelID, "pfm",
		fmt.Sprintf("%d%s", transferAmount, denom),
		"--memo", string(memoBytes))
	s.Require().NoError(err)

	// Verify tokens arrive on chain C
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		cEndBalance, err := s.Chains[2].GetBalance(s.GetContext(), s.CDelegatorWallet.FormattedAddress(), expectedDenomC)
		assert.NoError(c, err)
		assert.Truef(c, cEndBalance.Sub(cStartBalance).Equal(sdkmath.NewInt(transferAmount)),
			"expected balance increase of %d on chain C, got %d", transferAmount, cEndBalance.Sub(cStartBalance))
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "tokenfactory token did not arrive on chain C via PFM")

	// Verify balance decreased on chain A
	aEndBalance, err := s.Chains[0].GetBalance(s.GetContext(), s.ADelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-transferAmount), aEndBalance)
}

func TestTokenFactoryPFM(t *testing.T) {
	s := &TokenFactoryPFMSuite{
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
