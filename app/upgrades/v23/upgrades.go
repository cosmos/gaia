package v23

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"

	"github.com/cosmos/gaia/v23/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v23.
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
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		// Add the Wasm client type to the allowed clients
		Add08WasmToAllowedClients(ctx, keepers.IBCKeeper.ClientKeeper)

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}

func Add08WasmToAllowedClients(ctx sdk.Context, clientKeeper clientkeeper.Keeper) {
	// explicitly update the IBC 02-client params, adding the wasm client type
	params := clientKeeper.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, ibcwasmtypes.Wasm)
	clientKeeper.SetParams(ctx, params)
}
