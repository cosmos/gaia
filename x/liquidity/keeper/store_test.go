package keeper_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestGetAllLiquidityPoolBatchSwapMsgs(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		r := rand.New(rand.NewSource(seed))

		simapp, ctx := createTestInput()
		simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
		params := simapp.LiquidityKeeper.GetParams(ctx)

		// define test denom X, Y for Liquidity Pool
		denomX := "denomX" //nolint:goconst // for testing
		denomY := "denomY" //nolint:goconst // for testing
		denomX, denomY = types.AlphabeticalDenomPair(denomX, denomY)

		// get random X, Y amount for create pool
		X, Y := app.GetRandPoolAmt(r, params.MinInitDepositAmount)
		deposit := sdk.NewCoins(sdk.NewCoin(denomX, X), sdk.NewCoin(denomY, Y))

		// set pool creator account, balance for deposit
		addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)
		app.SaveAccount(simapp, ctx, addrs[0], deposit) // pool creator
		depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomX)
		depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomY)
		depositBalance := sdk.NewCoins(depositA, depositB)
		require.Equal(t, deposit, depositBalance)

		// create Liquidity pool
		poolTypeID := types.DefaultPoolTypeID
		msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)
		_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
		require.NoError(t, err)

		var xToY []*types.MsgSwapWithinBatch // buying Y from X
		var yToX []*types.MsgSwapWithinBatch // selling Y for X

		// make random orders, set buyer, seller accounts for the orders
		xToY, yToX = app.GetRandomOrders(denomX, denomY, X, Y, r, 11, 11)
		buyerAddrs := app.AddTestAddrsIncremental(simapp, ctx, len(xToY), sdk.ZeroInt())
		sellerAddrs := app.AddTestAddrsIncremental(simapp, ctx, len(yToX), sdk.ZeroInt())

		poolID := uint64(1)
		pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
		require.True(t, found)

		poolBatch, _ := simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
		require.Equal(t, uint64(1), poolBatch.SwapMsgIndex)

		for i, msg := range xToY {
			app.SaveAccountWithFee(simapp, ctx, buyerAddrs[i], sdk.NewCoins(msg.OfferCoin), msg.OfferCoin)
			msg.SwapRequesterAddress = buyerAddrs[i].String()
			msg.PoolId = pool.Id
			msg.OfferCoinFee = types.GetOfferCoinFee(msg.OfferCoin, params.SwapFeeRate)
		}
		for i, msg := range yToX {
			app.SaveAccountWithFee(simapp, ctx, sellerAddrs[i], sdk.NewCoins(msg.OfferCoin), msg.OfferCoin)
			msg.SwapRequesterAddress = sellerAddrs[i].String()
			msg.PoolId = pool.Id
			msg.OfferCoinFee = types.GetOfferCoinFee(msg.OfferCoin, params.SwapFeeRate)
		}

		// handle msgs, set order msgs to batch
		for _, msg := range xToY[:10] {
			_, err := simapp.LiquidityKeeper.SwapWithinBatch(ctx, msg, 0)
			require.NoError(t, err)
		}
		for _, msg := range yToX[:10] {
			_, err := simapp.LiquidityKeeper.SwapWithinBatch(ctx, msg, 0)
			require.NoError(t, err)
		}

		msgs := simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, poolBatch)
		require.Equal(t, 20, len(msgs))

		simapp.LiquidityKeeper.IterateAllPoolBatchSwapMsgStates(ctx, poolBatch, func(msg types.SwapMsgState) bool {
			if msg.MsgIndex%2 == 1 {
				simapp.LiquidityKeeper.DeletePoolBatchSwapMsgState(ctx, msg.Msg.PoolId, msg.MsgIndex)
			}
			return false
		})

		msgs = simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, poolBatch)
		require.Equal(t, 10, len(msgs))

		poolBatch, _ = simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
		require.Equal(t, uint64(21), poolBatch.SwapMsgIndex)

		poolBatch.SwapMsgIndex = uint64(18446744073709551610)
		simapp.LiquidityKeeper.SetPoolBatch(ctx, poolBatch)

		_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, xToY[10], 0)
		require.NoError(t, err)
		_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, yToX[10], 0)
		require.NoError(t, err)

		msgs = simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, poolBatch)
		require.Equal(t, 12, len(msgs))
		require.Equal(t, xToY[10], msgs[10].Msg)
		require.Equal(t, yToX[10], msgs[11].Msg)
	}
}

