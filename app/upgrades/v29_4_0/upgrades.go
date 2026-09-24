package v29_4_0

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

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

	acc := keepers.AccountKeeper.GetAccount(ctx, recipient)
	if acc == nil {
		acc = keepers.AccountKeeper.NewAccountWithAddress(ctx, recipient)
	}
	baseAcc, ok := acc.(*authtypes.BaseAccount)
	if !ok {
		return fmt.Errorf("recipient %s is %T, want *BaseAccount", RefundRecipient, acc)
	}

	vestingAcc, err := vestingtypes.NewDelayedVestingAccount(baseAcc, amount, RefundUnlockTime)
	if err != nil {
		return err
	}
	keepers.AccountKeeper.SetAccount(ctx, vestingAcc)

	if err := keepers.DistrKeeper.DistributeFromFeePool(ctx, amount, recipient); err != nil {
		return err
	}

	ctx.Logger().Info("Refunded from community pool", "recipient", RefundRecipient, "amount", amount.String(), "unlock_time", RefundUnlockTime)
	return nil
}
