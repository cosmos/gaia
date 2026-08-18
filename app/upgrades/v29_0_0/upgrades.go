package v29_0_0

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
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

		ctx.Logger().Info("Deleting contents of the now-unused x/params subspaces...")
		if err := deleteLegacyParamSubspaces(ctx, keepers); err != nil {
			return vm, errorsmod.Wrapf(err, "deleting legacy params subspaces")
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

// legacyParamSubspaces lists the x/params subspace names that were
// registered by the now-removed initParamsKeeper (part of the
// "chore/remove-params-module" work) and are no longer read by any module —
// each of these already migrated its own params out of x/params into its own
// store in a prior upgrade, so this data has been dead weight ever since.
//
// "ratelimit" is deliberately excluded: the ratelimit module keeper still
// depends on a live x/params Subspace at runtime (see
// app/keepers/keepers.go), so its data must be preserved.
var legacyParamSubspaces = []string{
	"auth", "staking", "bank", "mint", "distribution", "slashing", "gov",
	"ibc", "transfer", "icacontroller", "icahost", "packetfowardmiddleware",
	"wasm", "tokenfactory",
}

// deleteLegacyParamSubspaces empties the stale per-module key ranges left
// behind in the x/params kv-store by legacyParamSubspaces. The params
// kv-store itself is NOT deleted via StoreUpgrades.Deleted — it stays
// mounted (see app/keepers/keys.go) because the ratelimit module keeper
// still needs a live Subspace backed by it. Wiping the stale key ranges as
// an ordinary state mutation, rather than deleting the whole store, avoids
// the same "version mismatch" LoadVersion risk described in
// providerStoreKey's doc comment in constants.go — just scoped per-subspace
// instead of per-store, since the store as a whole must survive.
//
// Each subspace's keys are prefixed with "<name>/" (see x/params's own
// Subspace.kvStore in cosmos-sdk), so prefix.NewStore reproduces exactly the
// view each subspace itself used to see.
func deleteLegacyParamSubspaces(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	store := ctx.KVStore(keepers.GetKey(paramstypes.StoreKey))

	var deleted int
	for _, name := range legacyParamSubspaces {
		subStore := prefix.NewStore(store, append([]byte(name), '/'))

		var keys [][]byte
		itr := subStore.Iterator(nil, nil)
		for ; itr.Valid(); itr.Next() {
			keys = append(keys, itr.Key())
		}
		if err := itr.Close(); err != nil {
			return err
		}
		for _, key := range keys {
			subStore.Delete(key)
		}
		deleted += len(keys)
	}

	ctx.Logger().Info("Deleted legacy params subspace contents", "keys_deleted", deleted)
	return nil
}
