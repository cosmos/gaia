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

		// Set the minimum height of a valid consumer equivocation evidence
		// for the existing consumer chains: neutron-1 and stride-1
		keepers.ProviderKeeper.SetEquivocationEvidenceMinHeight(ctx, "neutron-1", 4552189)
		keepers.ProviderKeeper.SetEquivocationEvidenceMinHeight(ctx, "stride-1", 6375035)

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
