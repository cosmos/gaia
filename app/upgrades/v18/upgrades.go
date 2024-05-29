package v18

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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

		err = ConfigureFeeMarketModule(ctx, keepers)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade v18 complete")
		return vm, nil
	}
}

func ConfigureFeeMarketModule(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	params, err := keepers.FeeMarketKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	params.Enabled = true
	params.FeeDenom = "uatom"
	params.DistributeFees = true
	params.MinBaseFee = sdk.MustNewDecFromStr("0.005")
	params.TargetBlockUtilization = 50000000
	params.MaxBlockUtilization = 100000000
	if err := keepers.FeeMarketKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	state, err := keepers.FeeMarketKeeper.GetState(ctx)
	if err != nil {
		return err
	}

	state.BaseFee = sdk.MustNewDecFromStr("0.005")

	return keepers.FeeMarketKeeper.SetState(ctx, state)
}
