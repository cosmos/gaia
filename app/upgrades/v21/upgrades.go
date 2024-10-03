package v21

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/cosmos/gaia/v21/app/keepers"
)

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
			return vm, err
		}

		err = InitializeConstitutionCollection(ctx, *keepers.GovKeeper)
		if err != nil {
			ctx.Logger().Error("Error initializing Constitution Collection:", "message", err.Error())
		}

		ctx.Logger().Info("Upgrade v21 complete")
		return vm, nil
	}
}

// setting the default constitution for the chain
// this is in line with cosmos-sdk v5 gov migration: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/gov/migrations/v5/store.go#L57
func InitializeConstitutionCollection(ctx sdk.Context, govKeeper govkeeper.Keeper) error {
	return govKeeper.Constitution.Set(ctx, "This chain has no constitution.")
}
