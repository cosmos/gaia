package v23_2_0 //nolint:revive

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v23/app/keepers"
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

		ctx.Logger().Info("Starting Migrate IBC Wasm...")
		if err = MigrateIBCWasm(ctx); err != nil {
			ctx.Logger().Info("Error running migrate for IBC Wasm client", "message", err.Error())
		}

		ctx.Logger().Info("Upgrade v23.2.0 complete")
		return vm, nil
	}
}

func MigrateIBCWasm(_ sdk.Context) error {
	return nil
}
