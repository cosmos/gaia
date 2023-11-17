package v14

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v14/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Set the equivocation evidence min height for the Neutron and Stride consumer chains
		keepers.ProviderKeeper.SetEquivocationEvidenceMinHeight(ctx, NeutronChainID, EquivocationEvidenceMinHeight)
		keepers.ProviderKeeper.SetEquivocationEvidenceMinHeight(ctx, NeutronChainID, EquivocationEvidenceMinHeight)

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
