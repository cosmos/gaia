package v24

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/gaia/v23/app/keepers"
	liquidkeeper "github.com/cosmos/gaia/v23/x/liquid/keeper"
	liquidtypes "github.com/cosmos/gaia/v23/x/liquid/types"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v24.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		err = MigrateLSMState(ctx, keepers)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "migrating LSM state to x/liquid")
		}

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}

func MigrateLSMState(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	sk := keepers.StakingKeeper
	lsmk := keepers.LiquidKeeper

	err := migrateParams(ctx, sk, lsmk)
	if err != nil {
		return fmt.Errorf("error migrating params: %w", err)
	}

	err = migrateTokenizeShareRecords(ctx, sk, lsmk)
	if err != nil {
		return fmt.Errorf("error migrating tokenize records: %w", err)
	}

	migrateLastTokenizeShareRecordID(ctx, sk, lsmk)
	migrateTotalLiquidStakedTokens(ctx, sk, lsmk)
	migrateTokenizeShareLocks(ctx, sk, lsmk)

	return nil
}

func migrateParams(ctx sdk.Context, sk *stakingkeeper.Keeper, lsmk *liquidkeeper.Keeper) error {
	stakingParams, err := sk.GetParams(ctx)
	if err != nil {
		return err
	}

	liquidParams, err := lsmk.GetParams(ctx)
	if err != nil {
		return err
	}

	liquidParams.GlobalLiquidStakingCap = stakingParams.GlobalLiquidStakingCap
	liquidParams.ValidatorLiquidStakingCap = stakingParams.ValidatorLiquidStakingCap

	return lsmk.SetParams(ctx, liquidParams)
}

func migrateTokenizeShareRecords(ctx sdk.Context, sk *stakingkeeper.Keeper, lsmk *liquidkeeper.Keeper) error {
	tokenizeShareRecords := sk.GetAllTokenizeShareRecords(ctx)
	for _, record := range tokenizeShareRecords {
		lsmRecord := liquidtypes.TokenizeShareRecord{
			Id:            record.Id,
			Owner:         record.Owner,
			ModuleAccount: record.ModuleAccount,
			Validator:     record.Validator,
		}
		if err := lsmk.AddTokenizeShareRecord(ctx, lsmRecord); err != nil {
			return err
		}
	}

	return nil
}

func migrateLastTokenizeShareRecordID(ctx sdk.Context, sk *stakingkeeper.Keeper, lsmk *liquidkeeper.Keeper) {
	lastTokenizeShareRecordID := sk.GetLastTokenizeShareRecordID(ctx)
	lsmk.SetLastTokenizeShareRecordID(ctx, lastTokenizeShareRecordID)
}

func migrateTotalLiquidStakedTokens(ctx sdk.Context, sk *stakingkeeper.Keeper, lsmk *liquidkeeper.Keeper) {
	totalLiquidStaked := sk.GetTotalLiquidStakedTokens(ctx)
	lsmk.SetTotalLiquidStakedTokens(ctx, totalLiquidStaked)
}

func migrateTokenizeShareLocks(ctx sdk.Context, sk *stakingkeeper.Keeper, lsmk *liquidkeeper.Keeper) {
	tokenizeShareLocks := sk.GetAllTokenizeSharesLocks(ctx)
	converted := make([]liquidtypes.TokenizeShareLock, len(tokenizeShareLocks))
	for i, tokenizeShareLock := range tokenizeShareLocks {
		converted[i] = liquidtypes.TokenizeShareLock{
			Address:        tokenizeShareLock.Address,
			Status:         tokenizeShareLock.Status,
			CompletionTime: tokenizeShareLock.CompletionTime,
		}
	}
	lsmk.SetTokenizeShareLocks(ctx, converted)
}
