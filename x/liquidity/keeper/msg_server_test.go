package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// Although written in msg_server_test.go, it is approached at the keeper level rather than at the msgServer level
// so is not included in the coverage.

func TestMsgCreatePool(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	require.Equal(t, deposit, depositBalance)

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	require.Equal(t, 1, len(pools))
	require.Equal(t, uint64(1), pools[0].Id)
	require.Equal(t, uint64(1), simapp.LiquidityKeeper.GetNextPoolID(ctx)-1)
	require.Equal(t, denomA, pools[0].ReserveCoinDenoms[0])
	require.Equal(t, denomB, pools[0].ReserveCoinDenoms[1])

	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pools[0])
	creatorBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pools[0].PoolCoinDenom)
	require.Equal(t, poolCoin, creatorBalance.Amount)

	_, err = simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.ErrorIs(t, err, types.ErrPoolAlreadyExists)
}

func TestMsgDepositWithinBatch(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 4, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)
	app.SaveAccount(simapp, ctx, addrs[1], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	require.Equal(t, deposit, depositBalance)

	depositA = simapp.BankKeeper.GetBalance(ctx, addrs[1], denomA)
	depositB = simapp.BankKeeper.GetBalance(ctx, addrs[1], denomB)
	depositBalance = sdk.NewCoins(depositA, depositB)

	require.Equal(t, deposit, depositBalance)

	createMsg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, createMsg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)

	depositMsg := types.NewMsgDepositWithinBatch(addrs[1], pool.Id, deposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	depositorBalance := simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom)
	require.Equal(t, poolCoin.Sub(poolCoinBefore), depositorBalance.Amount)
}

func TestMsgWithdrawWithinBatch(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	require.Equal(t, deposit, depositBalance)

	createMsg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, createMsg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	fmt.Println(poolCoinBefore, withdrawerPoolCoinBefore.Amount)
	require.Equal(t, poolCoinBefore, withdrawerPoolCoinBefore.Amount)
	withdrawMsg := types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, poolCoinBefore))

	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteWithdrawal(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinAfter := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)
	require.True(t, true, poolCoinAfter.IsZero())
	require.True(t, true, withdrawerPoolCoinAfter.IsZero())
	withdrawerDenomABalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[0]), withdrawerDenomABalance.Amount)
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[1]), withdrawerDenomBBalance.Amount)
}

func TestMsgGetLiquidityPoolMetadata(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	require.Equal(t, deposit, depositBalance)

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	require.Equal(t, 1, len(pools))
	require.Equal(t, uint64(1), pools[0].Id)
	require.Equal(t, uint64(1), simapp.LiquidityKeeper.GetNextPoolID(ctx)-1)
	require.Equal(t, denomA, pools[0].ReserveCoinDenoms[0])
	require.Equal(t, denomB, pools[0].ReserveCoinDenoms[1])

	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pools[0])
	creatorBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pools[0].PoolCoinDenom)
	require.Equal(t, poolCoin, creatorBalance.Amount)

	_, err = simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.ErrorIs(t, err, types.ErrPoolAlreadyExists)

	metaData := simapp.LiquidityKeeper.GetPoolMetaData(ctx, pools[0])
	require.Equal(t, pools[0].Id, metaData.PoolId)

	reserveCoin := simapp.LiquidityKeeper.GetReserveCoins(ctx, pools[0])
	require.Equal(t, reserveCoin, metaData.ReserveCoins)
	require.Equal(t, msg.DepositCoins, metaData.ReserveCoins)

	totalSupply := sdk.NewCoin(pools[0].PoolCoinDenom, simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pools[0]))
	require.Equal(t, totalSupply, metaData.PoolCoinTotalSupply)
	require.Equal(t, creatorBalance, metaData.PoolCoinTotalSupply)
}

