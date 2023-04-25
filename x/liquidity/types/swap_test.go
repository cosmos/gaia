package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestSwapScenario(t *testing.T) {
	// init test app and context
	simapp, ctx := app.CreateTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)
	X := params.MinInitDepositAmount
	Y := params.MinInitDepositAmount

	// init addresses for the test
	addrs := app.AddTestAddrs(simapp, ctx, 20, params.PoolCreationFee)

	// Create pool
	// The create pool msg is not run in batch, but is processed immediately.
	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])

	// In case of deposit, withdraw, and swap msg, unlike other normal tx msgs,
	// collect them in the batch and perform an execution at once at the endblock.

	// add a deposit to pool and run batch execution on endblock
	app.TestDepositPool(t, simapp, ctx, X, Y, addrs[1:2], poolID, true)

	// next block, reinitialize batch and increase batchIndex at beginBlocker,
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	// Create swap msg for test purposes and put it in the batch.
	price, _ := sdk.NewDecFromStr("1.1")
	priceY, _ := sdk.NewDecFromStr("1.2")
	xOfferCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000))}
	yOfferCoins := []sdk.Coin{sdk.NewCoin(denomY, sdk.NewInt(5000))}
	xOrderPrices := []sdk.Dec{price}
	yOrderPrices := []sdk.Dec{priceY}
	xOrderAddrs := addrs[1:2]
	yOrderAddrs := addrs[2:3]
	_, batch := app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, yOfferCoins, yOrderPrices, yOrderAddrs, poolID, false)

	// Set the execution status flag of messages to true.
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, batch)
	for _, msg := range msgs {
		msg.Executed = true
	}
	simapp.LiquidityKeeper.SetPoolBatchSwapMsgStatesByPointer(ctx, poolID, msgs)

	// Generate an orderbook by arranging swap messages in order price
	orderMap, xToY, yToX := types.MakeOrderMap(msgs, denomX, denomY, false)
	orderBook := orderMap.SortOrderBook()
	currentPrice := X.Quo(Y).ToDec()
	require.Equal(t, orderMap[xOrderPrices[0].String()].BuyOfferAmt, xOfferCoins[0].Amount.MulRaw(3))
	require.Equal(t, orderMap[xOrderPrices[0].String()].Price, xOrderPrices[0])

	require.Equal(t, 3, len(xToY))
	require.Equal(t, 1, len(yToX))
	require.Equal(t, 3, len(orderMap[xOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 1, len(orderMap[yOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 3, len(orderBook[0].SwapMsgStates))
	require.Equal(t, 1, len(orderBook[1].SwapMsgStates))

	require.Equal(t, len(orderBook), orderBook.Len())

	fmt.Println(orderBook, currentPrice)
	fmt.Println(xToY, yToX)

	types.ValidateStateAndExpireOrders(xToY, ctx.BlockHeight(), false)
	types.ValidateStateAndExpireOrders(yToX, ctx.BlockHeight(), false)

	// The price and coins of swap messages in orderbook are calculated
	// to derive match result with the price direction.
	result, found := orderBook.Match(X.ToDec(), Y.ToDec())
	require.True(t, found)
	require.NotEqual(t, types.NoMatch, result.MatchType)

	matchResultXtoY, poolXDeltaXtoY, poolYDeltaXtoY := types.FindOrderMatch(types.DirectionXtoY, xToY, result.EX,
		result.SwapPrice, ctx.BlockHeight())
	matchResultYtoX, poolXDeltaYtoX, poolYDeltaYtoX := types.FindOrderMatch(types.DirectionYtoX, yToX, result.EY,
		result.SwapPrice, ctx.BlockHeight())

	xToY, yToX, XDec, YDec, poolXDelta2, poolYDelta2 := types.UpdateSwapMsgStates(X.ToDec(), Y.ToDec(), xToY, yToX, matchResultXtoY, matchResultYtoX)

	require.Equal(t, 0, types.CountNotMatchedMsgs(xToY))
	require.Equal(t, 0, types.CountFractionalMatchedMsgs(xToY))
	require.Equal(t, 1, types.CountNotMatchedMsgs(yToX))
	require.Equal(t, 0, types.CountFractionalMatchedMsgs(yToX))
	require.Equal(t, 3, len(xToY))
	require.Equal(t, 1, len(yToX))

	fmt.Println(matchResultXtoY)
	fmt.Println(poolXDeltaXtoY)
	fmt.Println(poolYDeltaXtoY)

	fmt.Println(poolXDeltaYtoX, poolYDeltaYtoX)
	fmt.Println(poolXDelta2, poolYDelta2)
	fmt.Println(XDec, YDec)

	// Verify swap result by creating an orderbook with remaining messages that have been matched and not transacted.
	orderMapExecuted, _, _ := types.MakeOrderMap(append(xToY, yToX...), denomX, denomY, true)
	orderBookExecuted := orderMapExecuted.SortOrderBook()
	lastPrice := XDec.Quo(YDec)
	fmt.Println("lastPrice", lastPrice)
	fmt.Println("X", XDec)
	fmt.Println("Y", YDec)
	require.True(t, orderBookExecuted.Validate(lastPrice))

	require.Equal(t, 0, types.CountNotMatchedMsgs(orderMapExecuted[xOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 1, types.CountNotMatchedMsgs(orderMapExecuted[yOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 1, types.CountNotMatchedMsgs(orderBookExecuted[0].SwapMsgStates))

	types.ValidateStateAndExpireOrders(xToY, ctx.BlockHeight(), true)
	types.ValidateStateAndExpireOrders(yToX, ctx.BlockHeight(), true)

	orderMapCleared, _, _ := types.MakeOrderMap(append(xToY, yToX...), denomX, denomY, true)
	orderBookCleared := orderMapCleared.SortOrderBook()
	require.True(t, orderBookCleared.Validate(lastPrice))

	require.Equal(t, 0, types.CountNotMatchedMsgs(orderMapCleared[xOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 0, types.CountNotMatchedMsgs(orderMapCleared[yOrderPrices[0].String()].SwapMsgStates))
	require.Equal(t, 0, len(orderBookCleared))

	// next block
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	// test genesisState with export, init
	genesis := simapp.LiquidityKeeper.ExportGenesis(ctx)
	simapp.LiquidityKeeper.InitGenesis(ctx, *genesis)
	err := types.ValidateGenesis(*genesis)
	require.NoError(t, err)
	genesisNew := simapp.LiquidityKeeper.ExportGenesis(ctx)
	err = types.ValidateGenesis(*genesisNew)
	require.NoError(t, err)
	require.Equal(t, genesis, genesisNew)
	for _, record := range genesisNew.PoolRecords {
		err = record.Validate()
		require.NoError(t, err)
	}

	// validate genesis fail case
	batch.DepositMsgIndex = 0
	simapp.LiquidityKeeper.SetPoolBatch(ctx, batch)
	genesisNew = simapp.LiquidityKeeper.ExportGenesis(ctx)
	err = types.ValidateGenesis(*genesisNew)
	require.ErrorIs(t, err, types.ErrBadBatchMsgIndex)
	batch.WithdrawMsgIndex = 0
	simapp.LiquidityKeeper.SetPoolBatch(ctx, batch)
	genesisNew = simapp.LiquidityKeeper.ExportGenesis(ctx)
	err = types.ValidateGenesis(*genesisNew)
	require.ErrorIs(t, err, types.ErrBadBatchMsgIndex)
	batch.SwapMsgIndex = 20
	simapp.LiquidityKeeper.SetPoolBatch(ctx, batch)
	genesisNew = simapp.LiquidityKeeper.ExportGenesis(ctx)
	err = types.ValidateGenesis(*genesisNew)
	require.ErrorIs(t, err, types.ErrBadBatchMsgIndex)
}

func TestMaxOrderRatio(t *testing.T) {
	simapp, ctx := app.CreateTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)

	X := params.MinInitDepositAmount
	Y := params.MinInitDepositAmount

	addrs := app.AddTestAddrs(simapp, ctx, 20, params.PoolCreationFee)
	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])

	// begin block, init
	app.TestDepositPool(t, simapp, ctx, X, Y, addrs[1:2], poolID, true)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	maxOrderRatio := params.MaxOrderAmountRatio

	// Success case, not exceed GetMaxOrderRatio orders
	priceBuy, _ := sdk.NewDecFromStr("1.1")
	priceSell, _ := sdk.NewDecFromStr("1.2")

	offerCoin := sdk.NewCoin(denomX, sdk.NewInt(1000))
	offerCoinY := sdk.NewCoin(denomY, sdk.NewInt(1000))

	app.SaveAccountWithFee(simapp, ctx, addrs[1], sdk.NewCoins(offerCoin), offerCoin)
	app.SaveAccountWithFee(simapp, ctx, addrs[2], sdk.NewCoins(offerCoinY), offerCoinY)

	msgBuy := types.NewMsgSwapWithinBatch(addrs[1], poolID, DefaultSwapTypeId, offerCoin, DenomY, priceBuy, params.SwapFeeRate)
	msgSell := types.NewMsgSwapWithinBatch(addrs[2], poolID, DefaultSwapTypeId, offerCoinY, DenomX, priceSell, params.SwapFeeRate)

	_, err := simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgBuy, 0)
	require.NoError(t, err)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgSell, 0)
	require.NoError(t, err)

	// Fail case, exceed GetMaxOrderRatio orders
	offerCoin = sdk.NewCoin(denomX, X)
	offerCoinY = sdk.NewCoin(denomY, Y)

	app.SaveAccountWithFee(simapp, ctx, addrs[1], sdk.NewCoins(offerCoin), offerCoin)
	app.SaveAccountWithFee(simapp, ctx, addrs[2], sdk.NewCoins(offerCoinY), offerCoinY)

	msgBuy = types.NewMsgSwapWithinBatch(addrs[1], poolID, DefaultSwapTypeId, offerCoin, DenomY, priceBuy, params.SwapFeeRate)
	msgSell = types.NewMsgSwapWithinBatch(addrs[2], poolID, DefaultSwapTypeId, offerCoinY, DenomX, priceSell, params.SwapFeeRate)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgBuy, 0)
	require.Equal(t, types.ErrExceededMaxOrderable, err)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgSell, 0)
	require.Equal(t, types.ErrExceededMaxOrderable, err)

	// Success case, same GetMaxOrderRatio orders
	offerCoin = sdk.NewCoin(denomX, X.ToDec().Mul(maxOrderRatio).TruncateInt())
	offerCoinY = sdk.NewCoin(denomY, Y.ToDec().Mul(maxOrderRatio).TruncateInt())

	app.SaveAccountWithFee(simapp, ctx, addrs[1], sdk.NewCoins(offerCoin), offerCoin)
	app.SaveAccountWithFee(simapp, ctx, addrs[2], sdk.NewCoins(offerCoinY), offerCoinY)

	msgBuy = types.NewMsgSwapWithinBatch(addrs[1], poolID, DefaultSwapTypeId, offerCoin, DenomY, priceBuy, params.SwapFeeRate)
	msgSell = types.NewMsgSwapWithinBatch(addrs[2], poolID, DefaultSwapTypeId, offerCoinY, DenomX, priceSell, params.SwapFeeRate)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgBuy, 0)
	require.NoError(t, err)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgSell, 0)
	require.NoError(t, err)

	// Success case, same GetMaxOrderRatio orders
	_ = sdk.NewCoin(denomX, X.ToDec().Mul(maxOrderRatio).TruncateInt().AddRaw(1))
	_ = sdk.NewCoin(denomY, Y.ToDec().Mul(maxOrderRatio).TruncateInt().AddRaw(1))

	offerCoin = sdk.NewCoin(denomX, params.MinInitDepositAmount.Quo(sdk.NewInt(2)))
	offerCoinY = sdk.NewCoin(denomY, params.MinInitDepositAmount.Quo(sdk.NewInt(10)))
	app.SaveAccountWithFee(simapp, ctx, addrs[1], sdk.NewCoins(offerCoin), offerCoin)
	app.SaveAccountWithFee(simapp, ctx, addrs[2], sdk.NewCoins(offerCoinY), offerCoinY)

	msgBuy = types.NewMsgSwapWithinBatch(addrs[1], poolID, DefaultSwapTypeId, offerCoin, DenomY, priceBuy, params.SwapFeeRate)
	msgSell = types.NewMsgSwapWithinBatch(addrs[2], poolID, DefaultSwapTypeId, offerCoinY, DenomX, priceSell, params.SwapFeeRate)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgBuy, 0)
	require.Equal(t, types.ErrExceededMaxOrderable, err)

	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, msgSell, 0)
	require.NoError(t, err)
}

