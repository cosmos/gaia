package delegator_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type TokenFactorySuite struct {
	*delegator.Suite
}

// createDenom creates a new tokenfactory denom and returns the full denom string
func (s *TokenFactorySuite) createDenom(subdenom string) string {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)
}

// mint mints tokens for a given denom
func (s *TokenFactorySuite) mint(denom string, amount int64) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	s.Require().NoError(err)
}

// createAndMint creates a denom and mints tokens in one go
func (s *TokenFactorySuite) createAndMint(subdenom string, amount int64) string {
	denom := s.createDenom(subdenom)
	s.mint(denom, amount)
	return denom
}

// TestCreateDenomCLI tests creating a tokenfactory denom via CLI
func (s *TokenFactorySuite) TestCreateDenomCLI() {
	subdenom := "mytoken"

	// Execute create-denom command
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)

	// Construct expected denom
	denom := fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)

	// Query denom authority metadata to verify admin
	admin, err := s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet.FormattedAddress(), admin.String())
}

// TestMintBurnCLI tests minting and burning tokens via CLI
func (s *TokenFactorySuite) TestMintBurnCLI() {
	// Create a denom first
	denom := s.createDenom("testtoken")

	// Mint tokens
	mintAmount := int64(1000000)
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", mintAmount, denom),
	)
	s.Require().NoError(err)

	// Verify balance
	balance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)

	// Burn tokens
	burnAmount := int64(500000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "burn",
		fmt.Sprintf("%d%s", burnAmount, denom),
	)
	s.Require().NoError(err)

	// Verify reduced balance
	balance, err = s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-burnAmount), balance)
}

// TestSetDenomMetadataCLI tests setting denom metadata via CLI
func (s *TokenFactorySuite) TestSetDenomMetadataCLI() {
	// Create a denom
	subdenom := "metatoken"
	denom := s.createDenom(subdenom)

	// Set metadata via modify-metadata CLI
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "modify-metadata",
		denom,
		"META",                       // ticker-symbol
		"A test token for metadata",  // description
		"6",                          // exponent
	)
	s.Require().NoError(err)

	// Query metadata via bank module
	retrievedMetadata, err := s.Chain.QueryJSON(s.GetContext(),
		"metadata", "bank", "denom-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal("META", retrievedMetadata.Get("symbol").String())
	s.Require().Equal("A test token for metadata", retrievedMetadata.Get("description").String())
}

// TestAdminTransferBetweenWallets tests transferring admin privileges between wallets
func (s *TokenFactorySuite) TestAdminTransferBetweenWallets() {
	// Create denom with DelegatorWallet
	denom := s.createDenom("admintransfer")

	// Change admin to DelegatorWallet2
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom, s.DelegatorWallet2.FormattedAddress(),
	)
	s.Require().NoError(err)

	// Verify new admin
	admin, err := s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet2.FormattedAddress(), admin.String())

	// Verify DelegatorWallet cannot mint
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("100000%s", denom),
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")

	// Verify DelegatorWallet2 can mint
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("100000%s", denom),
	)
	s.Require().NoError(err)

	// Verify balance
	balance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(100000), balance)
}

// TestUnauthorizedMint tests that non-admin cannot mint tokens
func (s *TokenFactorySuite) TestUnauthorizedMint() {
	// Create denom with DelegatorWallet
	denom := s.createDenom("unauthorized")

	// Attempt to mint with DelegatorWallet2 (non-admin)
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("100000%s", denom),
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestUnauthorizedBurn tests that non-admin cannot burn tokens
func (s *TokenFactorySuite) TestUnauthorizedBurn() {
	// Create denom and mint to DelegatorWallet2
	denom := s.createAndMint("burnunauth", 1000000)

	// Transfer some tokens to DelegatorWallet2
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("500000%s", denom),
	)
	s.Require().NoError(err)

	// Attempt to burn with DelegatorWallet2 (non-admin)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet2.KeyName(),
		"tokenfactory", "burn",
		fmt.Sprintf("100000%s", denom),
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestBankSendWithTokenFactoryToken tests that tokenfactory tokens work with bank send
func (s *TokenFactorySuite) TestBankSendWithTokenFactoryToken() {
	// Create denom and mint tokens
	denom := s.createAndMint("banksend", 1000000)

	// Get balance before
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)

	// Send tokens using bank send
	sendAmount := int64(500000)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", sendAmount, denom),
	)
	s.Require().NoError(err)

	// Verify recipient balance
	balanceAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.Add(sdkmath.NewInt(sendAmount)), balanceAfter)

	// Verify sender balance
	senderBalance, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(1000000-sendAmount), senderBalance)
}

// TestMultipleDenoms tests creating and managing multiple denoms from one address
func (s *TokenFactorySuite) TestMultipleDenoms() {
	// Create multiple denoms
	denom1 := s.createDenom("token1")
	denom2 := s.createDenom("token2")
	denom3 := s.createDenom("token3")

	// Mint different amounts to each
	s.mint(denom1, 1000000)
	s.mint(denom2, 2000000)
	s.mint(denom3, 3000000)

	// Verify each has correct balance
	balance1, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom1)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(1000000), balance1)

	balance2, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom2)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(2000000), balance2)

	balance3, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom3)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(3000000), balance3)

	// Change admin on one denom
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "change-admin",
		denom1, s.DelegatorWallet2.FormattedAddress(),
	)
	s.Require().NoError(err)

	// Verify DelegatorWallet can still mint to denom2 and denom3
	s.mint(denom2, 1000000)
	s.mint(denom3, 1000000)

	// Verify new balances
	balance2, err = s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom2)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(3000000), balance2)

	balance3, err = s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), denom3)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(4000000), balance3)

	// Verify DelegatorWallet cannot mint to denom1 anymore
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("1000000%s", denom1),
	)
	s.Require().Error(err)
}

