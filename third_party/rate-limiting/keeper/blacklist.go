package keeper

import (
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v11/types"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Adds a denom to a blacklist to prevent all IBC transfers with this denom
func (k Keeper) AddDenomToBlacklist(ctx sdk.Context, denom string) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Set(key, []byte{1})
}

// Removes a denom from a blacklist to re-enable IBC transfers for that denom
func (k Keeper) RemoveDenomFromBlacklist(ctx sdk.Context, denom string) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.DenomBlacklistKeyPrefix)
	key := types.KeyPrefix(denom)
	store.Delete(key)
}

// Check if a denom is currently blacklisted
func (k Keeper) IsDenomBlacklisted(ctx sdk.Context, denom string) bool {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.DenomBlacklistKeyPrefix)

	key := types.KeyPrefix(denom)
	value := store.Get(key)
	found := len(value) != 0

	return found
}

// Get all the blacklisted denoms
func (k Keeper) GetAllBlacklistedDenoms(ctx sdk.Context) []string {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.DenomBlacklistKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allBlacklistedDenoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		allBlacklistedDenoms = append(allBlacklistedDenoms, string(iterator.Key()))
	}

	return allBlacklistedDenoms
}
