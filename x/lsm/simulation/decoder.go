package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/cosmos/gaia/v22/x/lsm/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding lsm type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.TokenizeShareRecordPrefix):
			var recordA, recordB types.TokenizeShareRecord

			cdc.MustUnmarshal(kvA.Value, &recordA)
			cdc.MustUnmarshal(kvB.Value, &recordB)

			return fmt.Sprintf("%v\n%v", recordA, recordB)
		case bytes.Equal(kvA.Key[:1], types.TokenizeShareRecordIDByOwnerPrefix),
			bytes.Equal(kvA.Key[:1], types.TokenizeShareRecordIDByDenomPrefix),
			bytes.Equal(kvA.Key[:1], types.LastTokenizeShareRecordIDKey):
			var idA, idB uint64

			idA = sdk.BigEndianToUint64(kvA.Value)
			idB = sdk.BigEndianToUint64(kvB.Value)

			return fmt.Sprintf("%v\n%v", idA, idB)
		case bytes.Equal(kvA.Key[:1], types.TotalLiquidStakedTokensKey):
			var tokensA, tokensB sdk.IntProto

			cdc.MustUnmarshal(kvA.Value, &tokensA)
			cdc.MustUnmarshal(kvB.Value, &tokensB)

			return fmt.Sprintf("%v\n%v", tokensA, tokensB)
		case bytes.Equal(kvA.Key[:1], types.TokenizeSharesLockPrefix):
			var lockA, lockB types.TokenizeShareLock

			cdc.MustUnmarshal(kvA.Value, &lockA)
			cdc.MustUnmarshal(kvB.Value, &lockB)

			return fmt.Sprintf("%v\n%v", lockA, lockB)
		case bytes.Equal(kvA.Key[:1], types.TokenizeSharesUnlockQueuePrefix):
			var authsA, authsB types.PendingTokenizeShareAuthorizations

			cdc.MustUnmarshal(kvA.Value, &authsA)
			cdc.MustUnmarshal(kvB.Value, &authsB)

			return fmt.Sprintf("%v\n%v", authsA, authsB)
		default:
			panic(fmt.Sprintf("invalid lsm key prefix %X", kvA.Key[:1]))
		}
	}
}
