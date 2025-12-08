package delegator_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryVestingSuite struct {
	*TokenFactoryBaseSuite
}

// createVestingAccount creates a vesting account funded with tokenfactory tokens
func (s *TokenFactoryVestingSuite) createVestingAccount(
	denom string,
	amount int64,
	vestingDuration time.Duration,
) (ibc.Wallet, int64) {
	ctx := s.GetContext()

	vestingEnd := time.Now().Add(vestingDuration).Unix()
	wallet, err := s.Chain.BuildWallet(ctx,
		fmt.Sprintf("vesting-%d", time.Now().UnixNano()), "")
	s.Require().NoError(err)

	// Create vesting account with tokenfactory tokens
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"vesting", "create-vesting-account", wallet.FormattedAddress(),
		fmt.Sprintf("%d%s", amount, denom),
		fmt.Sprintf("%d", vestingEnd))
	s.Require().NoError(err)

	// Fund with gas money (300M uatom for creation fees + tx fees)
	err = s.Chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName,
		ibc.WalletAmount{
			Amount:  sdkmath.NewInt(300_000_000),
			Denom:   s.Chain.Config().Denom,
			Address: wallet.FormattedAddress(),
		})
	s.Require().NoError(err)

	return wallet, vestingEnd
}

// TestDelayedVestingAsAdmin verifies vesting account can perform admin operations during vesting
func (s *TokenFactoryVestingSuite) TestDelayedVestingAsAdmin() {
	ctx := s.GetContext()

	// Create tokenfactory denom for this test
	subdenom := "admintest1"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint tokens to DelegatorWallet
	mintAmt := int64(50_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, mintAmt)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with these tokens (30s vesting)
	vestingWallet, _ := s.createVestingAccount(denom, mintAmt, 30*time.Second)

	// Create another denom with vesting account as admin
	adminDenom, err := s.CreateDenom(vestingWallet, "vadmin")
	s.Require().NoError(err, "failed to create denom with vesting wallet as admin")

	// Verify vesting account is admin
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", adminDenom)
	s.Require().NoError(err)
	s.Require().Equal(vestingWallet.FormattedAddress(), admin.String())

	// Perform admin operations (mint, burn, modify-metadata)
	adminMintAmount := int64(1_000_000)
	err = s.Mint(vestingWallet, adminDenom, adminMintAmount)
	s.Require().NoError(err, "failed to mint tokens with vesting wallet")

	// Verify balance
	balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), adminDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(adminMintAmount), balance)

	// Burn tokens (should succeed)
	burnAmount := int64(500_000)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"tokenfactory", "burn", fmt.Sprintf("%d%s", burnAmount, adminDenom))
	s.Require().NoError(err)

	// Verify reduced balance
	balance, err = s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), adminDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(adminMintAmount-burnAmount), balance)

	// Change admin to DelegatorWallet2
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"tokenfactory", "change-admin", adminDenom, s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err)

	// Verify admin changed
	admin, err = s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", adminDenom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet2.FormattedAddress(), admin.String())

	// VestingWallet can no longer mint (admin transferred)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, adminDenom))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestVestingMultipleAdminOperations verifies vesting accounts can perform multiple admin operations
func (s *TokenFactoryVestingSuite) TestVestingMultipleAdminOperations() {
	ctx := s.GetContext()

	// Create tokenfactory denom for this test
	subdenom := "admintest2"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint tokens to DelegatorWallet
	mintAmt := int64(50_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, mintAmt)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with these tokens (30s vesting)
	vestingWallet, _ := s.createVestingAccount(denom, mintAmt, 30*time.Second)

	// Create denom with VestingWallet
	adminDenom, err := s.CreateDenom(vestingWallet, "admintest2b")
	s.Require().NoError(err, "failed to create denom with vesting wallet")

	// Mint tokens before vesting ends
	adminMintAmount := int64(2_000_000)
	err = s.Mint(vestingWallet, adminDenom, adminMintAmount)
	s.Require().NoError(err, "failed to mint tokens for admin operations test")

	// Verify balance
	balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), adminDenom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(adminMintAmount), balance)

	// Modify metadata (test another admin operation)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"tokenfactory", "modify-metadata", adminDenom, "TEST", "Test Token", "6")
	s.Require().NoError(err)

	// Verify metadata
	metadata, err := s.Chain.QueryJSON(ctx, "metadata", "bank", "denom-metadata", adminDenom)
	s.Require().NoError(err)
	s.Require().Equal("TEST", metadata.Get("symbol").String())

	// Create multiple denoms from same vesting account
	adminDenom2, err := s.CreateDenom(vestingWallet, "admintest3")
	s.Require().NoError(err, "failed to create denom 'admintest3' with vesting wallet")
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", adminDenom2)
	s.Require().NoError(err)
	s.Require().Equal(vestingWallet.FormattedAddress(), admin.String())
}