func TestGetAllNotProcessedPoolBatchSwapMsgs(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)

	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))
	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])

	// begin block, init
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, true)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, true)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	price, _ := sdk.NewDecFromStr("1.1")
	offerCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000)), sdk.NewCoin(denomX, sdk.NewInt(10000)), sdk.NewCoin(denomX, sdk.NewInt(10000))}
	orderPrices := []sdk.Dec{price, price, price}
	orderAddrs := addrs[1:4]
	batchMsgs, _ := app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, false)
	batchMsgs2, batch := app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, false)
	require.Equal(t, 3, len(batchMsgs))
	for _, msg := range batchMsgs2 {
		msg.Executed = true
		msg.Succeeded = true
		msg.ToBeDeleted = true
	}
	require.Equal(t, 3, len(batchMsgs2))
	simapp.LiquidityKeeper.SetPoolBatchSwapMsgStatesByPointer(ctx, poolID, batchMsgs2)

	resultMsgs := simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, batch)
	resultProcessedMsgs := simapp.LiquidityKeeper.GetAllNotProcessedPoolBatchSwapMsgStates(ctx, batch)
	require.Equal(t, 6, len(resultMsgs))
	require.Equal(t, 3, len(resultProcessedMsgs))
}

func TestIterateAllBatchMsgs(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)
	denomA, denomB := types.AlphabeticalDenomPair("denomA", "denomB")

	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(500000000)
	A := sdk.NewInt(500000000)
	B := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))
	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])
	poolID2 := app.TestCreatePool(t, simapp, ctx, A, B, denomA, denomB, addrs[4])
	batch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
	require.True(t, found)

	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)

	price, _ := sdk.NewDecFromStr("1.1")
	priceY, _ := sdk.NewDecFromStr("1.2")
	xOfferCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000))}
	yOfferCoins := []sdk.Coin{sdk.NewCoin(denomY, sdk.NewInt(5000))}

	xOrderPrices := []sdk.Dec{price}
	yOrderPrices := []sdk.Dec{priceY}
	xOrderAddrs := addrs[1:2]
	yOrderAddrs := addrs[2:3]

	offerCoins2 := []sdk.Coin{sdk.NewCoin(denomA, sdk.NewInt(5000))}

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	app.TestDepositPool(t, simapp, ctx, A, B.QuoRaw(10), addrs[4:5], poolID2, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(1000), addrs[4:5], poolID2, false)
	app.TestSwapPool(t, simapp, ctx, offerCoins2, xOrderPrices, addrs[4:5], poolID2, true)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	app.TestDepositPool(t, simapp, ctx, A, B.QuoRaw(10), addrs[4:5], poolID2, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(1000), addrs[4:5], poolID2, false)
	app.TestSwapPool(t, simapp, ctx, offerCoins2, xOrderPrices, addrs[4:5], poolID2, true)

	// next block,
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// Reinitialize batch messages that were not executed in the previous batch and delete batch messages that were executed or ready to delete.
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(50), addrs[1:2], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(500), addrs[1:2], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(50), addrs[2:3], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(500), addrs[2:3], poolID, false)

	depositMsgsRemaining := simapp.LiquidityKeeper.GetAllRemainingPoolBatchDepositMsgStates(ctx, batch)
	require.Equal(t, 0, len(depositMsgsRemaining))

	var depositMsgs []types.DepositMsgState
	simapp.LiquidityKeeper.IterateAllDepositMsgStates(ctx, func(msg types.DepositMsgState) bool {
		depositMsgs = append(depositMsgs, msg)
		return false
	})
	require.Equal(t, 4, len(depositMsgs))

	depositMsgs[0].ToBeDeleted = true
	simapp.LiquidityKeeper.SetPoolBatchDepositMsgStates(ctx, poolID, []types.DepositMsgState{depositMsgs[0]})
	depositMsgsNotToDelete := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgStatesNotToBeDeleted(ctx, batch)
	require.Equal(t, 3, len(depositMsgsNotToDelete))

	var withdrawMsgs []types.WithdrawMsgState
	simapp.LiquidityKeeper.IterateAllWithdrawMsgStates(ctx, func(msg types.WithdrawMsgState) bool {
		withdrawMsgs = append(withdrawMsgs, msg)
		return false
	})
	withdrawMsgs[0].ToBeDeleted = true
	simapp.LiquidityKeeper.SetPoolBatchWithdrawMsgStates(ctx, poolID, withdrawMsgs[0:1])

	withdrawMsgsNotToDelete := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStatesNotToBeDeleted(ctx, batch)
	require.Equal(t, 4, len(withdrawMsgs))
	require.Equal(t, 3, len(withdrawMsgsNotToDelete))
	require.NotEqual(t, withdrawMsgsNotToDelete, withdrawMsgs)

	app.TestDepositPool(t, simapp, ctx, A, B.QuoRaw(10), addrs[4:5], poolID2, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(1000), addrs[4:5], poolID2, false)

	depositMsgs = simapp.LiquidityKeeper.GetAllDepositMsgStates(ctx)
	require.Equal(t, 5, len(depositMsgs))
	withdrawMsgs = simapp.LiquidityKeeper.GetAllWithdrawMsgStates(ctx)
	require.Equal(t, 5, len(depositMsgs))

	var depositMsgs2 []types.DepositMsgState
	simapp.LiquidityKeeper.IterateAllDepositMsgStates(ctx, func(msg types.DepositMsgState) bool {
		depositMsgs2 = append(depositMsgs2, msg)
		return false
	})

	var withdrawMsgs2 []types.WithdrawMsgState
	simapp.LiquidityKeeper.IterateAllWithdrawMsgStates(ctx, func(msg types.WithdrawMsgState) bool {
		withdrawMsgs2 = append(withdrawMsgs2, msg)
		return false
	})

	require.Equal(t, 5, len(depositMsgs2))

	require.Equal(t, 5, len(withdrawMsgs2))

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	depositMsgsRemaining = simapp.LiquidityKeeper.GetAllRemainingPoolBatchDepositMsgStates(ctx, batch)
	require.Equal(t, 0, len(depositMsgsRemaining))

	// next block,
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// Reinitialize batch messages that were not executed in the previous batch and delete batch messages that were executed or ready to delete.
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	var depositMsgs3 []types.DepositMsgState
	simapp.LiquidityKeeper.IterateAllDepositMsgStates(ctx, func(msg types.DepositMsgState) bool {
		depositMsgs3 = append(depositMsgs3, msg)
		return false
	})
	require.Equal(t, 0, len(depositMsgs3))

	var withdrawMsgs3 []types.WithdrawMsgState
	simapp.LiquidityKeeper.IterateAllWithdrawMsgStates(ctx, func(msg types.WithdrawMsgState) bool {
		withdrawMsgs3 = append(withdrawMsgs3, msg)
		return false
	})
	require.Equal(t, 0, len(withdrawMsgs3))

	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, yOfferCoins, yOrderPrices, yOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, offerCoins2, xOrderPrices, addrs[4:5], poolID2, false)

	swapMsgsPool1 := simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStates(ctx, batch)
	require.Equal(t, 4, len(swapMsgsPool1))

	swapMsg, found := simapp.LiquidityKeeper.GetPoolBatchSwapMsgState(ctx, batch.PoolId, 1)
	require.True(t, found)
	require.Equal(t, swapMsg, swapMsgsPool1[0])

	var swapMsgsAllPool []types.SwapMsgState
	simapp.LiquidityKeeper.IterateAllSwapMsgStates(ctx, func(msg types.SwapMsgState) bool {
		swapMsgsAllPool = append(swapMsgsAllPool, msg)
		return false
	})
	require.Equal(t, 5, len(swapMsgsAllPool))

	swapMsgsAllPool = simapp.LiquidityKeeper.GetAllSwapMsgStates(ctx)
	require.Equal(t, 5, len(swapMsgsAllPool))
	require.Equal(t, swapMsgsPool1, swapMsgsAllPool[:len(swapMsgsPool1)])

	swapMsgsAllPool[1].Executed = true
	simapp.LiquidityKeeper.SetPoolBatchSwapMsgStates(ctx, poolID, swapMsgsAllPool[1:2])

	remainingSwapMsgs := simapp.LiquidityKeeper.GetAllRemainingPoolBatchSwapMsgStates(ctx, batch)
	require.Equal(t, 1, len(remainingSwapMsgs))

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	// next block,
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// Reinitialize batch messages that were not executed in the previous batch and delete batch messages that were executed or ready to delete.
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	var swapMsg2 []types.SwapMsgState
	simapp.LiquidityKeeper.IterateAllSwapMsgStates(ctx, func(msg types.SwapMsgState) bool {
		swapMsg2 = append(swapMsg2, msg)
		return false
	})
	require.Equal(t, 0, len(swapMsg2))

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	genesis := simapp.LiquidityKeeper.ExportGenesis(ctx)
	simapp.LiquidityKeeper.InitGenesis(ctx, *genesis)
	genesisNew := simapp.LiquidityKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis, genesisNew)

	simapp.LiquidityKeeper.DeletePoolBatch(ctx, batch)
	batch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, batch.PoolId)
	require.Equal(t, types.PoolBatch{}, batch)
	require.False(t, found)
}
