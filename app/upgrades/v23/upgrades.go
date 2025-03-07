package v23

import (
	"context"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/cosmos/gaia/v23/app/keepers"
)

// CreateRCUpgradeHandler returns an upgrade handler for Gaia v23.0.0-rc3.
// This should only be executed on networks which
func CreateRCUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting custom migration...")

		if err := AuthzGrantWasmLightClient(c, keepers.AuthzKeeper, *keepers.GovKeeper); err != nil {
			ctx.Logger().Error("Error running authz grant for ibc wasm client", "message", err.Error())
			return vm, err
		}

		ctx.Logger().Info("Upgrade v23.0.0-rc3 complete")
		return vm, nil
	}
}

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
		ctx.Logger().Info("Setting IBC Client AllowedClients")
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = []string{ibctmtypes.ModuleName, ibcwasmtypes.ModuleName}
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		ctx.Logger().Info("Running authz ibc wasm client grant")
		if err := AuthzGrantWasmLightClient(ctx, keepers.AuthzKeeper, *keepers.GovKeeper); err != nil {
			ctx.Logger().Error("Error running authz grant for ibc wasm client", "message", err.Error())
			return nil, err
		}

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}

func AuthzGrantWasmLightClient(ctx context.Context, authzKeeper authzkeeper.Keeper, govKeeper govkeeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	grant, err := authz.NewGrant(
		sdkCtx.BlockTime(),
		authz.NewGenericAuthorization(IBCWasmStoreCodeTypeURL),
		nil,
	)
	if err != nil {
		return err
	}
	sdkCtx.Logger().Info("Granting IBC Wasm Store Code", "granter", govKeeper.GetAuthority(), "grantee", ClientUploaderAddress)
	resp, err := authzKeeper.Grant(ctx, &authz.MsgGrant{
		Granter: govKeeper.GetAuthority(),
		Grantee: ClientUploaderAddress,
		Grant:   grant,
	})
	if err != nil {
		return err
	}
	if resp != nil {
		sdkCtx.Logger().Info("Authz Keeper Grant", "response", resp.String())
	}
	return nil
}
