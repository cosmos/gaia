package v22

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	providerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/provider/keeper"
	providertypes "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v22/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v22.
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

		infractionParameters, err := providertypes.DefaultConsumerInfractionParameters(ctx, keepers.SlashingKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}
		if err := SetConsumerInfractionParams(ctx, keepers.ProviderKeeper, infractionParameters); err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("Upgrade v22 complete")
		return vm, nil
	}
}

func SetConsumerInfractionParams(ctx sdk.Context, pk providerkeeper.Keeper, infractionParameters providertypes.InfractionParameters) error {
	activeConsumerIDs := pk.GetAllActiveConsumerIds(ctx)
	for _, consumerID := range activeConsumerIDs {
		if err := pk.SetInfractionParameters(ctx, consumerID, infractionParameters); err != nil {
			return err
		}
	}

	return nil
}
