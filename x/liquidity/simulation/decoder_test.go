package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/cosmos/gaia/v9/x/liquidity/simulation"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

var (
	pk1               = ed25519.GenPrivKey().PubKey()
	reserveAccAddr1   = sdk.AccAddress(pk1.Address())
	reserveCoinDenoms = []string{"dzkiv", "imwo"}
	poolName          = types.PoolName(reserveCoinDenoms, uint32(1))
	poolCoinDenom     = types.GetPoolCoinDenom(poolName)
)

func TestDecodeLiquidityStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	pool := types.Pool{
		Id:                    uint64(1),
		TypeId:                uint32(1),
		ReserveCoinDenoms:     reserveCoinDenoms,
		ReserveAccountAddress: reserveAccAddr1.String(),
		PoolCoinDenom:         poolCoinDenom,
	}
	batch := types.NewPoolBatch(1, 1)
	depositMsgState := types.DepositMsgState{
		MsgHeight:   int64(50),
		MsgIndex:    uint64(1),
		Executed:    true,
		Succeeded:   true,
		ToBeDeleted: true,
		Msg:         &types.MsgDepositWithinBatch{PoolId: uint64(1)},
	}
	withdrawMsgState := types.WithdrawMsgState{
		MsgHeight:   int64(50),
		MsgIndex:    uint64(1),
		Executed:    true,
		Succeeded:   true,
		ToBeDeleted: true,
		Msg:         &types.MsgWithdrawWithinBatch{PoolId: uint64(1)},
	}
	swapMsgState := types.SwapMsgState{
		MsgHeight:   int64(50),
		MsgIndex:    uint64(1),
		Executed:    true,
		Succeeded:   true,
		ToBeDeleted: true,
		Msg:         &types.MsgSwapWithinBatch{PoolId: uint64(1)},
	}

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.PoolKeyPrefix, Value: cdc.MustMarshal(&pool)},
			{Key: types.PoolByReserveAccIndexKeyPrefix, Value: reserveAccAddr1.Bytes()},
			{Key: types.PoolBatchKeyPrefix, Value: cdc.MustMarshal(&batch)},
			{Key: types.PoolBatchDepositMsgStateIndexKeyPrefix, Value: cdc.MustMarshal(&depositMsgState)},
			{Key: types.PoolBatchWithdrawMsgStateIndexKeyPrefix, Value: cdc.MustMarshal(&withdrawMsgState)},
			{Key: types.PoolBatchSwapMsgStateIndexKeyPrefix, Value: cdc.MustMarshal(&swapMsgState)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Pool", fmt.Sprintf("%v\n%v", pool, pool)},
		{"PoolByReserveAccIndex", fmt.Sprintf("%v\n%v", reserveAccAddr1, reserveAccAddr1)},
		{"PoolBatchKey", fmt.Sprintf("%v\n%v", batch, batch)},
		{"PoolBatchDepositMsgStateIndex", fmt.Sprintf("%v\n%v", depositMsgState, depositMsgState)},
		{"PoolBatchWithdrawMsgStateIndex", fmt.Sprintf("%v\n%v", withdrawMsgState, withdrawMsgState)},
		{"PoolBatchSwapMsgStateIndex", fmt.Sprintf("%v\n%v", swapMsgState, swapMsgState)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
