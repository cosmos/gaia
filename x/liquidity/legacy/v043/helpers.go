package v043

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// MigratePrefixAddress is a helper function that migrates all keys of format:
// prefix_bytes | address_bytes
// into format:
// prefix_bytes | address_len (1 byte) | address_bytes
func MigratePrefixAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		// Set new key on store. Values don't change.
		store.Set(append(prefixBz, address.MustLengthPrefix(oldStoreIter.Key())...), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// DeleteDeprecatedPrefix is a helper function that deletes all keys which started the prefix
func DeleteDeprecatedPrefix(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		oldStore.Delete(oldStoreIter.Key())
	}
}