func TestOrderBookSort(t *testing.T) {
	orderMap := make(types.OrderMap)
	a, _ := sdk.NewDecFromStr("0.1")
	b, _ := sdk.NewDecFromStr("0.2")
	c, _ := sdk.NewDecFromStr("0.3")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.ZeroInt(),
	}
	// make orderbook to sort orderMap
	orderBook := orderMap.SortOrderBook()
	fmt.Println(orderBook)

	res := orderBook.Less(0, 1)
	require.True(t, res)
	res = orderBook.Less(1, 2)
	require.True(t, res)
	res = orderBook.Less(2, 1)
	require.False(t, res)

	orderBook.Swap(1, 2)
	fmt.Println(orderBook)
	require.Equal(t, c, orderBook[1].Price)
	require.Equal(t, b, orderBook[2].Price)

	orderBook.Sort()
	fmt.Println(orderBook)
	require.Equal(t, a, orderBook[0].Price)
	require.Equal(t, b, orderBook[1].Price)
	require.Equal(t, c, orderBook[2].Price)

	orderBook.Reverse()
	fmt.Println(orderBook)
	require.Equal(t, a, orderBook[2].Price)
	require.Equal(t, b, orderBook[1].Price)
	require.Equal(t, c, orderBook[0].Price)
}

func TestExecutableAmt(t *testing.T) {
	orderMap := make(types.OrderMap)
	a, _ := sdk.NewDecFromStr("0.1")
	b, _ := sdk.NewDecFromStr("0.2")
	c, _ := sdk.NewDecFromStr("0.3")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.NewInt(30000000),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.NewInt(90000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.NewInt(50000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	// make orderbook to sort orderMap
	orderBook := orderMap.SortOrderBook()

	executableBuyAmtX, executableSellAmtY := orderBook.ExecutableAmt(b)
	require.Equal(t, sdk.NewInt(140000000), executableBuyAmtX)
	require.Equal(t, sdk.NewInt(30000000), executableSellAmtY)
}

func TestPriceDirection(t *testing.T) {
	// increase case
	orderMap := make(types.OrderMap)
	a, _ := sdk.NewDecFromStr("1")
	b, _ := sdk.NewDecFromStr("1.1")
	c, _ := sdk.NewDecFromStr("1.2")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(40000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.NewInt(40000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.NewInt(20000000),
	}
	// make orderbook to sort orderMap
	orderBook := orderMap.SortOrderBook()
	poolPrice, _ := sdk.NewDecFromStr("1.0")
	result := orderBook.PriceDirection(poolPrice)
	require.Equal(t, types.Increasing, result)

	// decrease case
	orderMap = make(types.OrderMap)
	a, _ = sdk.NewDecFromStr("0.7")
	b, _ = sdk.NewDecFromStr("0.9")
	c, _ = sdk.NewDecFromStr("0.8")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(20000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.NewInt(40000000),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.NewInt(10000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	// make orderbook to sort orderMap
	orderBook = orderMap.SortOrderBook()
	poolPrice, _ = sdk.NewDecFromStr("1.0")
	result = orderBook.PriceDirection(poolPrice)
	require.Equal(t, types.Decreasing, result)

	// stay case
	orderMap = make(types.OrderMap)
	a, _ = sdk.NewDecFromStr("1.0")

	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(50000000),
		SellOfferAmt: sdk.NewInt(50000000),
	}
	orderBook = orderMap.SortOrderBook()
	poolPrice, _ = sdk.NewDecFromStr("1.0")
	result = orderBook.PriceDirection(poolPrice)
	require.Equal(t, types.Staying, result)
}

func TestComputePriceDirection(t *testing.T) {
	// increase case
	orderMap := make(types.OrderMap)
	a, _ := sdk.NewDecFromStr("1")
	b, _ := sdk.NewDecFromStr("1.1")
	c, _ := sdk.NewDecFromStr("1.2")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(40000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.NewInt(40000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.NewInt(20000000),
	}
	// make orderbook to sort orderMap
	orderBook := orderMap.SortOrderBook()

	X := orderMap[a.String()].BuyOfferAmt.ToDec().Add(orderMap[b.String()].BuyOfferAmt.ToDec())
	Y := orderMap[c.String()].SellOfferAmt.ToDec()

	poolPrice := X.Quo(Y)
	direction := orderBook.PriceDirection(poolPrice)
	result, found := orderBook.Match(X, Y)
	result2, found2 := orderBook.CalculateMatch(direction, X, Y)
	require.Equal(t, found2, found)
	require.Equal(t, result2, result)

	// decrease case
	orderMap = make(types.OrderMap)
	a, _ = sdk.NewDecFromStr("0.7")
	b, _ = sdk.NewDecFromStr("0.9")
	c, _ = sdk.NewDecFromStr("0.8")
	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(20000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	orderMap[b.String()] = types.Order{
		Price:        b,
		BuyOfferAmt:  sdk.ZeroInt(),
		SellOfferAmt: sdk.NewInt(40000000),
	}
	orderMap[c.String()] = types.Order{
		Price:        c,
		BuyOfferAmt:  sdk.NewInt(10000000),
		SellOfferAmt: sdk.ZeroInt(),
	}
	// make orderbook to sort orderMap
	orderBook = orderMap.SortOrderBook()

	X = orderMap[a.String()].BuyOfferAmt.ToDec().Add(orderMap[c.String()].BuyOfferAmt.ToDec())
	Y = orderMap[b.String()].SellOfferAmt.ToDec()

	poolPrice = X.Quo(Y)
	direction = orderBook.PriceDirection(poolPrice)
	result, found = orderBook.Match(X, Y)
	result2, found2 = orderBook.CalculateMatch(direction, X, Y)
	require.Equal(t, found2, found)
	require.Equal(t, result2, result)

	// stay case
	orderMap = make(types.OrderMap)
	a, _ = sdk.NewDecFromStr("1.0")

	orderMap[a.String()] = types.Order{
		Price:        a,
		BuyOfferAmt:  sdk.NewInt(50000000),
		SellOfferAmt: sdk.NewInt(50000000),
	}
	orderBook = orderMap.SortOrderBook()

	X = orderMap[a.String()].BuyOfferAmt.ToDec()
	Y = orderMap[a.String()].SellOfferAmt.ToDec()
	poolPrice = X.Quo(Y)

	result, _ = orderBook.Match(X, Y)
	result2 = orderBook.CalculateMatchStay(poolPrice)
	require.Equal(t, result2, result)
}

func TestCalculateMatchStay(t *testing.T) {
	currentPrice := sdk.MustNewDecFromStr("1.0")
	orderBook := types.OrderBook{
		{Price: sdk.MustNewDecFromStr("1.0"), BuyOfferAmt: sdk.NewInt(5), SellOfferAmt: sdk.NewInt(7)},
	}
	require.Equal(t, types.Staying, orderBook.PriceDirection(currentPrice))
	r := orderBook.CalculateMatchStay(currentPrice)
	require.Equal(t, sdk.NewDec(5), r.EX)
	require.Equal(t, sdk.NewDec(5), r.EY)
}

// Match Stay case with fractional match type
func TestCalculateMatchStayEdgeCase(t *testing.T) {
	currentPrice, err := sdk.NewDecFromStr("1.844380246375231658")
	require.NoError(t, err)
	var orderBook types.OrderBook
	orderbookEdgeCase := `[{"Price":"1.827780824157854573","BuyOfferAmt":"12587364000","SellOfferAmt":"6200948000","BatchPoolSwapMsgs":[{"msg_index":12,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2097894000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg36er2cp","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2097894000"},"demand_coin_denom":"denomY","order_price":"1.827780824157854573"}},{"msg_index":16,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4669506000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg44npvhm","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4669506000"},"demand_coin_denom":"denomY","order_price":"1.827780824157854573"}},{"msg_index":23,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"609066000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfzwk37gt","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"609066000"},"demand_coin_denom":"denomY","order_price":"1.827780824157854573"}},{"msg_index":39,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5210898000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfckxufsg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5210898000"},"demand_coin_denom":"denomY","order_price":"1.827780824157854573"}},{"msg_index":56,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1284220000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg44npvhm","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1284220000"},"demand_coin_denom":"denomX","order_price":"1.827780824157854573"}},{"msg_index":78,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1981368000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfhft040s","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1981368000"},"demand_coin_denom":"denomX","order_price":"1.827780824157854573"}},{"msg_index":85,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2935360000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2yrhrufk","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2935360000"},"demand_coin_denom":"denomX","order_price":"1.827780824157854573"}}]},{"Price":"1.829625204404229805","BuyOfferAmt":"9203664000","SellOfferAmt":"6971480000","BatchPoolSwapMsgs":[{"msg_index":18,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5210898000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cghxkq0yk","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5210898000"},"demand_coin_denom":"denomY","order_price":"1.829625204404229805"}},{"msg_index":36,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3992766000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf46wwkua","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3992766000"},"demand_coin_denom":"denomY","order_price":"1.829625204404229805"}},{"msg_index":44,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3155512000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgrua237l","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3155512000"},"demand_coin_denom":"denomX","order_price":"1.829625204404229805"}},{"msg_index":55,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"513688000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg5g94e2f","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"513688000"},"demand_coin_denom":"denomX","order_price":"1.829625204404229805"}},{"msg_index":61,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3302280000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfqansamx","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3302280000"},"demand_coin_denom":"denomX","order_price":"1.829625204404229805"}}]},{"Price":"1.831469584650605036","BuyOfferAmt":"18001284000","SellOfferAmt":"2311596000","BatchPoolSwapMsgs":[{"msg_index":21,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3248352000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfqansamx","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3248352000"},"demand_coin_denom":"denomY","order_price":"1.831469584650605036"}},{"msg_index":32,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5007876000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf34yvsn8","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5007876000"},"demand_coin_denom":"denomY","order_price":"1.831469584650605036"}},{"msg_index":33,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5955312000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfjmhexac","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5955312000"},"demand_coin_denom":"denomY","order_price":"1.831469584650605036"}},{"msg_index":34,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3789744000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfnxpdnq2","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3789744000"},"demand_coin_denom":"denomY","order_price":"1.831469584650605036"}},{"msg_index":65,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2311596000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfyjejm5u","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2311596000"},"demand_coin_denom":"denomX","order_price":"1.831469584650605036"}}]},{"Price":"1.833313964896980268","BuyOfferAmt":"12113646000","SellOfferAmt":"4806652000","BatchPoolSwapMsgs":[{"msg_index":6,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6632052000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg9qjf5zg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6632052000"},"demand_coin_denom":"denomY","order_price":"1.833313964896980268"}},{"msg_index":28,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5481594000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf8u28d6r","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5481594000"},"demand_coin_denom":"denomY","order_price":"1.833313964896980268"}},{"msg_index":41,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"660456000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"660456000"},"demand_coin_denom":"denomX","order_price":"1.833313964896980268"}},{"msg_index":64,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2421672000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfrnq9t4e","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2421672000"},"demand_coin_denom":"denomX","order_price":"1.833313964896980268"}},{"msg_index":73,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1724524000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfjmhexac","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1724524000"},"demand_coin_denom":"denomX","order_price":"1.833313964896980268"}}]},{"Price":"1.835158345143355500","BuyOfferAmt":"0","SellOfferAmt":"6421100000","BatchPoolSwapMsgs":[{"msg_index":47,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2715208000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2715208000"},"demand_coin_denom":"denomX","order_price":"1.835158345143355500"}},{"msg_index":58,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2678516000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cghxkq0yk","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2678516000"},"demand_coin_denom":"denomX","order_price":"1.835158345143355500"}},{"msg_index":82,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1027376000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2p3t40m7","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1027376000"},"demand_coin_denom":"denomX","order_price":"1.835158345143355500"}}]},{"Price":"1.837002725389730731","BuyOfferAmt":"9135990000","SellOfferAmt":"3852660000","BatchPoolSwapMsgs":[{"msg_index":13,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"744414000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgj52kuk7","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"744414000"},"demand_coin_denom":"denomY","order_price":"1.837002725389730731"}},{"msg_index":19,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5143224000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgcemnnmw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5143224000"},"demand_coin_denom":"denomY","order_price":"1.837002725389730731"}},{"msg_index":22,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"541392000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfpq9ygx5","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"541392000"},"demand_coin_denom":"denomY","order_price":"1.837002725389730731"}},{"msg_index":35,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2706960000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf58c6rp0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2706960000"},"demand_coin_denom":"denomY","order_price":"1.837002725389730731"}},{"msg_index":48,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2274904000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg8nhgh39","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2274904000"},"demand_coin_denom":"denomX","order_price":"1.837002725389730731"}},{"msg_index":51,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1394296000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgs80hl9n","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1394296000"},"demand_coin_denom":"denomX","order_price":"1.837002725389730731"}},{"msg_index":80,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"183460000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfetsgud6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"183460000"},"demand_coin_denom":"denomX","order_price":"1.837002725389730731"}}]},{"Price":"1.838847105636105963","BuyOfferAmt":"6226008000","SellOfferAmt":"2715208000","BatchPoolSwapMsgs":[{"msg_index":5,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6226008000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgyayapl6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6226008000"},"demand_coin_denom":"denomY","order_price":"1.838847105636105963"}},{"msg_index":43,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2715208000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgzpt7yrd","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2715208000"},"demand_coin_denom":"denomX","order_price":"1.838847105636105963"}}]},{"Price":"1.840691485882481195","BuyOfferAmt":"6496704000","SellOfferAmt":"3155512000","BatchPoolSwapMsgs":[{"msg_index":8,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6496704000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg8nhgh39","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6496704000"},"demand_coin_denom":"denomY","order_price":"1.840691485882481195"}},{"msg_index":81,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3155512000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2qvap6xv","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3155512000"},"demand_coin_denom":"denomX","order_price":"1.840691485882481195"}}]},{"Price":"1.842535866128856426","BuyOfferAmt":"0","SellOfferAmt":"1137452000","BatchPoolSwapMsgs":[{"msg_index":45,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1137452000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgyayapl6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1137452000"},"demand_coin_denom":"denomX","order_price":"1.842535866128856426"}}]},{"Price":"1.844380246375231658","BuyOfferAmt":"15700368000","SellOfferAmt":"2274904000","BatchPoolSwapMsgs":[{"msg_index":14,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"1759524000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgnfuzftv","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"1759524000"},"demand_coin_denom":"denomY","order_price":"1.844380246375231658"}},{"msg_index":24,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"1624176000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfrnq9t4e","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"1624176000"},"demand_coin_denom":"denomY","order_price":"1.844380246375231658"}},{"msg_index":25,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3248352000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfyjejm5u","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3248352000"},"demand_coin_denom":"denomY","order_price":"1.844380246375231658"}},{"msg_index":29,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4263462000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfgr8539m","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4263462000"},"demand_coin_denom":"denomY","order_price":"1.844380246375231658"}},{"msg_index":31,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4804854000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfsgjc9w4","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4804854000"},"demand_coin_denom":"denomY","order_price":"1.844380246375231658"}},{"msg_index":59,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1651140000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgcemnnmw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1651140000"},"demand_coin_denom":"denomX","order_price":"1.844380246375231658"}},{"msg_index":62,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"623764000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfpq9ygx5","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"623764000"},"demand_coin_denom":"denomX","order_price":"1.844380246375231658"}}]},{"Price":"1.846224626621606890","BuyOfferAmt":"19963830000","SellOfferAmt":"3338972000","BatchPoolSwapMsgs":[{"msg_index":11,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6429030000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgs80hl9n","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6429030000"},"demand_coin_denom":"denomY","order_price":"1.846224626621606890"}},{"msg_index":20,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5143224000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgeyd8xxu","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5143224000"},"demand_coin_denom":"denomY","order_price":"1.846224626621606890"}},{"msg_index":27,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2300916000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfxpunc83","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2300916000"},"demand_coin_denom":"denomY","order_price":"1.846224626621606890"}},{"msg_index":38,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6090660000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfhft040s","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6090660000"},"demand_coin_denom":"denomY","order_price":"1.846224626621606890"}},{"msg_index":42,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"660456000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"660456000"},"demand_coin_denom":"denomX","order_price":"1.846224626621606890"}},{"msg_index":68,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2678516000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf8u28d6r","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2678516000"},"demand_coin_denom":"denomX","order_price":"1.846224626621606890"}}]},{"Price":"1.848069006867982121","BuyOfferAmt":"0","SellOfferAmt":"3302280000","BatchPoolSwapMsgs":[{"msg_index":46,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2201520000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg9qjf5zg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2201520000"},"demand_coin_denom":"denomX","order_price":"1.848069006867982121"}},{"msg_index":70,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1100760000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cff73qycf","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1100760000"},"demand_coin_denom":"denomX","order_price":"1.848069006867982121"}}]},{"Price":"1.849913387114357353","BuyOfferAmt":"2233242000","SellOfferAmt":"10420528000","BatchPoolSwapMsgs":[{"msg_index":4,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2233242000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgrua237l","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2233242000"},"demand_coin_denom":"denomY","order_price":"1.849913387114357353"}},{"msg_index":54,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"917300000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgnfuzftv","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"917300000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}},{"msg_index":60,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3485740000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgeyd8xxu","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3485740000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}},{"msg_index":63,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"697148000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfzwk37gt","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"697148000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}},{"msg_index":66,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2421672000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf900xwfw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2421672000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}},{"msg_index":84,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1357604000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2rzw5vgn","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1357604000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}},{"msg_index":87,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1541064000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2xsjzl6m","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1541064000"},"demand_coin_denom":"denomX","order_price":"1.849913387114357353"}}]},{"Price":"1.851757767360732585","BuyOfferAmt":"23550552000","SellOfferAmt":"1577756000","BatchPoolSwapMsgs":[{"msg_index":1,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5075550000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5075550000"},"demand_coin_denom":"denomY","order_price":"1.851757767360732585"}},{"msg_index":7,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4128114000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4128114000"},"demand_coin_denom":"denomY","order_price":"1.851757767360732585"}},{"msg_index":9,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4940202000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cggv6mtwa","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4940202000"},"demand_coin_denom":"denomY","order_price":"1.851757767360732585"}},{"msg_index":15,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3113004000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg5g94e2f","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3113004000"},"demand_coin_denom":"denomY","order_price":"1.851757767360732585"}},{"msg_index":26,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"6293682000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf900xwfw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"6293682000"},"demand_coin_denom":"denomY","order_price":"1.851757767360732585"}},{"msg_index":67,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"146768000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfxpunc83","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"146768000"},"demand_coin_denom":"denomX","order_price":"1.851757767360732585"}},{"msg_index":71,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1430988000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfsgjc9w4","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1430988000"},"demand_coin_denom":"denomX","order_price":"1.851757767360732585"}}]},{"Price":"1.853602147607107816","BuyOfferAmt":"3519048000","SellOfferAmt":"5577184000","BatchPoolSwapMsgs":[{"msg_index":10,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3519048000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgf3v07n0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3519048000"},"demand_coin_denom":"denomY","order_price":"1.853602147607107816"}},{"msg_index":52,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"403612000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg36er2cp","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"403612000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}},{"msg_index":53,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"770532000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgj52kuk7","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"770532000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}},{"msg_index":72,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"146768000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf34yvsn8","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"146768000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}},{"msg_index":74,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3155512000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfnxpdnq2","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3155512000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}},{"msg_index":75,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"183460000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf58c6rp0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"183460000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}},{"msg_index":76,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"917300000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf46wwkua","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"917300000"},"demand_coin_denom":"denomX","order_price":"1.853602147607107816"}}]},{"Price":"1.855446527853483048","BuyOfferAmt":"5752290000","SellOfferAmt":"1357604000","BatchPoolSwapMsgs":[{"msg_index":3,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"3654396000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgzpt7yrd","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"3654396000"},"demand_coin_denom":"denomY","order_price":"1.855446527853483048"}},{"msg_index":17,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2097894000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgkmq56ey","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2097894000"},"demand_coin_denom":"denomY","order_price":"1.855446527853483048"}},{"msg_index":49,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1357604000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cggv6mtwa","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1357604000"},"demand_coin_denom":"denomX","order_price":"1.855446527853483048"}}]},{"Price":"1.857290908099858280","BuyOfferAmt":"2774634000","SellOfferAmt":"4256272000","BatchPoolSwapMsgs":[{"msg_index":37,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"2774634000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfk5amqjz","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"2774634000"},"demand_coin_denom":"denomY","order_price":"1.857290908099858280"}},{"msg_index":50,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2128136000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgf3v07n0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2128136000"},"demand_coin_denom":"denomX","order_price":"1.857290908099858280"}},{"msg_index":77,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"256844000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfk5amqjz","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"256844000"},"demand_coin_denom":"denomX","order_price":"1.857290908099858280"}},{"msg_index":83,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1871292000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c2zlcqe4p","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1871292000"},"demand_coin_denom":"denomX","order_price":"1.857290908099858280"}}]},{"Price":"1.859135288346233511","BuyOfferAmt":"10760166000","SellOfferAmt":"5283648000","BatchPoolSwapMsgs":[{"msg_index":2,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"1421154000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"1421154000"},"demand_coin_denom":"denomY","order_price":"1.859135288346233511"}},{"msg_index":30,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"4331136000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cff73qycf","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"4331136000"},"demand_coin_denom":"denomY","order_price":"1.859135288346233511"}},{"msg_index":40,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"5007876000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfetsgud6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"5007876000"},"demand_coin_denom":"denomY","order_price":"1.859135288346233511"}},{"msg_index":57,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1137452000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgkmq56ey","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1137452000"},"demand_coin_denom":"denomX","order_price":"1.859135288346233511"}},{"msg_index":69,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"293536000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfgr8539m","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"293536000"},"demand_coin_denom":"denomX","order_price":"1.859135288346233511"}},{"msg_index":79,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3302280000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfckxufsg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3302280000"},"demand_coin_denom":"denomX","order_price":"1.859135288346233511"}},{"msg_index":86,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"550380000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5c297phf5y","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"550380000"},"demand_coin_denom":"denomX","order_price":"1.859135288346233511"}}]}]`
	json.Unmarshal([]byte(orderbookEdgeCase), &orderBook)
	r := orderBook.CalculateMatchStay(currentPrice)
	require.Equal(t, types.FractionalMatch, r.MatchType)
	// stay case with fractional
}

// Match Stay case with no match type
func TestCalculateNoMatchEdgeCase(t *testing.T) {
	currentPrice, err := sdk.NewDecFromStr("1.007768598527187219")
	require.NoError(t, err)
	var orderBook types.OrderBook
	orderbookEdgeCase := `[{"Price":"1.007768598527187219","BuyOfferAmt":"0","SellOfferAmt":"417269600","BatchPoolSwapMsgs":[{"msg_index":1,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"417269600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"417269600"},"demand_coin_denom":"denomX","order_price":"1.007768598527187219"}}]},{"Price":"1.011799672921295968","BuyOfferAmt":"0","SellOfferAmt":"2190665400","BatchPoolSwapMsgs":[{"msg_index":2,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2190665400"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2190665400"},"demand_coin_denom":"denomX","order_price":"1.011799672921295968"}}]}]`
	_ = json.Unmarshal([]byte(orderbookEdgeCase), &orderBook)
	r := orderBook.CalculateMatchStay(currentPrice)
	require.Equal(t, types.NoMatch, r.MatchType)
	// stay case with fractional
}

// Reproduce GetOrderMapEdgeCase, selling Y for X case, ErrInvalidDenom case
func TestMakeOrderMapEdgeCase(t *testing.T) {
	onlyNotMatched := false
	var swapMsgs []*types.SwapMsgState
	swapMsgsJSON := `[{"msg_index":1,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"19228500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"19228500"},"demand_coin_denom":"denomY","order_price":"0.027506527499265415"}},{"msg_index":2,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"141009000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"141009000"},"demand_coin_denom":"denomY","order_price":"0.027341323129900457"}},{"msg_index":3,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"23501500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgzpt7yrd","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"23501500"},"demand_coin_denom":"denomY","order_price":"0.027616663745508720"}},{"msg_index":4,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"200831000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgrua237l","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"200831000"},"demand_coin_denom":"denomY","order_price":"0.027589129683947893"}},{"msg_index":5,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"160237500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgyayapl6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"160237500"},"demand_coin_denom":"denomY","order_price":"0.027313789068339631"}},{"msg_index":6,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"175193000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg9qjf5zg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"175193000"},"demand_coin_denom":"denomY","order_price":"0.027478993437704589"}},{"msg_index":7,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"183739000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"183739000"},"demand_coin_denom":"denomY","order_price":"0.027699265930191198"}},{"msg_index":8,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"32047500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg8nhgh39","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"32047500"},"demand_coin_denom":"denomY","order_price":"0.027451459376143762"}},{"msg_index":9,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"111098000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cggv6mtwa","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"111098000"},"demand_coin_denom":"denomY","order_price":"0.027286255006778805"}},{"msg_index":10,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"166647000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgf3v07n0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"166647000"},"demand_coin_denom":"denomY","order_price":"0.027341323129900457"}},{"msg_index":11,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"98279000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgs80hl9n","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"98279000"},"demand_coin_denom":"denomY","order_price":"0.027368857191461284"}},{"msg_index":12,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"8546000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg36er2cp","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"8546000"},"demand_coin_denom":"denomY","order_price":"0.027396391253022110"}},{"msg_index":13,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"87596500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgj52kuk7","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"87596500"},"demand_coin_denom":"denomY","order_price":"0.027451459376143762"}},{"msg_index":14,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"111098000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgnfuzftv","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"111098000"},"demand_coin_denom":"denomY","order_price":"0.027478993437704589"}},{"msg_index":15,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"38457000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg5g94e2f","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"38457000"},"demand_coin_denom":"denomY","order_price":"0.027451459376143762"}},{"msg_index":16,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"153828000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg44npvhm","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"153828000"},"demand_coin_denom":"denomY","order_price":"0.027616663745508720"}},{"msg_index":17,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"70504500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgkmq56ey","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"70504500"},"demand_coin_denom":"denomY","order_price":"0.027451459376143762"}},{"msg_index":18,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"47003000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cghxkq0yk","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"47003000"},"demand_coin_denom":"denomY","order_price":"0.027396391253022110"}},{"msg_index":19,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"132463000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgcemnnmw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"132463000"},"demand_coin_denom":"denomY","order_price":"0.027726799991752025"}},{"msg_index":20,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"66231500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgeyd8xxu","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"66231500"},"demand_coin_denom":"denomY","order_price":"0.027561595622387067"}},{"msg_index":21,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"119644000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfqansamx","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"119644000"},"demand_coin_denom":"denomY","order_price":"0.027506527499265415"}},{"msg_index":22,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"17092000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfpq9ygx5","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"17092000"},"demand_coin_denom":"denomY","order_price":"0.027341323129900457"}},{"msg_index":23,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"209377000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfzwk37gt","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"209377000"},"demand_coin_denom":"denomY","order_price":"0.027478993437704589"}},{"msg_index":24,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"207240500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfrnq9t4e","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"207240500"},"demand_coin_denom":"denomY","order_price":"0.027396391253022110"}},{"msg_index":25,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"155964500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfyjejm5u","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"155964500"},"demand_coin_denom":"denomY","order_price":"0.027423925314582936"}},{"msg_index":26,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"194421500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf900xwfw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"194421500"},"demand_coin_denom":"denomY","order_price":"0.027286255006778805"}},{"msg_index":27,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"102552000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfxpunc83","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"102552000"},"demand_coin_denom":"denomY","order_price":"0.027368857191461284"}},{"msg_index":28,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"151691500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf8u28d6r","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"151691500"},"demand_coin_denom":"denomY","order_price":"0.027478993437704589"}},{"msg_index":29,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"113234500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfgr8539m","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"113234500"},"demand_coin_denom":"denomY","order_price":"0.027368857191461284"}},{"msg_index":30,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"117507500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cff73qycf","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"117507500"},"demand_coin_denom":"denomY","order_price":"0.027423925314582936"}},{"msg_index":31,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"141009000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfsgjc9w4","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"141009000"},"demand_coin_denom":"denomY","order_price":"0.027423925314582936"}},{"msg_index":32,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"200831000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf34yvsn8","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"200831000"},"demand_coin_denom":"denomY","order_price":"0.027534061560826241"}},{"msg_index":33,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"141009000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfjmhexac","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"141009000"},"demand_coin_denom":"denomY","order_price":"0.027726799991752025"}},{"msg_index":34,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"98279000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfnxpdnq2","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"98279000"},"demand_coin_denom":"denomY","order_price":"0.027478993437704589"}},{"msg_index":35,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"76914000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf58c6rp0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"76914000"},"demand_coin_denom":"denomY","order_price":"0.027423925314582936"}},{"msg_index":36,"executed":true,"exchanged_offer_coin":{"denom":"denomX","amount":"0"},"remaining_offer_coin":{"denom":"denomX","amount":"23501500"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf46wwkua","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomX","amount":"23501500"},"demand_coin_denom":"denomY","order_price":"0.027754334053312851"}},{"msg_index":37,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4733282800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4733282800"},"demand_coin_denom":"denomX","order_price":"0.027699265930191198"}},{"msg_index":38,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3957334800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3957334800"},"demand_coin_denom":"denomX","order_price":"0.027478993437704589"}},{"msg_index":39,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2483033600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgzpt7yrd","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2483033600"},"demand_coin_denom":"denomX","order_price":"0.027589129683947893"}},{"msg_index":40,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"5509230800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgrua237l","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"5509230800"},"demand_coin_denom":"denomX","order_price":"0.027561595622387067"}},{"msg_index":41,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2327844000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgyayapl6","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2327844000"},"demand_coin_denom":"denomX","order_price":"0.027423925314582936"}},{"msg_index":42,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4733282800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg9qjf5zg","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4733282800"},"demand_coin_denom":"denomX","order_price":"0.027451459376143762"}},{"msg_index":43,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"7061126800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"7061126800"},"demand_coin_denom":"denomX","order_price":"0.027726799991752025"}},{"msg_index":44,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4655688000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg8nhgh39","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4655688000"},"demand_coin_denom":"denomX","order_price":"0.027589129683947893"}},{"msg_index":45,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3026197200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cggv6mtwa","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3026197200"},"demand_coin_denom":"denomX","order_price":"0.027589129683947893"}},{"msg_index":46,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"7293911200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgf3v07n0","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"7293911200"},"demand_coin_denom":"denomX","order_price":"0.027616663745508720"}},{"msg_index":47,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4810877600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgs80hl9n","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4810877600"},"demand_coin_denom":"denomX","order_price":"0.027534061560826241"}},{"msg_index":48,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4345308800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg36er2cp","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4345308800"},"demand_coin_denom":"denomX","order_price":"0.027451459376143762"}},{"msg_index":49,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"5509230800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgj52kuk7","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"5509230800"},"demand_coin_denom":"denomX","order_price":"0.027368857191461284"}},{"msg_index":50,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4190119200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgnfuzftv","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4190119200"},"demand_coin_denom":"denomX","order_price":"0.027451459376143762"}},{"msg_index":51,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"543163600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg5g94e2f","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"543163600"},"demand_coin_denom":"denomX","order_price":"0.027286255006778805"}},{"msg_index":52,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4578093200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cg44npvhm","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4578093200"},"demand_coin_denom":"denomX","order_price":"0.027506527499265415"}},{"msg_index":53,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"6517963200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgkmq56ey","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"6517963200"},"demand_coin_denom":"denomX","order_price":"0.027368857191461284"}},{"msg_index":54,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"4190119200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cghxkq0yk","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"4190119200"},"demand_coin_denom":"denomX","order_price":"0.027368857191461284"}},{"msg_index":55,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1939870000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgcemnnmw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1939870000"},"demand_coin_denom":"denomX","order_price":"0.027754334053312851"}},{"msg_index":56,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"1163922000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgeyd8xxu","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"1163922000"},"demand_coin_denom":"denomX","order_price":"0.027478993437704589"}},{"msg_index":57,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"5897204800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfqansamx","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"5897204800"},"demand_coin_denom":"denomX","order_price":"0.027644197807069546"}},{"msg_index":58,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"155189600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfpq9ygx5","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"155189600"},"demand_coin_denom":"denomX","order_price":"0.027671731868630372"}},{"msg_index":59,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2250249200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfzwk37gt","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2250249200"},"demand_coin_denom":"denomX","order_price":"0.027286255006778805"}},{"msg_index":60,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"2948602400"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfrnq9t4e","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"2948602400"},"demand_coin_denom":"denomX","order_price":"0.027286255006778805"}},{"msg_index":61,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"7449100800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfyjejm5u","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"7449100800"},"demand_coin_denom":"denomX","order_price":"0.027313789068339631"}},{"msg_index":62,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"6129989200"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf900xwfw","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"6129989200"},"demand_coin_denom":"denomX","order_price":"0.027341323129900457"}},{"msg_index":63,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3491766000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfxpunc83","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3491766000"},"demand_coin_denom":"denomX","order_price":"0.027534061560826241"}},{"msg_index":64,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"6362773600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf8u28d6r","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"6362773600"},"demand_coin_denom":"denomX","order_price":"0.027726799991752025"}},{"msg_index":65,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"7138721600"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfgr8539m","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"7138721600"},"demand_coin_denom":"denomX","order_price":"0.027534061560826241"}},{"msg_index":66,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3724550400"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cff73qycf","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3724550400"},"demand_coin_denom":"denomX","order_price":"0.027616663745508720"}},{"msg_index":67,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"3103792000"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfsgjc9w4","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"3103792000"},"demand_coin_denom":"denomX","order_price":"0.027589129683947893"}},{"msg_index":68,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"232784400"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cf34yvsn8","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"232784400"},"demand_coin_denom":"denomX","order_price":"0.027478993437704589"}},{"msg_index":69,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"6052394400"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfjmhexac","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"6052394400"},"demand_coin_denom":"denomX","order_price":"0.027478993437704589"}},{"msg_index":70,"executed":true,"exchanged_offer_coin":{"denom":"denomY","amount":"0"},"remaining_offer_coin":{"denom":"denomY","amount":"5121256800"},"msg":{"swap_requester_address":"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cfnxpdnq2","pool_id":1,"pool_type_id":1,"offer_coin":{"denom":"denomY","amount":"5121256800"},"demand_coin_denom":"denomX","order_price":"0.027644197807069546"}}]`
	_ = json.Unmarshal([]byte(swapMsgsJSON), &swapMsgs)
	orderMap, xToY, yToX := types.MakeOrderMap(swapMsgs, DenomX, DenomY, onlyNotMatched)
	require.NotZero(t, len(orderMap))
	require.NotNil(t, xToY)
	require.NotNil(t, yToX)

	// ErrInvalidDenom case
	require.Panics(t, func() {
		types.MakeOrderMap(swapMsgs, "12421miklfdjnfiasdjfidosa8381813818---", DenomY, onlyNotMatched)
	})
}

func TestOrderbookValidate(t *testing.T) {
	for _, testCase := range []struct {
		currentPrice string
		buyPrice     string
		sellPrice    string
		valid        bool
	}{
		{
			currentPrice: "1.0",
			buyPrice:     "0.99",
			sellPrice:    "1.01",
			valid:        true,
		},
		{
			// maxBuyOrderPrice > minSellOrderPrice
			currentPrice: "1.0",
			buyPrice:     "1.01",
			sellPrice:    "0.99",
			valid:        false,
		},
		{
			currentPrice: "1.0",
			buyPrice:     "1.1",
			sellPrice:    "1.2",
			valid:        true,
		},
		{
			// maxBuyOrderPrice/currentPrice > 1.10
			currentPrice: "1.0",
			buyPrice:     "1.11",
			sellPrice:    "1.2",
			valid:        false,
		},
		{
			currentPrice: "1.0",
			buyPrice:     "0.8",
			sellPrice:    "0.9",
			valid:        true,
		},
		{
			// minSellOrderPrice/currentPrice < 0.90
			currentPrice: "1.0",
			buyPrice:     "0.8",
			sellPrice:    "0.89",
			valid:        false,
		},
		{
			// not positive price
			currentPrice: "0.0",
			buyPrice:     "0.00000000001",
			sellPrice:    "0.000000000011",
			valid:        false,
		},
	} {
		buyPrice := sdk.MustNewDecFromStr(testCase.buyPrice)
		sellPrice := sdk.MustNewDecFromStr(testCase.sellPrice)
		orderMap := types.OrderMap{
			buyPrice.String(): {
				Price:        buyPrice,
				BuyOfferAmt:  sdk.OneInt(),
				SellOfferAmt: sdk.ZeroInt(),
			},
			sellPrice.String(): {
				Price:        sellPrice,
				BuyOfferAmt:  sdk.ZeroInt(),
				SellOfferAmt: sdk.OneInt(),
			},
		}
		orderBook := orderMap.SortOrderBook()
		require.Equal(t, testCase.valid, orderBook.Validate(sdk.MustNewDecFromStr(testCase.currentPrice)))
	}
}

func TestCountNotMatchedMsgs(t *testing.T) {
	for _, tc := range []struct {
		msgs []*types.SwapMsgState
		cnt  int
	}{
		{
			[]*types.SwapMsgState{},
			0,
		},
		{
			[]*types.SwapMsgState{
				{Executed: true, Succeeded: false},
				{Executed: true, Succeeded: false},
			},
			2,
		},
		{
			[]*types.SwapMsgState{
				{},
				{Executed: true, Succeeded: true, ToBeDeleted: false},
				{Executed: true, Succeeded: true, ToBeDeleted: true},
			},
			0,
		},
	} {
		require.Equal(t, tc.cnt, types.CountNotMatchedMsgs(tc.msgs))
	}
}

func TestCountFractionalMatchedMsgs(t *testing.T) {
	for _, tc := range []struct {
		msgs []*types.SwapMsgState
		cnt  int
	}{
		{
			[]*types.SwapMsgState{},
			0,
		},
		{
			[]*types.SwapMsgState{
				{Executed: true, Succeeded: true, ToBeDeleted: false},
				{Executed: true, Succeeded: true, ToBeDeleted: false},
			},
			2,
		},
		{
			[]*types.SwapMsgState{
				{},
				{Executed: true, Succeeded: false},
				{Executed: true, Succeeded: true, ToBeDeleted: true},
			},
			0,
		},
	} {
		require.Equal(t, tc.cnt, types.CountFractionalMatchedMsgs(tc.msgs))
	}
}
