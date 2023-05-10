package types_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestParams(t *testing.T) {
	require.IsType(t, paramstypes.KeyTable{}, types.ParamKeyTable())

	simapp, ctx := app.CreateTestInput()
	defaultParams := types.DefaultParams()
	require.Equal(t, defaultParams, simapp.LiquidityKeeper.GetParams(ctx))

	paramsStr := `pool_types:
- id: 1
  name: StandardLiquidityPool
  min_reserve_coin_num: 2
  max_reserve_coin_num: 2
  description: Standard liquidity pool with pool price function X/Y, ESPM constraint,
    and two kinds of reserve coins
min_init_deposit_amount: "1000000"
init_pool_coin_mint_amount: "1000000"
max_reserve_coin_amount: "0"
pool_creation_fee:
- denom: stake
  amount: "40000000"
swap_fee_rate: "0.003000000000000000"
withdraw_fee_rate: "0.000000000000000000"
max_order_amount_ratio: "0.100000000000000000"
unit_batch_height: 1
circuit_breaker_enabled: false
`
	require.Equal(t, paramsStr, defaultParams.String())
}

func TestParams_Validate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	testCases := []struct {
		name      string
		configure func(*types.Params)
		errString string
	}{
		{
			"EmptyPoolTypes",
			func(params *types.Params) {
				params.PoolTypes = []types.PoolType{}
			},
			"pool types must not be empty",
		},
		{
			"TooManyPoolTypes",
			func(params *types.Params) {
				params.PoolTypes = []types.PoolType{types.DefaultPoolType, types.DefaultPoolType}
			},
			"pool type ids must be sorted",
		},
		{
			"CustomPoolType",
			func(params *types.Params) {
				poolType := types.DefaultPoolType
				poolType.Name = "CustomPoolType"
				params.PoolTypes = []types.PoolType{poolType}
			},
			"the only supported pool type is 1",
		},
		{
			"NilMinInitDepositAmount",
			func(params *types.Params) {
				params.MinInitDepositAmount = sdk.Int{}
			},
			"minimum initial deposit amount must not be nil",
		},
		{
			"NonPositiveMinInitDepositAmount",
			func(params *types.Params) {
				params.MinInitDepositAmount = sdk.NewInt(0)
			},
			"minimum initial deposit amount must be positive: 0",
		},
		{
			"NilInitPoolCoinMintAmount",
			func(params *types.Params) {
				params.InitPoolCoinMintAmount = sdk.Int{}
			},
			"initial pool coin mint amount must not be nil",
		},
		{
			"NonPositiveInitPoolCoinMintAmount",
			func(params *types.Params) {
				params.InitPoolCoinMintAmount = sdk.ZeroInt()
			},
			"initial pool coin mint amount must be positive: 0",
		},
		{
			"TooSmallInitPoolCoinMintAmount",
			func(params *types.Params) {
				params.InitPoolCoinMintAmount = sdk.NewInt(10)
			},
			"initial pool coin mint amount must be greater than or equal to 1000000: 10",
		},
		{
			"NilMaxReserveCoinAmount",
			func(params *types.Params) {
				params.MaxReserveCoinAmount = sdk.Int{}
			},
			"max reserve coin amount must not be nil",
		},
		{
			"NegativeMaxReserveCoinAmount",
			func(params *types.Params) {
				params.MaxReserveCoinAmount = sdk.NewInt(-1)
			},
			"max reserve coin amount must not be negative: -1",
		},
		{
			"NilSwapFeeRate",
			func(params *types.Params) {
				params.SwapFeeRate = sdk.Dec{}
			},
			"swap fee rate must not be nil",
		},
		{
			"NegativeSwapFeeRate",
			func(params *types.Params) {
				params.SwapFeeRate = sdk.NewDec(-1)
			},
			"swap fee rate must not be negative: -1.000000000000000000",
		},
		{
			"TooLargeSwapFeeRate",
			func(params *types.Params) {
				params.SwapFeeRate = sdk.NewDec(2)
			},
			"swap fee rate too large: 2.000000000000000000",
		},
		{
			"NilWithdrawFeeRate",
			func(params *types.Params) {
				params.WithdrawFeeRate = sdk.Dec{}
			},
			"withdraw fee rate must not be nil",
		},
		{
			"NegativeWithdrawFeeRate",
			func(params *types.Params) {
				params.WithdrawFeeRate = sdk.NewDec(-1)
			},
			"withdraw fee rate must not be negative: -1.000000000000000000",
		},
		{
			"TooLargeWithdrawFeeRate",
			func(params *types.Params) {
				params.WithdrawFeeRate = sdk.NewDec(2)
			},
			"withdraw fee rate too large: 2.000000000000000000",
		},
		{
			"NilMaxOrderAmountRatio",
			func(params *types.Params) {
				params.MaxOrderAmountRatio = sdk.Dec{}
			},
			"max order amount ratio must not be nil",
		},
		{
			"NegativeMaxOrderAmountRatio",
			func(params *types.Params) {
				params.MaxOrderAmountRatio = sdk.NewDec(-1)
			},
			"max order amount ratio must not be negative: -1.000000000000000000",
		},
		{
			"TooLargeMaxOrderAmountRatio",
			func(params *types.Params) {
				params.MaxOrderAmountRatio = sdk.NewDec(2)
			},
			"max order amount ratio too large: 2.000000000000000000",
		},
		{
			"EmptyPoolCreationFee",
			func(params *types.Params) {
				params.PoolCreationFee = sdk.NewCoins()
			},
			"pool creation fee must not be empty",
		},
		{
			"InvalidPoolCreationFeeDenom",
			func(params *types.Params) {
				params.PoolCreationFee = sdk.Coins{
					sdk.Coin{
						Denom:  "invalid denom---",
						Amount: params.PoolCreationFee.AmountOf(params.PoolCreationFee.GetDenomByIndex(0)),
					},
				}
			},
			"invalid denom: invalid denom---",
		},
		{
			"NotPositivePoolCreationFeeAmount",
			func(params *types.Params) {
				params.PoolCreationFee = sdk.Coins{
					sdk.Coin{
						Denom:  params.PoolCreationFee.GetDenomByIndex(0),
						Amount: sdk.ZeroInt(),
					},
				}
			},
			"coin 0stake amount is not positive",
		},
		{
			"NonPositiveUnitBatchHeight",
			func(params *types.Params) {
				params.UnitBatchHeight = 0
			},
			"unit batch height must be positive: 0",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.configure(&params)
			err := params.Validate()
			require.EqualError(t, err, tc.errString)
			var err2 error
			for _, p := range params.ParamSetPairs() {
				err := p.ValidatorFn(reflect.ValueOf(p.Value).Elem().Interface())
				if err != nil {
					err2 = err
					break
				}
			}
			require.EqualError(t, err2, tc.errString)
		})
	}
}
