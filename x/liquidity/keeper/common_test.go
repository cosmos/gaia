package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// createTestInput Returns a simapp with custom LiquidityKeeper
// to avoid messing with the hooks.
func createTestInput() (*app.GaiaApp, sdk.Context) {
	return app.CreateTestInput()
}

func createLiquidity(t *testing.T, ctx sdk.Context, simapp *app.GaiaApp) (
	[]sdk.AccAddress, []types.Pool, []types.PoolBatch,
	[]types.DepositMsgState, []types.WithdrawMsgState,
) {
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
	app.TestCreatePool(t, simapp, ctx, A, B, denomA, denomB, addrs[1])

	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	price, _ := sdk.NewDecFromStr("1.1")
	priceY, _ := sdk.NewDecFromStr("1.2")
	xOfferCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000))}
	yOfferCoins := []sdk.Coin{sdk.NewCoin(denomY, sdk.NewInt(5000))}
	xOrderPrices := []sdk.Dec{price}
	yOrderPrices := []sdk.Dec{priceY}
	xOrderAddrs := addrs[1:2]
	yOrderAddrs := addrs[2:3]

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(50), addrs[1:2], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(500), addrs[1:2], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(50), addrs[2:3], poolID, false)
	app.TestWithdrawPool(t, simapp, ctx, sdk.NewInt(500), addrs[2:3], poolID, false)

	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, xOfferCoins, xOrderPrices, xOrderAddrs, poolID, false)
	app.TestSwapPool(t, simapp, ctx, yOfferCoins, yOrderPrices, yOrderAddrs, poolID, false)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	batches := simapp.LiquidityKeeper.GetAllPoolBatches(ctx)
	depositMsgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, batches[0])
	withdrawMsgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, batches[0])
	return addrs, pools, batches, depositMsgs, withdrawMsgs
}

func createTestPool(X, Y sdk.Coin) (*app.GaiaApp, sdk.Context, types.Pool, sdk.AccAddress, error) {
	simapp, ctx := createTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)

	depositCoins := sdk.NewCoins(X, Y)
	creatorAddr := app.AddRandomTestAddr(simapp, ctx, depositCoins.Add(params.PoolCreationFee...))

	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creatorAddr, types.DefaultPoolTypeID, depositCoins))
	if err != nil {
		return nil, sdk.Context{}, types.Pool{}, sdk.AccAddress{}, err
	}

	return simapp, ctx, pool, creatorAddr, nil
}
