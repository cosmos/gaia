package v18

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v18/app/keepers"
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

		expeditedPeriod := 24 * 7 * time.Hour // 7 days
		govParams := keepers.GovKeeper.GetParams(ctx)
		govParams.ExpeditedVotingPeriod = &expeditedPeriod
		govParams.ExpeditedThreshold = govv1.DefaultExpeditedThreshold.String() // 66.7%
		govParams.ExpeditedMinDeposit = govParams.MinDeposit                    // full deposit amount is required
		keepers.GovKeeper.SetParams(ctx, govParams)

		ctx.Logger().Info("Upgrade v18 complete")
		return vm, nil
	}
}
