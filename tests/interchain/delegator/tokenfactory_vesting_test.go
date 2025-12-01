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
	*delegator.Suite
	VestingWallet ibc.Wallet // DelayedVestingAccount (10 seconds)
}

func (s *TokenFactoryVestingSuite) SetupSuite() {
	s.Suite.SetupSuite()
	ctx := s.GetContext()

	// Create delayed vesting account (10 second vesting period)
	vestingEnd := time.Now().Add(10 * time.Second).Unix()
	wallet, err := s.Chain.BuildWallet(ctx, fmt.Sprintf("vesting-delayed-%d", vestingEnd), "")
	s.Require().NoError(err)
	s.VestingWallet = wallet

	vestingAmount := sdkmath.NewInt(50_000_000_000)
	_, err = s.Chain.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName,
		"vesting", "create-vesting-account", s.VestingWallet.FormattedAddress(),
		fmt.Sprintf("%s%s", vestingAmount.String(), s.Chain.Config().Denom),
		fmt.Sprintf("%d", vestingEnd))
	s.Require().NoError(err)

	// Give vesting account gas money
	err = s.Chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Amount:  sdkmath.NewInt(10_000_000),
		Denom:   s.Chain.Config().Denom,
		Address: s.VestingWallet.FormattedAddress(),
	})
	s.Require().NoError(err)
}

// createDenom creates a tokenfactory denom using the specified wallet
func (s *TokenFactoryVestingSuite) createDenom(wallet ibc.Wallet, subdenom string) string {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		wallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", wallet.FormattedAddress(), subdenom)
}

// mint mints tokens for a given denom using the specified wallet
func (s *TokenFactoryVestingSuite) mint(wallet ibc.Wallet, denom string, amount int64) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		wallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	s.Require().NoError(err)
}

// TestDelayedVestingAsAdmin verifies vesting account can perform admin operations during vesting
func (s *TokenFactoryVestingSuite) TestDelayedVestingAsAdmin() {
	ctx := s.GetContext()

	// Create denom with VestingWallet (becomes admin)
	denom := s.createDenom(s.VestingWallet, "admintest1")

	// Verify admin
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.VestingWallet.FormattedAddress(), admin.String())

	// Mint tokens (should succeed - admin operations allowed during vesting)
	mintAmount := int64(1000000)
	s.mint(s.VestingWallet, denom, mintAmount)

	// Verify balance
	balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)

	// Burn tokens (should succeed)
	burnAmount := int64(500000)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"tokenfactory", "burn", fmt.Sprintf("%d%s", burnAmount, denom))
	s.Require().NoError(err)

	// Verify reduced balance
	balance, err = s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-burnAmount), balance)

	// Change admin to DelegatorWallet2
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"tokenfactory", "change-admin", denom, s.DelegatorWallet2.FormattedAddress())
	s.Require().NoError(err)

	// Verify admin changed
	admin, err = s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet2.FormattedAddress(), admin.String())

	// VestingWallet can no longer mint (admin transferred)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestVestingMultipleAdminOperations verifies vesting accounts can perform multiple admin operations
func (s *TokenFactoryVestingSuite) TestVestingMultipleAdminOperations() {
	ctx := s.GetContext()

	// Create denom with VestingWallet
	denom := s.createDenom(s.VestingWallet, "admintest2")

	// Mint tokens before vesting ends
	mintAmount := int64(2000000)
	s.mint(s.VestingWallet, denom, mintAmount)

	// Verify balance
	balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)

	// Modify metadata (test another admin operation)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"tokenfactory", "modify-metadata", denom, "TEST", "Test Token", "6")
	s.Require().NoError(err)

	// Verify metadata
	metadata, err := s.Chain.QueryJSON(ctx, "metadata", "bank", "denom-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal("TEST", metadata.Get("symbol").String())

	// Create multiple denoms from same vesting account
	denom2 := s.createDenom(s.VestingWallet, "admintest3")
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom2)
	s.Require().NoError(err)
	s.Require().Equal(s.VestingWallet.FormattedAddress(), admin.String())
}

