package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		name      string
		configure func(*types.GenesisState)
		errString string
	}{
		{
			"InvalidParams",
			func(genState *types.GenesisState) {
				params := types.DefaultParams()
				params.SwapFeeRate = sdk.NewDec(-1)
				genState.Params = params
			},
			"swap fee rate must not be negative: -1.000000000000000000",
		},
		{
			"InvalidPoolRecords",
			func(genState *types.GenesisState) {
				genState.PoolRecords = []types.PoolRecord{{}}
			},
			"bad msg index of the batch",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genState := types.DefaultGenesisState()
			tc.configure(genState)
			err := types.ValidateGenesis(*genState)
			require.EqualError(t, err, tc.errString)
		})
	}
}

func TestPoolRecord_Validate(t *testing.T) {
	testCases := []struct {
		name       string
		poolRecord types.PoolRecord
		shouldFail bool
	}{
		{
			"ValidPoolRecord",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  1,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     1,
				},
				DepositMsgStates:  nil,
				WithdrawMsgStates: nil,
				SwapMsgStates:     nil,
			},
			false,
		},
		{
			"InvalidPoolBatchDepositMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  0,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     1,
				},
			},
			true,
		},
		{
			"InvalidPoolBatchWithdrawMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  0,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     0,
				},
			},
			true,
		},
		{
			"InvalidPoolBatchSwapMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  1,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     0,
				},
			},
			true,
		},
		{
			"MismatchingPoolBatchDepositMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  10,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     1,
				},
				DepositMsgStates:  []types.DepositMsgState{{MsgIndex: 1}},
				WithdrawMsgStates: nil,
				SwapMsgStates:     nil,
			},
			true,
		},
		{
			"MismatchingPoolBatchWithdrawMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  1,
					WithdrawMsgIndex: 10,
					SwapMsgIndex:     1,
				},
				DepositMsgStates:  nil,
				WithdrawMsgStates: []types.WithdrawMsgState{{MsgIndex: 1}},
				SwapMsgStates:     nil,
			},
			true,
		},
		{
			"MismatchingPoolBatchSwapMsgIndex",
			types.PoolRecord{
				PoolBatch: types.PoolBatch{
					DepositMsgIndex:  1,
					WithdrawMsgIndex: 1,
					SwapMsgIndex:     10,
				},
				DepositMsgStates:  nil,
				WithdrawMsgStates: nil,
				SwapMsgStates:     []types.SwapMsgState{{MsgIndex: 1}},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.poolRecord.Validate()
			if tc.shouldFail {
				require.ErrorIs(t, err, types.ErrBadBatchMsgIndex)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
