package simulation_test

import (
	"math/rand"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	lapp "github.com/cosmos/gaia/v9/app"
	liquidityparams "github.com/cosmos/gaia/v9/app/params"
	"github.com/cosmos/gaia/v9/x/liquidity/simulation"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(false)

	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := simtypes.RandomAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{liquidityparams.DefaultWeightMsgCreatePool, types.ModuleName, types.TypeMsgCreatePool},
		{liquidityparams.DefaultWeightMsgDepositWithinBatch, types.ModuleName, types.TypeMsgDepositWithinBatch},
		{liquidityparams.DefaultWeightMsgWithdrawWithinBatch, types.ModuleName, types.TypeMsgWithdrawWithinBatch},
		{liquidityparams.DefaultWeightMsgSwapWithinBatch, types.ModuleName, types.TypeMsgSwapWithinBatch},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgCreatePool tests the normal scenario of a valid message of type TypeMsgCreatePool.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgCreatePool(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup randomly generated liquidity pool creation fees
	feeCoins := simulation.GenPoolCreationFee(r)
	params := app.LiquidityKeeper.GetParams(ctx)
	params.PoolCreationFee = feeCoins
	app.LiquidityKeeper.SetParams(ctx, params)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgCreatePool(app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgCreatePool
	require.NoError(t, types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg))

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos10kn7aww37y27c4lggjx6mycyhr927677rkp7x0", msg.GetPoolCreator().String())
	require.Equal(t, types.DefaultPoolTypeID, msg.PoolTypeId)
	require.Equal(t, "90341750hsuk,28960633ijmo", msg.DepositCoins.String())
	require.Equal(t, types.TypeMsgCreatePool, msg.Type())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgDepositWithinBatch tests the normal scenario of a valid message of type TypeMsgDepositWithinBatch.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgDepositWithinBatch(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup random liquidity pools
	setupLiquidityPools(t, r, app, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgDepositWithinBatch(app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgDepositWithinBatch
	require.NoError(t, types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg))

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos10kn7aww37y27c4lggjx6mycyhr927677rkp7x0", msg.GetDepositor().String())
	require.Equal(t, "38511541fgae,71186277jxulr", msg.DepositCoins.String())
	require.Equal(t, types.TypeMsgDepositWithinBatch, msg.Type())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgWithdrawWithinBatch tests the normal scenario of a valid message of type TypeMsgWithdrawWithinBatch.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgWithdrawWithinBatch(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup random liquidity pools
	setupLiquidityPools(t, r, app, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgWithdrawWithinBatch(app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgWithdrawWithinBatch
	require.NoError(t, types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg))

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos10kn7aww37y27c4lggjx6mycyhr927677rkp7x0", msg.GetWithdrawer().String())
	require.Equal(t, "3402627138556poolA295B958C22781F58E51E1E4E8205F5E8E041D65F7E7AB5D7DDECCFFA7A75B01", msg.PoolCoin.String())
	require.Equal(t, types.TypeMsgWithdrawWithinBatch, msg.Type())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgSwapWithinBatch tests the normal scenario of a valid message of type TypeMsgSwapWithinBatch.
// Abnormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgSwapWithinBatch(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup a single account
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 1)

	// setup random liquidity pools
	setupLiquidityPools(t, r, app, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgSwapWithinBatch(app.AccountKeeper, app.BankKeeper, app.LiquidityKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgSwapWithinBatch
	require.NoError(t, types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg))

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos10kn7aww37y27c4lggjx6mycyhr927677rkp7x0", msg.GetSwapRequester().String())
	require.Equal(t, "6453297fgae", msg.OfferCoin.String())
	require.Equal(t, "jxulr", msg.DemandCoinDenom)
	require.Equal(t, types.TypeMsgSwapWithinBatch, msg.Type())
	require.Len(t, futureOperations, 0)
}

func createTestApp(isCheckTx bool) (*lapp.GaiaApp, sdk.Context) {
	app := lapp.Setup(false)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.MintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *lapp.GaiaApp, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(1_000_000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		lapp.SaveAccount(app, ctx, account.Address, initCoins)
	}

	return accounts
}

func setupLiquidityPools(t *testing.T, r *rand.Rand, app *lapp.GaiaApp, ctx sdk.Context, accounts []simtypes.Account) {
	params := app.StakingKeeper.GetParams(ctx)

	for _, account := range accounts {
		// random denom with a length from 4 to 6 characters
		denomA := simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 4, 6))
		denomB := simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 4, 6))
		denomA, denomB = types.AlphabeticalDenomPair(strings.ToLower(denomA), strings.ToLower(denomB))

		fees := sdk.NewCoin(params.GetBondDenom(), sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e10, 1e12))))

		// mint random amounts of denomA and denomB coins
		mintCoinA := sdk.NewCoin(denomA, sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e14, 1e15))))
		mintCoinB := sdk.NewCoin(denomB, sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e14, 1e15))))
		mintCoins := sdk.NewCoins(mintCoinA, mintCoinB, fees)
		err := app.BankKeeper.MintCoins(ctx, types.ModuleName, mintCoins)
		require.NoError(t, err)

		// transfer random amounts to the simulated random account
		coinA := sdk.NewCoin(denomA, sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e12, 1e14))))
		coinB := sdk.NewCoin(denomB, sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e12, 1e14))))
		coins := sdk.NewCoins(coinA, coinB, fees)
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, account.Address, coins)
		require.NoError(t, err)

		// create liquidity pool with random deposit amounts
		account := app.AccountKeeper.GetAccount(ctx, account.Address)
		depositCoinA := sdk.NewCoin(denomA, sdk.NewInt(int64(simtypes.RandIntBetween(r, int(types.DefaultMinInitDepositAmount.Int64()), 1e8))))
		depositCoinB := sdk.NewCoin(denomB, sdk.NewInt(int64(simtypes.RandIntBetween(r, int(types.DefaultMinInitDepositAmount.Int64()), 1e8))))
		depositCoins := sdk.NewCoins(depositCoinA, depositCoinB)

		createPoolMsg := types.NewMsgCreatePool(account.GetAddress(), types.DefaultPoolTypeID, depositCoins)

		_, err = app.LiquidityKeeper.CreatePool(ctx, createPoolMsg)
		require.NoError(t, err)
	}
}
