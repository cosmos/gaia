package v23

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	tokenfactorykeeper "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"

	"github.com/cosmos/gaia/v23/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v23.
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

		err = setTokenFactoryParams(ctx, keepers.TokenFactoryKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "setting token factory params")
		}

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}

func setTokenFactoryParams(ctx sdk.Context, keeper tokenfactorykeeper.Keeper) error {
	return keeper.SetParams(ctx, tokenfactorytypes.Params{
		// TODO(wllmshao): set this to a fee we agree on
		DenomCreationFee: sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(1000000))),
	})
}
