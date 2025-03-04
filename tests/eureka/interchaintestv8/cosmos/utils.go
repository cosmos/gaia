package cosmos

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// CloneAppend returns a new slice with the contents of the provided slices.
func CloneAppend(bz []byte, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))
	copy(res, bz)
	copy(res[len(bz):], tail)
	return
}

// BankBalanceKey returns the store key for a given address and denomination pair.
func BankBalanceKey(addr sdk.AccAddress, denom string) ([]byte, error) {
	keyCodec := collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey)
	cKey := collections.Join(addr, denom)
	return collections.EncodeKeyWithPrefix(banktypes.BalancesPrefix, keyCodec, cKey)
}