func TestMsgSwapWithinBatch(t *testing.T) {
	simapp, ctx := app.CreateTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)

	depositCoins := sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1_000_000_000)), sdk.NewCoin(DenomY, sdk.NewInt(1_000_000_000)))
	depositorAddr := app.AddRandomTestAddr(simapp, ctx, depositCoins.Add(params.PoolCreationFee...))
	user := app.AddRandomTestAddr(simapp, ctx, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1_000_000_000)), sdk.NewCoin(DenomY, sdk.NewInt(1_000_000_000))))

	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, &types.MsgCreatePool{
		PoolCreatorAddress: depositorAddr.String(),
		PoolTypeId:         types.DefaultPoolTypeID,
		DepositCoins:       depositCoins,
	})
	require.NoError(t, err)

	cases := []struct {
		expectedErr  string // empty means no error expected
		swapFeeRate  sdk.Dec
		msg          *types.MsgSwapWithinBatch
		afterBalance sdk.Coins
	}{
		{
			"",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10000)),
				OfferCoinFee:         types.GetOfferCoinFee(sdk.NewCoin(DenomX, sdk.NewInt(10000)), params.SwapFeeRate),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("999989985denomX,1000009984denomY"),
		},
		{
			// bad offer coin denom
			"bad offer coin fee",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10000)),
				OfferCoinFee:         sdk.NewCoin(DenomY, sdk.NewInt(15)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"bad offer coin fee",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10000)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(14)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"bad offer coin fee",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10000)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(16)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10001)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(16)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("999989983denomX,1000009984denomY"),
		},
		{
			"",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(100)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(1)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("999999899denomX,1000000098denomY"),
		},
		{
			"bad offer coin fee",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(100)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(0)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(1000)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(2)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("999998998denomX,1000000997denomY"),
		},
		{
			"bad offer coin fee",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(1000)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.NewInt(1)),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"",
			sdk.ZeroDec(),
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(1000)),
				OfferCoinFee:         sdk.NewCoin(DenomX, sdk.ZeroInt()),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("999999000denomX,1000000999denomY"),
		},
		{
			"does not match the reserve coin of the pool",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(10000)),
				OfferCoinFee:         types.GetOfferCoinFee(sdk.NewCoin(DenomX, sdk.NewInt(10000)), params.SwapFeeRate),
				DemandCoinDenom:      DenomA,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
		{
			"can not exceed max order ratio of reserve coins that can be ordered at a order",
			types.DefaultSwapFeeRate,
			&types.MsgSwapWithinBatch{
				SwapRequesterAddress: user.String(),
				PoolId:               pool.Id,
				SwapTypeId:           pool.TypeId,
				OfferCoin:            sdk.NewCoin(DenomX, sdk.NewInt(100_000_001)),
				OfferCoinFee:         types.GetOfferCoinFee(sdk.NewCoin(DenomX, sdk.NewInt(100_000_001)), params.SwapFeeRate),
				DemandCoinDenom:      DenomY,
				OrderPrice:           sdk.MustNewDecFromStr("1.00002"),
			},
			types.MustParseCoinsNormalized("1000000000denomX,1000000000denomY"),
		},
	}

	for _, tc := range cases {
		cacheCtx, _ := ctx.CacheContext()
		cacheCtx = cacheCtx.WithBlockHeight(1)
		params.SwapFeeRate = tc.swapFeeRate
		simapp.LiquidityKeeper.SetParams(cacheCtx, params)
		_, err = simapp.LiquidityKeeper.SwapWithinBatch(cacheCtx, tc.msg, types.CancelOrderLifeSpan)
		if tc.expectedErr == "" {
			require.NoError(t, err)
			liquidity.EndBlocker(cacheCtx, simapp.LiquidityKeeper)
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
		moduleAccAddress := simapp.AccountKeeper.GetModuleAddress(types.ModuleName)
		require.True(t, simapp.BankKeeper.GetAllBalances(cacheCtx, moduleAccAddress).IsZero())
		require.Equal(t, tc.afterBalance, simapp.BankKeeper.GetAllBalances(cacheCtx, user))
	}
}
