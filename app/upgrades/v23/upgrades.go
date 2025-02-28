package v23

import (
	"context"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

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

		// Set IBC Client AllowedClients
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = []string{ibctmtypes.ModuleName, ibcwasmtypes.ModuleName}
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}