// TestCreationFee tests that the denom creation fee is charged correctly
func (s *TokenFactorySuite) TestCreationFee() {
	// Query creation fee from params
	params, err := s.Chain.QueryJSON(s.GetContext(),
		"params", "tokenfactory", "params")
	s.Require().NoError(err)

	// Get the creation fee amount
	creationFeeStr := params.Get("denom_creation_fee.0.amount").String()
	s.Require().NotEmpty(creationFeeStr)

	// Get balance before creation
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	// Create denom
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", "feetest",
	)
	s.Require().NoError(err)

	// Get balance after creation
	balanceAfter, err := s.Chain.GetBalance(s.GetContext(),
		s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	// Verify balance decreased (fee + gas)
	s.Require().True(balanceBefore.GT(balanceAfter))

	// The difference should be at least the creation fee
	creationFeeInt, ok := sdkmath.NewIntFromString(creationFeeStr)
	s.Require().True(ok)
	s.Require().True(balanceBefore.Sub(balanceAfter).GTE(creationFeeInt))
}

// TestCreationFeeFundsCommunityPool tests that denom creation fees are deposited to community pool
func (s *TokenFactorySuite) TestCreationFeeFundsCommunityPool() {
	// Helper to parse DecCoin string like "2678265.860000000000000000uatom"
	parsePoolAmount := func(coinStr string) sdkmath.Int {
		// Remove "uatom" suffix
		amountStr := coinStr[:len(coinStr)-5] // Remove "uatom"
		// Split on decimal point and take integer part
		parts := make([]string, 2)
		if idx := len(amountStr); idx > 0 {
			for i := 0; i < len(amountStr); i++ {
				if amountStr[i] == '.' {
					parts[0] = amountStr[:i]
					parts[1] = amountStr[i+1:]
					break
				}
			}
			if parts[0] == "" {
				parts[0] = amountStr
			}
		}
		intPart := parts[0]
		if intPart == "" {
			intPart = "0"
		}
		amt, ok := sdkmath.NewIntFromString(intPart)
		s.Require().True(ok, "failed to parse amount from "+coinStr)
		return amt
	}

	// Query initial community pool balance
	initialResult, err := s.Chain.QueryJSON(s.GetContext(), `pool.0`, "distribution", "community-pool")
	s.Require().NoError(err)
	initialBalance := parsePoolAmount(initialResult.String())

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(
		"Initial community pool balance: %s uatom", initialBalance)

	// Create a new denom (charges creation fee)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", "communitypooltest",
	)
	s.Require().NoError(err)

	// Query final community pool balance
	finalResult, err := s.Chain.QueryJSON(s.GetContext(), `pool.0`, "distribution", "community-pool")
	s.Require().NoError(err)
	finalBalance := parsePoolAmount(finalResult.String())

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(
		"Final community pool balance: %s uatom", finalBalance)

	// Calculate increase
	increase := finalBalance.Sub(initialBalance)

	// Expected fee is 100 ATOM = 100000000 uatom (from upgrade handler)
	expectedFee := sdkmath.NewInt(100000000)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(
		"Community pool increased by: %s (expected: %s)",
		increase, expectedFee)

	// Verify the increase is at least the creation fee (accounting for potential rounding)
	s.Require().True(increase.GTE(expectedFee),
		"community pool should increase by at least the creation fee")
}

// TestBurnMoreThanBalance tests that burning more than balance fails
func (s *TokenFactorySuite) TestBurnMoreThanBalance() {
	// Create denom and mint tokens
	denom := s.createAndMint("burnfail", 1000000)

	// Attempt to burn more than balance
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "burn",
		fmt.Sprintf("2000000%s", denom),
	)
	s.Require().Error(err)
}

// TestInvalidSubdenom tests that invalid subdenom names are rejected
func (s *TokenFactorySuite) TestInvalidSubdenom() {
	// Test with special characters
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", "invalid@token",
	)
	s.Require().Error(err)

	// Test with spaces
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", "invalid token",
	)
	s.Require().Error(err)
}

// TestQueryAllDenoms tests querying all denoms created by an address
func (s *TokenFactorySuite) TestQueryAllDenoms() {
	// Create multiple denoms
	denom1 := s.createDenom("query1")
	denom2 := s.createDenom("query2")

	// Query all denoms by creator
	denoms, err := s.Chain.QueryJSON(s.GetContext(),
		"denoms", "tokenfactory", "denoms-from-creator",
		s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify our denoms are in the list
	denomsList := denoms.Array()
	s.Require().GreaterOrEqual(len(denomsList), 2)

	denomsMap := make(map[string]bool)
	for _, d := range denomsList {
		denomsMap[d.String()] = true
	}

	s.Require().True(denomsMap[denom1])
	s.Require().True(denomsMap[denom2])
}

func TestTokenFactory(t *testing.T) {
	s := &TokenFactorySuite{
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: true,
			}),
		},
	}
	suite.Run(t, s)
}
