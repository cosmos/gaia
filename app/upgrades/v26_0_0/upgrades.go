package v26_0_0 //nolint:revive

import (
	"context"

	tokenfactorytypes "github.com/cosmos/tokenfactory/x/tokenfactory/types"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v26/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v26.
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

		// Initialize tokenfactory params
		ctx.Logger().Info("Initializing tokenfactory module...")
		tokenfactoryParams := tokenfactorytypes.DefaultParams()
		tokenfactoryParams.DenomCreationFee = sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000000)) // 1 ATOM
		if err := keepers.TokenFactoryKeeper.SetParams(ctx, tokenfactoryParams); err != nil {
			return vm, errorsmod.Wrapf(err, "setting tokenfactory params")
		}
		ctx.Logger().Info("Tokenfactory module initialized successfully")

		ctx.Logger().Info("Upgrade v26 complete")
		return vm, nil
	}
}