// TestMintToVestingAccount verifies that mint-to can send tokens directly to a vesting account
// and those tokens are NOT subject to vesting restrictions (since they weren't part of initial vesting)
func (s *TokenFactoryVestingSuite) TestMintToVestingAccount() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	subdenom := "minttovestingtest"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint some tokens to DelegatorWallet for creating vesting account
	initialMintAmount := int64(10_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, initialMintAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with a portion of the tokens (60s vesting)
	vestingAmount := int64(5_000_000_000)
	vestingWallet, _ := s.createVestingAccount(denom, vestingAmount, 60*time.Second)

	// Use mint-to to send additional tokens DIRECTLY to vesting account
	// These tokens should NOT be vested since they weren't part of initial vesting creation
	mintToAmount := int64(3_000_000_000)
	err = s.MintTo(s.DelegatorWallet, denom, mintToAmount, vestingWallet.FormattedAddress())
	s.Require().NoError(err, "mint-to should succeed")

	// Verify vesting account has both vested and mint-to'd tokens
	totalExpected := vestingAmount + mintToAmount
	balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(totalExpected), balance,
		"vesting account should have both vested and mint-to'd tokens")

	// The mint-to'd tokens should be immediately transferable
	// (only the originally vested tokens are locked)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", mintToAmount, denom))
	s.Require().NoError(err, "mint-to'd tokens should be transferable immediately")

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		recipientBalance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(mintToAmount), recipientBalance,
			"recipient should have received mint-to'd tokens")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Verify vesting account still has the vested portion
	remainingBalance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(vestingAmount), remainingBalance,
		"vesting account should still have vested tokens")

	// Attempting to transfer the vested tokens should fail (still locked)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", vestingAmount, denom))
	s.Require().Error(err, "vested tokens should still be locked")
}

// TestVestingAccountCannotTransferUnvestedTokens verifies vested tokenfactory
// tokens cannot be transferred before vesting period ends
func (s *TokenFactoryVestingSuite) TestVestingAccountCannotTransferUnvestedTokens() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	subdenom := "locktest"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint tokens
	vestingAmount := int64(50_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, vestingAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with 60 second vesting period
	vestingWallet, vestingEnd := s.createVestingAccount(denom, vestingAmount, 60*time.Second)

	// IMMEDIATELY attempt transfer (should FAIL - tokens locked)
	// At this point ~32s elapsed (4 txs * 8s)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", vestingAmount, denom))
	s.Require().Error(err, "transfer should fail before vesting ends")

	// Calculate remaining vesting time
	// ~40s elapsed so far (5 txs * 8s), need to wait until vesting end + buffer
	waitDuration := time.Until(time.Unix(vestingEnd, 0)) + (5 * chainsuite.CommitTimeout)
	time.Sleep(waitDuration)

	// Attempt transfer again (should SUCCEED - now vested)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", vestingAmount, denom))
	s.Require().NoError(err, "transfer should succeed after vesting completes")

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(vestingAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

// TestVestingAccountCanTransferNonVestedTokens verifies tokens sent TO a vesting account
// AFTER creation are NOT locked
func (s *TokenFactoryVestingSuite) TestVestingAccountCanTransferNonVestedTokens() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	subdenom := "nonvestedtest"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint tokens
	vestedAmount := int64(30_000_000_000)
	nonVestedAmount := int64(20_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, vestedAmount+nonVestedAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with ONLY vestedAmount (60s vesting)
	vestingWallet, _ := s.createVestingAccount(denom, vestedAmount, 60*time.Second)

	// Send additional tokens TO vesting account (these are NOT vested)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		vestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", nonVestedAmount, denom))
	s.Require().NoError(err)

	// Verify total balance
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(vestedAmount+nonVestedAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Transfer the non-vested tokens (should SUCCEED - only vested tokens locked)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", nonVestedAmount, denom))
	s.Require().NoError(err, "can transfer non-vested tokens")

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(nonVestedAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Attempt to transfer vested tokens (should FAIL - still locked)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", vestedAmount, denom))
	s.Require().Error(err, "cannot transfer vested tokens while vesting")
}

// TestVestingAccountHoldsMultipleDenoms verifies vesting accounts can hold multiple tokenfactory denoms
func (s *TokenFactoryVestingSuite) TestVestingAccountHoldsMultipleDenoms() {
	ctx := s.GetContext()

	// Create 3 different tokenfactory denoms
	denoms := make([]string, 3)
	totalAmount := int64(0)
	for i := 0; i < 3; i++ {
		denom, err := s.CreateDenom(s.DelegatorWallet, fmt.Sprintf("multidenom%d", i))
		s.Require().NoError(err, "failed to create denom 'multidenom%d'", i)
		denoms[i] = denom

		amount := int64(1_000_000 * (i + 1))
		err = s.Mint(s.DelegatorWallet, denom, amount)
		s.Require().NoError(err, "failed to mint tokens for multidenom%d", i)
		totalAmount += amount
	}

	// Create vesting account but only with the first denom vesting (30s)
	vestingWallet, _ := s.createVestingAccount(denoms[0], 1_000_000, 30*time.Second)

	// Transfer the other 2 denoms to vesting wallet (not vested)
	for i := 1; i < 3; i++ {
		amount := int64(1_000_000 * (i + 1))
		_, err := s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
			"bank", "send", s.DelegatorWallet.FormattedAddress(),
			vestingWallet.FormattedAddress(),
			fmt.Sprintf("%d%s", amount, denoms[i]))
		s.Require().NoError(err)
	}

	// Verify all balances on vesting account
	for i, denom := range denoms {
		amount := int64(1_000_000 * (i + 1))
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), denom)
			assert.NoError(c, err)
			assert.Equal(c, sdkmath.NewInt(amount), balance)
		}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	}

	// Transfer non-vested denoms (should succeed immediately)
	for i := 1; i < 3; i++ {
		amount := int64(1_000_000 * (i + 1))
		_, err := s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
			"bank", "send", vestingWallet.FormattedAddress(),
			s.DelegatorWallet2.FormattedAddress(),
			fmt.Sprintf("%d%s", amount, denoms[i]))
		s.Require().NoError(err, "non-vested denom should transfer immediately")
	}

	// Verify transfers succeeded
	for i := 1; i < 3; i++ {
		amount := int64(1_000_000 * (i + 1))
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denoms[i])
			assert.NoError(c, err)
			assert.Equal(c, sdkmath.NewInt(amount), balance)
		}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	}
}

