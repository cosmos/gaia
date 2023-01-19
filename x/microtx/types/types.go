package types

import (
	"encoding/binary"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UInt64FromBytesUnsafe create uint from binary big endian representation
// Note: This is unsafe because the function will panic if provided over 8 bytes
func UInt64FromBytesUnsafe(s []byte) uint64 {
	if len(s) > 8 {
		panic("Invalid uint64 bytes passed to UInt64FromBytes!")
	}
	return binary.BigEndian.Uint64(s)
}

// UInt64Bytes uses the SDK byte marshaling to encode a uint64
func UInt64Bytes(n uint64) []byte {
	return sdk.Uint64ToBigEndian(n)
}

// UInt64FromString to parse out a uint64 for a nonce
func UInt64FromString(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
