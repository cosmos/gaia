package v29_0_0

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v29/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v29.0.0.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting upgrade", "name", UpgradeName)

		ctx.Logger().Info("Deleting contents of the deprecated provider kv-store...")
		if err := deleteProviderStoreContents(ctx, keepers); err != nil {
			return vm, errorsmod.Wrapf(err, "deleting provider store contents")
		}

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("Upgrade complete", "name", UpgradeName)
		return vm, nil
	}
}

// deleteProviderStoreContents empties every key/value in the deprecated ICS
// "provider" kv-store — the same operation as rootmulti's deleteKVStore
// helper (collect keys via an iterator, since a KVStore can't be written to
// while iterating, then Delete each one), run as an ordinary state mutation
// instead of via StoreUpgrades.Deleted. See providerStoreKey's doc comment
// in constants.go for why.
func deleteProviderStoreContents(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	store := ctx.KVStore(keepers.GetKey(providerStoreKey))

	var keys [][]byte
	itr := store.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, itr.Key())
	}
	if err := itr.Close(); err != nil {
		return err
	}
	for _, key := range keys {
		store.Delete(key)
	}

	ctx.Logger().Info("Deleted provider store contents", "keys_deleted", len(keys))
	return nil
}
