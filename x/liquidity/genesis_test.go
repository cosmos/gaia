package liquidity_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestGenesisState(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	simapp := app.Setup(false)

	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})
	genesis := types.DefaultGenesisState()

	liquidity.InitGenesis(ctx, simapp.LiquidityKeeper, *genesis)

	defaultGenesisExported := liquidity.ExportGenesis(ctx, simapp.LiquidityKeeper)

	require.Equal(t, genesis, defaultGenesisExported)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair("denomX", "denomY")

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
	offerCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000))}
	orderPrices := []sdk.Dec{price}
	orderAddrs := addrs[1:2]
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, true)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolID, true)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	genesisExported := liquidity.ExportGenesis(ctx, simapp.LiquidityKeeper)
	bankGenesisExported := simapp.BankKeeper.ExportGenesis(ctx)

	simapp2 := app.Setup(false)

	ctx2 := simapp2.BaseApp.NewContext(false, tmproto.Header{})
	ctx2 = ctx2.WithBlockHeight(1)

	simapp2.BankKeeper.InitGenesis(ctx2, bankGenesisExported)
	liquidity.InitGenesis(ctx2, simapp2.LiquidityKeeper, *genesisExported)
	simapp2GenesisExported := liquidity.ExportGenesis(ctx2, simapp2.LiquidityKeeper)
	require.Equal(t, genesisExported, simapp2GenesisExported)
}
