package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the lsm module
	ModuleName = "lsm"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the lsm module
	RouterKey = ModuleName

	// Prefix for module accounts that custodian tokenized shares
	TokenizeShareModuleAccountPrefix = "tokenizeshare_"
)

var (
	// Keys for store prefixes
	// Last* values are constant during a block.
	ParamsKey = []byte{0x51} // prefix for parameters for module x/lsm

	TokenizeShareRecordPrefix          = []byte{0x81} // key for tokenizeshare record prefix
	TokenizeShareRecordIDByOwnerPrefix = []byte{0x82} // key for tokenizeshare record id by owner prefix
	TokenizeShareRecordIDByDenomPrefix = []byte{0x83} // key for tokenizeshare record id by denom prefix
	LastTokenizeShareRecordIDKey       = []byte{0x84} // key for last tokenize share record id
	TotalLiquidStakedTokensKey         = []byte{0x85} // key for total liquid staked tokens
	TokenizeSharesLockPrefix           = []byte{0x86} // key for locking tokenize shares
	TokenizeSharesUnlockQueuePrefix    = []byte{0x87} // key for the queue that unlocks tokenize shares
)

// GetTokenizeShareRecordByIndexKey returns the key of the specified id. Intended for querying the tokenizeShareRecord by the id.
func GetTokenizeShareRecordByIndexKey(id uint64) []byte {
	return append(TokenizeShareRecordPrefix, sdk.Uint64ToBigEndian(id)...)
}

// GetTokenizeShareRecordIdsByOwnerPrefix returns the key of the specified owner. Intended for querying all tokenizeShareRecords of an owner
func GetTokenizeShareRecordIdsByOwnerPrefix(owner sdk.AccAddress) []byte {
	return append(TokenizeShareRecordIDByOwnerPrefix, address.MustLengthPrefix(owner)...)
}

// GetTokenizeShareRecordIdByOwnerAndIdKey returns the key of the specified owner and id. Intended for setting tokenizeShareRecord of an owner
func GetTokenizeShareRecordIDByOwnerAndIDKey(owner sdk.AccAddress, id uint64) []byte {
	return append(append(TokenizeShareRecordIDByOwnerPrefix, address.MustLengthPrefix(owner)...), sdk.Uint64ToBigEndian(id)...)
}

func GetTokenizeShareRecordIDByDenomKey(denom string) []byte {
	return append(TokenizeShareRecordIDByDenomPrefix, []byte(denom)...)
}

// GetTokenizeSharesLockKey returns the key for storing a tokenize share lock for a specified account
func GetTokenizeSharesLockKey(owner sdk.AccAddress) []byte {
	return append(TokenizeSharesLockPrefix, address.MustLengthPrefix(owner)...)
}

// GetTokenizeShareAuthorizationTimeKey returns the prefix key used for getting a set of pending
// tokenize share unlocks that complete at the given time
func GetTokenizeShareAuthorizationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(TokenizeSharesUnlockQueuePrefix, bz...)
}
