package v27_6_0 //nolint:revive

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v27/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v27.
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

		// Governance proposal 1048, which carried the MsgRecoverClient for the
		// expired client below, failed. The client is therefore recovered as part
		// of this software upgrade instead.
		if ctx.ChainID() == MainnetChainID {
			// Substitute the IBC light client for the connection with the
			// gravity-bridge-3 chain.
			if err := keepers.IBCKeeper.ClientKeeper.RecoverClient(ctx, "07-tendermint-582", "07-tendermint-1496"); err != nil {
				// A failed recovery must not halt the chain at the upgrade height.
				ctx.Logger().Error("failed to recover gravity-bridge-3 client", "error", err)
			} else {
				ctx.Logger().Info("recovered gravity-bridge-3 client", "client-id", "07-tendermint-582")
			}
		}

		ctx.Logger().Info("Upgrade complete", "name", UpgradeName)
		return vm, nil
	}
}
