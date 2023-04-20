package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestGenesis(t *testing.T) {
	simapp, ctx := app.CreateTestInput()

	lk := simapp.LiquidityKeeper

	// default genesis state
	genState := types.DefaultGenesisState()
	require.Equal(t, sdk.NewDecWithPrec(3, 3), genState.Params.SwapFeeRate)

	// change swap fee rate
	params := lk.GetParams(ctx)
	params.SwapFeeRate = sdk.NewDecWithPrec(5, 3)

	// set params
	lk.SetParams(ctx, params)

	newGenState := lk.ExportGenesis(ctx)
	require.Equal(t, sdk.NewDecWithPrec(5, 3), newGenState.Params.SwapFeeRate)

	fmt.Println("newGenState: ", newGenState)
}

func TestGenesisState(t *testing.T) {
	simapp, ctx := app.CreateTestInput()

	params := simapp.LiquidityKeeper.GetParams(ctx)
	paramsDefault := simapp.LiquidityKeeper.GetParams(ctx)
	genesis := types.DefaultGenesisState()

	invalidDenom := "invalid denom---"
	invalidDenomErrMsg := fmt.Sprintf("invalid denom: %s", invalidDenom)
	params.PoolCreationFee = sdk.Coins{sdk.Coin{Denom: invalidDenom, Amount: sdk.NewInt(0)}}
	require.EqualError(t, params.Validate(), invalidDenomErrMsg)

	params = simapp.LiquidityKeeper.GetParams(ctx)
	params.SwapFeeRate = sdk.NewDec(-1)
	negativeSwapFeeErrMsg := fmt.Sprintf("swap fee rate must not be negative: %s", params.SwapFeeRate)
	genesisState := types.NewGenesisState(params, genesis.PoolRecords)
	require.EqualError(t, types.ValidateGenesis(*genesisState), negativeSwapFeeErrMsg)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)
	X := sdk.NewInt(100_000_000)
	Y := sdk.NewInt(200_000_000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10_000))
	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])

	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)

	poolCoins := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	app.TestDepositPool(t, simapp, ctx, sdk.NewInt(30_000_000), sdk.NewInt(20_000_000), addrs[1:2], poolID, false)

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	poolCoinBalanceCreator := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)
	poolCoinBalance := simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom)
	require.Equal(t, sdk.NewInt(100_000), poolCoinBalance.Amount)
	require.Equal(t, poolCoins.QuoRaw(10), poolCoinBalance.Amount)

	balanceXRefunded := simapp.BankKeeper.GetBalance(ctx, addrs[1], denomX)
	balanceYRefunded := simapp.BankKeeper.GetBalance(ctx, addrs[1], denomY)
	require.Equal(t, sdk.NewInt(20000000), balanceXRefunded.Amount)
	require.Equal(t, sdk.ZeroInt(), balanceYRefunded.Amount)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	// validate pool records
	newGenesis := simapp.LiquidityKeeper.ExportGenesis(ctx)
	genesisState = types.NewGenesisState(paramsDefault, newGenesis.PoolRecords)
	require.NoError(t, types.ValidateGenesis(*genesisState))

	pool.TypeId = 5
	simapp.LiquidityKeeper.SetPool(ctx, pool)
	newGenesisBrokenPool := simapp.LiquidityKeeper.ExportGenesis(ctx)
	require.NoError(t, types.ValidateGenesis(*newGenesisBrokenPool))
	require.Equal(t, 1, len(newGenesisBrokenPool.PoolRecords))

	err := simapp.LiquidityKeeper.ValidatePoolRecord(ctx, newGenesisBrokenPool.PoolRecords[0])
	require.ErrorIs(t, err, types.ErrPoolTypeNotExists)

	// not initialized genState of other module (auth, bank, ... ) only liquidity module
	reserveCoins := simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.Equal(t, 2, len(reserveCoins))
	simapp2 := app.Setup(false)
	ctx2 := simapp2.BaseApp.NewContext(false, tmproto.Header{})
	require.Panics(t, func() {
		simapp2.LiquidityKeeper.InitGenesis(ctx2, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp2, ctx, pool.GetReserveAccount(), reserveCoins)
		simapp2.LiquidityKeeper.InitGenesis(ctx2, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp2, ctx, addrs[0], sdk.Coins{poolCoinBalanceCreator})
		simapp2.LiquidityKeeper.InitGenesis(ctx2, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp2, ctx2, addrs[1], sdk.Coins{poolCoinBalance})
		simapp2.LiquidityKeeper.InitGenesis(ctx2, *newGenesis)
	})

	simapp3 := app.Setup(false)
	ctx3 := simapp3.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeight(ctx.BlockHeight())
	require.Panics(t, func() {
		simapp3.LiquidityKeeper.InitGenesis(ctx3, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp3, ctx, pool.GetReserveAccount(), reserveCoins)
		simapp3.LiquidityKeeper.InitGenesis(ctx3, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp3, ctx, addrs[0], sdk.Coins{poolCoinBalanceCreator})
		simapp3.LiquidityKeeper.InitGenesis(ctx3, *newGenesis)
	})
	require.Panics(t, func() {
		app.SaveAccount(simapp3, ctx3, addrs[1], sdk.Coins{poolCoinBalance})
		simapp3.LiquidityKeeper.InitGenesis(ctx3, *newGenesis)
	})
}
