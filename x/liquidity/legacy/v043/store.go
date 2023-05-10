package v043

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v042liquidity "github.com/cosmos/gaia/v9/x/liquidity/legacy/v042"
)

// MigrateStore performs in-place store migrations from v0.42 to v0.43. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey) error {
	store := ctx.KVStore(storeKey)

	// old key format v042:
	// PoolByReserveAccIndex: `0x12 | ReserveAcc -> ProtocolBuffer(uint64)`
	// PoolBatchIndex: `0x21 | PoolId -> ProtocolBuffer(uint64)`
	// new key format v043:
	// PoolByReserveAccIndex: `0x12 | ReserveAccLen (1 byte) | ReserveAcc -> ProtocolBuffer(uint64)`
	// PoolBatchIndex: deprecated
	MigratePrefixAddress(store, v042liquidity.PoolByReserveAccIndexKeyPrefix)
	DeleteDeprecatedPrefix(store, v042liquidity.PoolBatchIndexKeyPrefix)
	return nil
}
