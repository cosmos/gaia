package v18

import (
	errorsmod "cosmossdk.io/errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
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
