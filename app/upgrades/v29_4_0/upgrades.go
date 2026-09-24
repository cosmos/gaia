package v29_4_0

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gaia/v29/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v29.4.0.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting upgrade", "name", UpgradeName)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		if err := refundFromCommunityPool(ctx, keepers); err != nil {
			return vm, errorsmod.Wrapf(err, "refunding from community pool")
		}

		ctx.Logger().Info("Upgrade complete", "name", UpgradeName)
		return vm, nil
	}
}

var (
	// ErrRecipientNotFound means RefundRecipient has no account on chain.
	ErrRecipientNotFound = errors.New("refund recipient account not found")
	// ErrRecipientNotBase means RefundRecipient is not a plain BaseAccount
	// (vesting, module, etc) and can't be safely converted.
	ErrRecipientNotBase = errors.New("refund recipient is not a base account")
	// ErrRecipientBlocked means bank would refuse sends to RefundRecipient.
	ErrRecipientBlocked = errors.New("refund recipient is a blocked address")
	// ErrInsufficientPool means the community pool, or the distribution
	// module balance backing it, can't cover RefundAmount.
	ErrInsufficientPool = errors.New("community pool can't cover refund")
	// ErrUnlockInPast means RefundUnlockTime is not after the block time.
	ErrUnlockInPast = errors.New("refund unlock time is not in the future")
	// ErrPostCondition means the state after the move isn't what we expect.
	ErrPostCondition = errors.New("refund post condition failed")
)

// refundFromCommunityPool sends RefundAmount from the community pool to
// RefundRecipient and locks it in a delayed vesting account until
// RefundUnlockTime. Any balance the account already had stays spendable.
func refundFromCommunityPool(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	recipient, err := sdk.AccAddressFromBech32(RefundRecipient)
	if err != nil {
		return err
	}

	bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}
	amount := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(RefundAmount)))

	// Pre conditions.
	if ctx.BlockTime().Unix() >= RefundUnlockTime {
		return ErrUnlockInPast
	}
	if keepers.BankKeeper.BlockedAddr(recipient) {
		return ErrRecipientBlocked
	}
	acc := keepers.AccountKeeper.GetAccount(ctx, recipient)
	if acc == nil {
		return ErrRecipientNotFound
	}
	baseAcc, ok := acc.(*authtypes.BaseAccount)
	if !ok {
		return fmt.Errorf("%w: got %T", ErrRecipientNotBase, acc)
	}

	feePool, err := keepers.DistrKeeper.FeePool.Get(ctx)
	if err != nil {
		return err
	}
	poolBefore := feePool.CommunityPool.AmountOf(bondDenom)
	if poolBefore.LT(math.LegacyNewDec(RefundAmount)) {
		return fmt.Errorf("%w: pool has %s%s", ErrInsufficientPool, poolBefore, bondDenom)
	}
	// The pool is just accounting, so check the coins behind it too.
	distrAddr := keepers.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	if distrBal := keepers.BankKeeper.GetBalance(ctx, distrAddr, bondDenom); distrBal.Amount.LT(math.NewInt(RefundAmount)) {
		return fmt.Errorf("%w: distribution module holds %s", ErrInsufficientPool, distrBal)
	}
	balBefore := keepers.BankKeeper.GetBalance(ctx, recipient, bondDenom).Amount

	// Move.
	vestingAcc, err := vestingtypes.NewDelayedVestingAccount(baseAcc, amount, RefundUnlockTime)
	if err != nil {
		return err
	}
	keepers.AccountKeeper.SetAccount(ctx, vestingAcc)

	if err := keepers.DistrKeeper.DistributeFromFeePool(ctx, amount, recipient); err != nil {
		return err
	}

	// Post conditions.
	balAfter := keepers.BankKeeper.GetBalance(ctx, recipient, bondDenom).Amount
	if !balAfter.Sub(balBefore).Equal(math.NewInt(RefundAmount)) {
		return fmt.Errorf("%w: recipient balance %s -> %s", ErrPostCondition, balBefore, balAfter)
	}
	feePool, err = keepers.DistrKeeper.FeePool.Get(ctx)
	if err != nil {
		return err
	}
	if poolAfter := feePool.CommunityPool.AmountOf(bondDenom); !poolBefore.Sub(poolAfter).Equal(math.LegacyNewDec(RefundAmount)) {
		return fmt.Errorf("%w: pool %s -> %s", ErrPostCondition, poolBefore, poolAfter)
	}
	locked := keepers.BankKeeper.LockedCoins(ctx, recipient)
	if !locked.Equal(amount) {
		return fmt.Errorf("%w: locked %s, want %s", ErrPostCondition, locked, amount)
	}

	ctx.Logger().Info("Refunded from community pool", "recipient", RefundRecipient, "amount", amount.String(), "unlock_time", RefundUnlockTime)
	return nil
}
