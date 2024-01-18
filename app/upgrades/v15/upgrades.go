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

		UpgradeCommissionRate(ctx, keepers)

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// UpgradeCommissionRate sets the minimum commission rate staking parameter to 5%
// and updates the commission rate for all validators that have a commission rate less than 5%
func UpgradeCommissionRate(ctx sdk.Context, keepers *keepers.AppKeepers) {
	params := keepers.StakingKeeper.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	err := keepers.StakingKeeper.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}

	for _, val := range keepers.StakingKeeper.GetAllValidators(ctx) {
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commission rate to 5%
			val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
			// set the max rate to 5% if it is less than 5%
			if val.Commission.CommissionRates.MaxRate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.CommissionRates.MaxRate = sdk.NewDecWithPrec(5, 2)
			}
			val.Commission.UpdateTime = ctx.BlockHeader().Time
			keepers.StakingKeeper.SetValidator(ctx, val)
		}
	}
}
