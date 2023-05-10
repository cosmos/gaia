package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding liquidity type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.PoolKeyPrefix):
			var poolA, poolB types.Pool
			cdc.MustUnmarshal(kvA.Value, &poolA)
			cdc.MustUnmarshal(kvB.Value, &poolB)
			return fmt.Sprintf("%v\n%v", poolA, poolB)

		case bytes.Equal(kvA.Key[:1], types.PoolByReserveAccIndexKeyPrefix):
			return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.PoolBatchKeyPrefix):
			var batchA, batchB types.PoolBatch
			cdc.MustUnmarshal(kvA.Value, &batchA)
			cdc.MustUnmarshal(kvB.Value, &batchB)
			return fmt.Sprintf("%v\n%v", batchA, batchB)

		case bytes.Equal(kvA.Key[:1], types.PoolBatchDepositMsgStateIndexKeyPrefix):
			var msgStateA, msgStateB types.DepositMsgState
			cdc.MustUnmarshal(kvA.Value, &msgStateA)
			cdc.MustUnmarshal(kvB.Value, &msgStateB)
			return fmt.Sprintf("%v\n%v", msgStateA, msgStateB)

		case bytes.Equal(kvA.Key[:1], types.PoolBatchWithdrawMsgStateIndexKeyPrefix):
			var msgStateA, msgStateB types.WithdrawMsgState
			cdc.MustUnmarshal(kvA.Value, &msgStateA)
			cdc.MustUnmarshal(kvB.Value, &msgStateB)
			return fmt.Sprintf("%v\n%v", msgStateA, msgStateB)

		case bytes.Equal(kvA.Key[:1], types.PoolBatchSwapMsgStateIndexKeyPrefix):
			var msgStateA, msgStateB types.SwapMsgState
			cdc.MustUnmarshal(kvA.Value, &msgStateA)
			cdc.MustUnmarshal(kvB.Value, &msgStateB)
			return fmt.Sprintf("%v\n%v", msgStateA, msgStateB)

		default:
			panic(fmt.Sprintf("invalid liquidity key prefix %X", kvA.Key[:1]))
		}
	}
}
