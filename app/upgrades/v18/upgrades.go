package v18

import (
	"time"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v19/app/keepers"
	"github.com/cosmos/gaia/v19/types"
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
		govParams.ExpeditedThreshold = govv1.DefaultExpeditedThreshold.String()                              // 66.7%
		govParams.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(types.UAtomDenom, sdk.NewInt(500_000_000))) // 500 ATOM
		err = keepers.GovKeeper.SetParams(ctx, govParams)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set gov params")
		}

		err = ConfigureFeeMarketModule(ctx, keepers)
		if err != nil {
			return vm, err
		}

		// Set CosmWasm params
		wasmParams := wasmtypes.DefaultParams()
		wasmParams.CodeUploadAccess = wasmtypes.AllowNobody
		// TODO(reece): only allow specific addresses to instantiate contracts or anyone with AccessTypeEverybody?
		wasmParams.InstantiateDefaultPermission = wasmtypes.AccessTypeAnyOfAddresses
		if err := keepers.WasmKeeper.SetParams(ctx, wasmParams); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set CosmWasm params")
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
	params.FeeDenom = types.UAtomDenom
	params.DistributeFees = false // burn fees
	params.MinBaseGasPrice = sdk.MustNewDecFromStr("0.005")
	params.MaxBlockUtilization = feemarkettypes.DefaultMaxBlockUtilization
	if err := keepers.FeeMarketKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	state, err := keepers.FeeMarketKeeper.GetState(ctx)
	if err != nil {
		return err
	}

	state.BaseGasPrice = sdk.MustNewDecFromStr("0.005")

	return keepers.FeeMarketKeeper.SetState(ctx, state)
}
