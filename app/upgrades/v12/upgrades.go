package v12

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v18/app/keepers"
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

		// Set liquid staking module parameters
		params, err := keepers.StakingKeeper.GetParams(ctx)
		if err != nil {
			return vm, err
		}
		params.ValidatorBondFactor = ValidatorBondFactor
		params.ValidatorLiquidStakingCap = ValidatorLiquidStakingCap
		params.GlobalLiquidStakingCap = GlobalLiquidStakingCap

		err = keepers.StakingKeeper.SetParams(ctx, params)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, nil
	}
}