// TestVestingAccountCannotTransferUnvestedTokens verifies vesting accounts
// can hold tokenfactory tokens but cannot transfer unvested amounts
func (s *TokenFactoryVestingSuite) TestVestingAccountCannotTransferUnvestedTokens() {
	ctx := s.GetContext()

	// Create denom with DelegatorWallet
	denom := s.createDenom(s.DelegatorWallet, "transfertest")

	// Mint tokenfactory tokens to DelegatorWallet
	mintAmount := int64(5000000)
	s.mint(s.DelegatorWallet, denom, mintAmount)

	// Transfer tokenfactory tokens to VestingWallet (still vesting)
	_, err := s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		s.VestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", mintAmount, denom))
	s.Require().NoError(err)

	// Verify VestingWallet has the tokenfactory tokens
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(mintAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Attempt bank send of ALL tokenfactory tokens (should FAIL - unvested)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"bank", "send", s.VestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", mintAmount, denom))
	s.Require().Error(err)

	// Wait for vesting period to complete
	time.Sleep(15 * time.Second)

	// Attempt bank send again (should SUCCEED - now vested)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"bank", "send", s.VestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", mintAmount, denom))
	s.Require().NoError(err)

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(mintAmount), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

// TestVestingAccountHoldsMultipleDenoms verifies vesting accounts can hold multiple tokenfactory denoms
func (s *TokenFactoryVestingSuite) TestVestingAccountHoldsMultipleDenoms() {
	ctx := s.GetContext()

	// Create 3 different tokenfactory denoms with DelegatorWallet
	denoms := make([]string, 3)
	for i := 0; i < 3; i++ {
		denoms[i] = s.createDenom(s.DelegatorWallet, fmt.Sprintf("multidenom%d", i))
		s.mint(s.DelegatorWallet, denoms[i], 1000000)
	}

	// Transfer all 3 denoms to VestingWallet (still vesting)
	for _, denom := range denoms {
		_, err := s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
			"bank", "send", s.DelegatorWallet.FormattedAddress(),
			s.VestingWallet.FormattedAddress(),
			fmt.Sprintf("%d%s", 1000000, denom))
		s.Require().NoError(err)
	}

	// Verify all balances on vesting account
	for _, denom := range denoms {
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
			assert.NoError(c, err)
			assert.Equal(c, sdkmath.NewInt(1000000), balance)
		}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	}

	// Wait for vesting to complete
	time.Sleep(15 * time.Second)

	// Transfer all 3 denoms out successfully
	for _, denom := range denoms {
		_, err := s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
			"bank", "send", s.VestingWallet.FormattedAddress(),
			s.DelegatorWallet2.FormattedAddress(),
			fmt.Sprintf("%d%s", 1000000, denom))
		s.Require().NoError(err)
	}

	// Verify all transfers succeeded
	for _, denom := range denoms {
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
			assert.NoError(c, err)
			assert.Equal(c, sdkmath.NewInt(1000000), balance)
		}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	}
}

// TestDelayedVestingCannotTransferBeforeEnd verifies delayed vesting is all-or-nothing
func (s *TokenFactoryVestingSuite) TestDelayedVestingCannotTransferBeforeEnd() {
	ctx := s.GetContext()

	// Create denom and mint
	denom := s.createDenom(s.DelegatorWallet, "delayedtest")
	s.mint(s.DelegatorWallet, denom, 3000000)

	// Fund DelayedVestingAccount with tokenfactory tokens
	_, err := s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		s.VestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", 3000000, denom))
	s.Require().NoError(err)

	// Verify balance
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(3000000), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Verify cannot transfer ANY amount before end time
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"bank", "send", s.VestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", 100, denom))
	s.Require().Error(err)

	// Wait for end time
	time.Sleep(15 * time.Second)

	// Verify can transfer all tokens
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"bank", "send", s.VestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", 3000000, denom))
	s.Require().NoError(err)

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(3000000), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

// TestVestingAccountDelegationDoesNotAffectTokenFactory verifies delegating native tokens
// doesn't affect tokenfactory token transfers
func (s *TokenFactoryVestingSuite) TestVestingAccountDelegationDoesNotAffectTokenFactory() {
	ctx := s.GetContext()

	// Create tokenfactory denom and mint to VestingWallet
	denom := s.createDenom(s.DelegatorWallet, "delegatetest")
	s.mint(s.DelegatorWallet, denom, 2000000)

	_, err := s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"bank", "send", s.DelegatorWallet.FormattedAddress(),
		s.VestingWallet.FormattedAddress(),
		fmt.Sprintf("%d%s", 2000000, denom))
	s.Require().NoError(err)

	// Verify balance
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.VestingWallet.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(2000000), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Delegate uatom to a validator
	validatorWallet := s.Chain.ValidatorWallets[0]
	delegateAmount := int64(10_000_000)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"staking", "delegate", validatorWallet.ValoperAddress,
		fmt.Sprintf("%d%s", delegateAmount, s.Chain.Config().Denom))
	s.Require().NoError(err)

	// Wait for vesting to complete
	time.Sleep(15 * time.Second)

	// Verify can still transfer tokenfactory tokens (they're independent)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.VestingWallet.KeyName(),
		"bank", "send", s.VestingWallet.FormattedAddress(),
		s.DelegatorWallet2.FormattedAddress(),
		fmt.Sprintf("%d%s", 2000000, denom))
	s.Require().NoError(err)

	// Verify transfer succeeded
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet2.FormattedAddress(), denom)
		assert.NoError(c, err)
		assert.Equal(c, sdkmath.NewInt(2000000), balance)
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

func TestTokenFactoryVesting(t *testing.T) {
	s := &TokenFactoryVestingSuite{
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: true,
			}),
		},
	}
	suite.Run(t, s)
}
