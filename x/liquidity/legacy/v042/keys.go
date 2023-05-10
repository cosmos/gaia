// Package v042 is copy-pasted from:
// https://github.com/tendermint/liquidity/blob/v1.2.9/x/liquidity/types/keys.go
package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the liquidity module
	ModuleName = "liquidity"
)

var (
	PoolByReserveAccIndexKeyPrefix = []byte{0x12}

	PoolBatchIndexKeyPrefix = []byte{0x21} // Last PoolBatchIndex
)

// - PoolByReserveAccIndex: `0x12 | ReserveAcc -> Id`
// GetPoolByReserveAccIndexKey returns kv indexing key of the pool indexed by reserve account
func GetPoolByReserveAccIndexKey(reserveAcc sdk.AccAddress) []byte {
	return append(PoolByReserveAccIndexKeyPrefix, reserveAcc.Bytes()...)
}

// GetPoolBatchIndexKey returns kv indexing key of the latest index value of the pool batch
func GetPoolBatchIndexKey(poolID uint64) []byte {
	key := make([]byte, 9)
	key[0] = PoolBatchIndexKeyPrefix[0]
	copy(key[1:9], sdk.Uint64ToBigEndian(poolID))
	return key
}