// TestVestingAccountDelegationDoesNotAffectTokenFactory verifies delegating native tokens
// doesn't affect tokenfactory token transfers
func (s *TokenFactoryVestingSuite) TestVestingAccountDelegationDoesNotAffectTokenFactory() {
	ctx := s.GetContext()

	// Create tokenfactory denom
	subdenom := "delegatetest"
	denom, err := s.CreateDenom(s.DelegatorWallet, subdenom)
	s.Require().NoError(err, "failed to create denom")

	// Mint tokenfactory tokens
	tfAmount := int64(50_000_000_000)
	err = s.Mint(s.DelegatorWallet, denom, tfAmount)
	s.Require().NoError(err, "failed to mint tokens")

	// Create vesting account with native uatom (60s vesting)
	nativeAmount := int64(100_000_000)
	nativeDenom := s.Chain.Config().Denom
	vestingEnd := time.Now().Add(60 * time.Second).Unix()
	vestingWallet, err := s.Chain.BuildWallet(ctx,
		fmt.Sprintf("vesting-%d", time.Now().UnixNano()), "")
	s.Require().NoError(err)

	// Create vesting account with native tokens
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"vesting", "create-vesting-account", vestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", nativeAmount, nativeDenom),
		fmt.Sprintf("%d", vestingEnd))
	s.Require().NoError(err)

	// Fund with more gas money
	err = s.Chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName,
		ibc.WalletAmount{
			Amount:  sdkmath.NewInt(300_000_000),
			Denom:   nativeDenom,
			Address: vestingWallet.FormattedAddress(),
		})
	s.Require().NoError(err)

	// Send tokenfactory tokens to vesting account (not vested)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		vestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", tfAmount, denom))
	s.Require().NoError(err)

	// Verify tokenfactory balance
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, vestingWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(tfAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Delegate native uatom to a validator
	validatorWallet := s.Chain.ValidatorWallets[0]
	delegateAmount := int64(10_000_000)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"staking", "delegate", validatorWallet.ValoperAddress,
		fmt.Sprintf("%d%s", delegateAmount, nativeDenom))
	s.Require().NoError(err)

	// Verify can transfer tokenfactory tokens (they're independent and not vested)
	_, err = s.Chain.GetNode().ExecTx(ctx, vestingWallet.KeyName(),
		"bank", "send", vestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", tfAmount, denom))
	s.Require().NoError(err, "tokenfactory tokens should be transferable regardless of native delegation")

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(tfAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

func TestTokenFactoryVesting(t *testing.T) {
	s := &TokenFactoryVestingSuite{
		TokenFactoryBaseSuite: &TokenFactoryBaseSuite{
			Suite: &delegator.Suite{
				Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
					UpgradeOnSetup: true,
				}),
			},
		},
	}
	suite.Run(t, s)
}
