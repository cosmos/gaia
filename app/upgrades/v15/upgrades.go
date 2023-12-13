package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v15/app/keepers"
)

// adhere to prop 826 which sets the minimum commission rate to 5% for all validators
// https://www.mintscan.io/cosmos/proposals/826
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

		params := keepers.StakingKeeper.GetParams(ctx)
		params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
		keepers.StakingKeeper.SetParams(ctx, params)

		for _, val := range keepers.StakingKeeper.GetAllValidators(ctx) {
			val := val
			// update validator commission rate if it is less than 5%
			if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.UpdateTime = ctx.BlockHeader().Time
				val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
				keepers.StakingKeeper.SetValidator(ctx, val)
			}
		}

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}
