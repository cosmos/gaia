package v29_0_0

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

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

		ctx.Logger().Info("Deleting contents of the now-unused x/params kv-store...")
		if err := deleteLegacyParamsStoreContents(ctx, keepers); err != nil {
			return vm, errorsmod.Wrapf(err, "deleting legacy params store contents")
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

// deleteAllStoreKeys empties every key/value in store — the same operation
// as rootmulti's deleteKVStore helper (collect keys via an iterator, since a
// KVStore can't be written to while iterating, then Delete each one) — and
// returns how many keys were removed.
func deleteAllStoreKeys(store storetypes.KVStore) (int, error) {
	var keys [][]byte
	itr := store.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		keys = append(keys, itr.Key())
	}
	if err := itr.Close(); err != nil {
		return 0, err
	}
	for _, key := range keys {
		store.Delete(key)
	}
	return len(keys), nil
}

// deleteProviderStoreContents empties every key/value in the deprecated ICS
// "provider" kv-store, run as an ordinary state mutation instead of via
// StoreUpgrades.Deleted. See providerStoreKey's doc comment in constants.go
// for why.
func deleteProviderStoreContents(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	store := ctx.KVStore(keepers.GetKey(providerStoreKey))
	deleted, err := deleteAllStoreKeys(store)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Deleted provider store contents", "keys_deleted", deleted)
	return nil
}

// deleteLegacyParamsStoreContents empties the entire x/params kv-store.
// Every subspace ever registered there has long since migrated its
// params out of x/params into its own store, and as of this same release the
// ratelimit module keeper — the last holdout, see app/keepers/keepers.go —
// is also constructed with a zero-value Subspace rather than one backed by
// this store. So nothing reads any part of this store live any more, and a
// full wipe (rather than deleting known subspace prefixes one by one) can't
// miss a name.
//
// The store itself is not deleted via StoreUpgrades.Deleted; it stays
// mounted (see app/keepers/keys.go) and its contents are wiped as an
// ordinary state mutation, rather than the store being deleted outright, to
// avoid the same "version mismatch" LoadVersion risk described in
// providerStoreKey's doc comment in constants.go (the store itself must survive).
func deleteLegacyParamsStoreContents(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	store := ctx.KVStore(keepers.GetKey(paramstypes.StoreKey))
	deleted, err := deleteAllStoreKeys(store)
	if err != nil {
		return err
	}

	ctx.Logger().Info("Deleted legacy params store contents", "keys_deleted", deleted)
	return nil
}
