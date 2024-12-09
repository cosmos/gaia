package v22

import (
	"context"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	providerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/provider/keeper"
	providertypes "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

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

		if err := SetConsumerInfractionParams(ctx, keepers.ProviderKeeper); err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade v22 complete")
		return vm, nil
	}
}

func SetConsumerInfractionParams(ctx sdk.Context, pk providerkeeper.Keeper) error {
	infractionParameters := DefaultInfractionParams()

	activeConsumerIDs := pk.GetAllActiveConsumerIds(ctx)
	for _, consumerID := range activeConsumerIDs {
		if err := pk.SetInfractionParameters(ctx, consumerID, infractionParameters); err != nil {
			return err
		}
	}

	return nil
}

func DefaultInfractionParams() providertypes.InfractionParameters {
	return providertypes.InfractionParameters{
		DoubleSign: &providertypes.SlashJailParameters{
			JailDuration:  time.Duration(1<<63 - 1),        // the largest value a time.Duration can hold 9223372036854775807 (approximately 292 years)
			SlashFraction: math.LegacyNewDecWithPrec(5, 2), // 0.05
		},
		Downtime: &providertypes.SlashJailParameters{
			JailDuration:  600 * time.Second,
			SlashFraction: math.LegacyNewDec(0), // no slashing for downtime on the consumer
		},
	}
}
