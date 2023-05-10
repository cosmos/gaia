package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/keeper"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestLiquidityPool(t *testing.T) {
	app, ctx := createTestInput()
	lp := types.Pool{
		Id:                    0,
		TypeId:                0,
		ReserveCoinDenoms:     []string{"a", "b"},
		ReserveAccountAddress: "",
		PoolCoinDenom:         "poolCoin",
	}
	app.LiquidityKeeper.SetPool(ctx, lp)

	lpGet, found := app.LiquidityKeeper.GetPool(ctx, 0)
	require.True(t, found)
	require.Equal(t, lp, lpGet)
}

func TestCreatePool(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH" //nolint:goconst // these are immediately reassigned in each case
	denomB := "uUSD" //nolint:goconst // these are immediately reassigned in each case
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

	invalidMsg := types.NewMsgCreatePool(addrs[0], 2, depositBalance)
	_, err = simapp.LiquidityKeeper.CreatePool(ctx, invalidMsg)
	require.ErrorIs(t, err, types.ErrPoolTypeNotExists)

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

func TestCreatePoolInsufficientAmount(t *testing.T) {
	simapp, ctx := createTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)

	depositCoins := sdk.NewCoins(sdk.NewInt64Coin(DenomX, 1000), sdk.NewInt64Coin(DenomY, 1000))
	creator := app.AddRandomTestAddr(simapp, ctx, depositCoins.Add(params.PoolCreationFee...))

	// Depositing coins that are less than params.MinInitDepositAmount.
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creator, types.DefaultPoolTypeID, depositCoins))
	require.ErrorIs(t, err, types.ErrLessThanMinInitDeposit)

	fakeDepositCoins := depositCoins.Add(
		sdk.NewCoin(DenomX, params.MinInitDepositAmount),
		sdk.NewCoin(DenomY, params.MinInitDepositAmount),
	)
	// Depositing coins that are greater than the depositor has.
	_, err = simapp.LiquidityKeeper.CreatePool(
		ctx, types.NewMsgCreatePool(creator, types.DefaultPoolTypeID, fakeDepositCoins),
	)
	require.ErrorIs(t, err, types.ErrInsufficientBalance)
}

func TestPoolCreationFee(t *testing.T) {
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

	// Set PoolCreationFee for fail (insufficient balances for pool creation fee)
	params.PoolCreationFee = depositBalance
	simapp.LiquidityKeeper.SetParams(ctx, params)

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.ErrorIs(t, types.ErrInsufficientPoolCreationFee, err)

	// Set PoolCreationFee for success
	params.PoolCreationFee = types.DefaultPoolCreationFee
	simapp.LiquidityKeeper.SetParams(ctx, params)
	feePoolAcc := simapp.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	feePoolBalance := simapp.BankKeeper.GetAllBalances(ctx, feePoolAcc)
	msg = types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)
	_, err = simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	// Verify PoolCreationFee pay successfully
	feePoolBalance = feePoolBalance.Add(params.PoolCreationFee...)
	require.Equal(t, params.PoolCreationFee, feePoolBalance)
	require.Equal(t, feePoolBalance, simapp.BankKeeper.GetAllBalances(ctx, feePoolAcc))
}

func TestExecuteDeposit(t *testing.T) {
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

func TestExecuteDepositTruncation(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 4, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	initDeposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(10000000*1000000)), sdk.NewCoin(denomB, sdk.NewInt(200000000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], initDeposit)
	app.SaveAccount(simapp, ctx, addrs[1], initDeposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	require.Equal(t, initDeposit, depositBalance)

	depositA = simapp.BankKeeper.GetBalance(ctx, addrs[1], denomA)
	depositB = simapp.BankKeeper.GetBalance(ctx, addrs[1], denomB)
	depositBalance = sdk.NewCoins(depositA, depositB)

	require.Equal(t, initDeposit, depositBalance)

	createMsg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, createMsg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(19*1000000)), sdk.NewCoin(denomB, sdk.NewInt(400*1000000)))
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
	depositorBalancePoolCoin := simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom)
	depositorBalanceDenomA := simapp.BankKeeper.GetBalance(ctx, addrs[1], denomA)
	depositorBalanceDenomB := simapp.BankKeeper.GetBalance(ctx, addrs[1], denomB)
	require.Equal(t, depositBalance.AmountOf(denomA).Sub(sdk.NewInt(10*1000000)), depositorBalanceDenomA.Amount)
	require.Equal(t, depositBalance.AmountOf(denomB).Sub(sdk.NewInt(200*1000000)), depositorBalanceDenomB.Amount)
	require.Equal(t, depositorBalancePoolCoin.Amount, sdk.OneInt())
	require.Equal(t, poolCoin.Sub(poolCoinBefore), depositorBalancePoolCoin.Amount)
}

func TestDepositDecimalTruncation(t *testing.T) {
	simapp, ctx := createTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)
	params.WithdrawFeeRate = sdk.ZeroDec()

	// Suppose that atom price is $40.
	denomA := "uatom"
	denomB := "uusd"

	addrs := app.AddTestAddrs(simapp, ctx, 2, sdk.NewCoins())
	creator, depositor := addrs[0], addrs[1]

	// Create a pool with initial value of $10M.
	depositCoins := sdk.NewCoins(sdk.NewInt64Coin(denomA, 125000*1000000), sdk.NewInt64Coin(denomB, 5000000*1000000))
	err := app.FundAccount(simapp, ctx, creator, depositCoins.Add(params.PoolCreationFee...))
	require.NoError(t, err)
	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creator, types.DefaultPoolTypeID, depositCoins))
	require.NoError(t, err)

	// Deposit 19$.
	depositCoins = sdk.NewCoins(sdk.NewInt64Coin(denomA, 0.2375*1000000), sdk.NewInt64Coin(denomB, 9.5*1000000))
	err = app.FundAccount(simapp, ctx, depositor, depositCoins)
	require.NoError(t, err)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(depositor, pool.Id, depositCoins))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	depositMsgStates := simapp.LiquidityKeeper.GetAllDepositMsgStates(ctx)
	require.Len(t, depositMsgStates, 0)

	depositorPoolCoin := simapp.BankKeeper.GetBalance(ctx, depositor, pool.PoolCoinDenom)
	require.True(sdk.IntEq(t, sdk.OneInt(), depositorPoolCoin.Amount))
	require.True(t, simapp.BankKeeper.GetBalance(ctx, depositor, denomA).Amount.IsPositive())
	require.True(t, simapp.BankKeeper.GetBalance(ctx, depositor, denomB).Amount.IsPositive())

	// Withdraw.
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(depositor, pool.Id, depositorPoolCoin))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	withdrawMsgStates := simapp.LiquidityKeeper.GetAllWithdrawMsgStates(ctx)
	require.Len(t, withdrawMsgStates, 0)

	depositorCoinsDelta := depositCoins.Sub(simapp.BankKeeper.GetAllBalances(ctx, depositor))
	require.True(t, depositorCoinsDelta.IsZero())
}

func TestDepositDecimalTruncation2(t *testing.T) {
	simapp, ctx := createTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)
	params.WithdrawFeeRate = sdk.ZeroDec()

	// Suppose that atom price is $40.
	denomA := "uatom"
	denomB := "uusd"

	addrs := app.AddTestAddrs(simapp, ctx, 2, sdk.NewCoins())
	creator, depositor := addrs[0], addrs[1]

	// Create a pool with initial value of $10M.
	depositCoins := sdk.NewCoins(sdk.NewInt64Coin(denomA, 125000*1000000), sdk.NewInt64Coin(denomB, 5000000*1000000))
	err := app.FundAccount(simapp, ctx, creator, depositCoins.Add(params.PoolCreationFee...))
	require.NoError(t, err)
	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creator, types.DefaultPoolTypeID, depositCoins))
	require.NoError(t, err)

	// Deposit 9$.
	depositCoins = sdk.NewCoins(sdk.NewInt64Coin(denomA, 0.1125*1000000), sdk.NewInt64Coin(denomB, 4.5*1000000))
	err = app.FundAccount(simapp, ctx, depositor, depositCoins)
	require.NoError(t, err)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(depositor, pool.Id, depositCoins))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	depositorPoolCoin := simapp.BankKeeper.GetBalance(ctx, depositor, pool.PoolCoinDenom)
	require.True(sdk.IntEq(t, sdk.ZeroInt(), depositorPoolCoin.Amount))
	require.True(t, simapp.BankKeeper.GetAllBalances(ctx, depositor).IsEqual(depositCoins))
	depositMsgStates := simapp.LiquidityKeeper.GetAllDepositMsgStates(ctx)
	require.Len(t, depositMsgStates, 1)
	require.True(t, depositMsgStates[0].Executed)
	require.False(t, depositMsgStates[0].Succeeded)
	require.True(t, depositMsgStates[0].ToBeDeleted)

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	depositMsgStates = simapp.LiquidityKeeper.GetAllDepositMsgStates(ctx)
	require.Len(t, depositMsgStates, 0)
}

func TestReserveCoinLimit(t *testing.T) {
	simapp, ctx := createTestInput()
	params := types.DefaultParams()
	params.MaxReserveCoinAmount = sdk.NewInt(1000000000000)
	simapp.LiquidityKeeper.SetParams(ctx, params)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, params.MaxReserveCoinAmount), sdk.NewCoin(denomB, sdk.NewInt(1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)
	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)
	require.Equal(t, deposit, depositBalance)

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.Equal(t, types.ErrExceededReserveCoinLimit, err)

	params.MaxReserveCoinAmount = sdk.ZeroInt()
	simapp.LiquidityKeeper.SetParams(ctx, params)
	_, err = simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	params.MaxReserveCoinAmount = sdk.NewInt(1000000000000)
	simapp.LiquidityKeeper.SetParams(ctx, params)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	deposit = sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(1000000)), sdk.NewCoin(denomB, sdk.NewInt(1000000)))
	app.SaveAccount(simapp, ctx, addrs[1], deposit)
	depositMsg := types.NewMsgDepositWithinBatch(addrs[1], pool.Id, deposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.Equal(t, types.ErrExceededReserveCoinLimit, err)

	params.MaxReserveCoinAmount = sdk.ZeroInt()
	simapp.LiquidityKeeper.SetParams(ctx, params)

	depositMsg = types.NewMsgDepositWithinBatch(addrs[1], pool.Id, deposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	simapp.LiquidityKeeper.ExecutePoolBatches(ctx)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	simapp.LiquidityKeeper.DeleteAndInitPoolBatches(ctx)
	app.SaveAccount(simapp, ctx, addrs[1], deposit)
	depositMsg = types.NewMsgDepositWithinBatch(addrs[1], pool.Id, deposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	params.MaxReserveCoinAmount = sdk.NewInt(1000000000000)
	simapp.LiquidityKeeper.SetParams(ctx, params)

	poolBatch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs = simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[0], poolBatch)
	require.Equal(t, types.ErrExceededReserveCoinLimit, err)
}

func TestExecuteWithdrawal(t *testing.T) {
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

	// Case for normal withdrawing
	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	require.Equal(t, poolCoinBefore, withdrawerPoolCoinBefore.Amount)
	withdrawMsg := types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, withdrawerPoolCoinBefore.Amount.QuoRaw(2)))

	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinAfter := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	require.Equal(t, poolCoinAfter, poolCoinBefore.QuoRaw(2))
	require.Equal(t, withdrawerPoolCoinAfter.Amount, withdrawerPoolCoinBefore.Amount.QuoRaw(2))
	withdrawerDenomABalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[0]).QuoRaw(2).ToDec().Mul(sdk.OneDec().Sub(params.WithdrawFeeRate)).TruncateInt(), withdrawerDenomABalance.Amount)
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[1]).QuoRaw(2).ToDec().Mul(sdk.OneDec().Sub(params.WithdrawFeeRate)).TruncateInt(), withdrawerDenomBBalance.Amount)

	// Case for withdrawing all reserve coins
	poolCoinBefore = simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore = simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	require.Equal(t, poolCoinBefore, withdrawerPoolCoinBefore.Amount)
	withdrawMsg = types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, poolCoinBefore))

	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
	require.NoError(t, err)

	poolBatch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	require.True(t, found)
	msgs = simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteWithdrawal(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	poolCoinAfter = simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinAfter = simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	require.True(t, true, poolCoinAfter.IsZero())
	require.True(t, true, withdrawerPoolCoinAfter.IsZero())
	withdrawerDenomABalance = simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalance = simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[0]), withdrawerDenomABalance.Amount)
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[1]), withdrawerDenomBBalance.Amount)
}

func TestSmallWithdrawalCase(t *testing.T) {
	simapp, ctx := createTestInput()
	params := types.DefaultParams()
	params.InitPoolCoinMintAmount = sdk.NewInt(1_000000_000000)
	simapp.LiquidityKeeper.SetParams(ctx, params)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(1250001*1000000)), sdk.NewCoin(denomB, sdk.NewInt(9*1000000)))
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

	// Case for normal withdrawing
	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	withdrawerDenomABalanceBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalanceBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])

	require.Equal(t, poolCoinBefore, withdrawerPoolCoinBefore.Amount)
	withdrawMsg := types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(1)))

	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinAfter := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)

	require.Equal(t, poolCoinAfter, poolCoinBefore)
	require.Equal(t, withdrawerPoolCoinAfter.Amount, withdrawerPoolCoinBefore.Amount)
	withdrawerDenomABalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])

	reservePoolBalanceA := simapp.BankKeeper.GetBalance(ctx, pool.GetReserveAccount(), pool.ReserveCoinDenoms[0])
	reservePoolBalanceB := simapp.BankKeeper.GetBalance(ctx, pool.GetReserveAccount(), pool.ReserveCoinDenoms[1])
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[0]), reservePoolBalanceA.Amount)
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[1]), reservePoolBalanceB.Amount)
	require.Equal(t, withdrawerDenomABalanceBefore, withdrawerDenomABalance)
	require.Equal(t, withdrawerDenomBBalanceBefore, withdrawerDenomBBalance)
}

func TestReinitializePool(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)
	params.WithdrawFeeRate = sdk.ZeroDec()
	simapp.LiquidityKeeper.SetParams(ctx, params)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(100*1000000)))
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

	reserveCoins := simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.IsEqual(deposit))

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

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.IsZero())

	// error when swap request to depleted pool
	offerCoin := sdk.NewCoin(denomA, withdrawerDenomABalance.Amount.QuoRaw(2))
	swapMsg := types.NewMsgSwapWithinBatch(addrs[0], pool.Id, types.DefaultSwapTypeID, offerCoin, denomB, sdk.MustNewDecFromStr("0.1"), params.SwapFeeRate)
	_, err = simapp.LiquidityKeeper.SwapWithinBatch(ctx, swapMsg, 0)
	require.ErrorIs(t, err, types.ErrDepletedPool)

	depositMsg := types.NewMsgDepositWithinBatch(addrs[0], pool.Id, deposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	depositMsgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(depositMsgs))

	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, depositMsgs[0], poolBatch)
	require.NoError(t, err)

	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	depositorBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)
	require.Equal(t, poolCoin, depositorBalance.Amount)

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.IsEqual(deposit))
}

func TestReserveAccManipulation(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))

	// depositor, withdrawer
	app.SaveAccount(simapp, ctx, addrs[0], deposit)
	// reserveAccount manipulator
	app.SaveAccount(simapp, ctx, addrs[1], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	// reserveAcc manipulation coinA
	manipulationReserveA1 := sdk.NewCoin(denomA, sdk.NewInt(30*1000000))
	manipulationReserveA2 := sdk.NewCoin(denomA, sdk.NewInt(70*1000000))
	// reserveAcc manipulation coin other than reserve coins
	manipulationReserveOther := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100*1000000))
	// manipulated reserve coinA
	addedDepositA := depositA.Add(manipulationReserveA1).Add(manipulationReserveA2)

	createMsg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, createMsg)
	require.NoError(t, err)
	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	reserveAcc := pool.GetReserveAccount()
	reserveAccBalances := simapp.BankKeeper.GetAllBalances(ctx, reserveAcc)
	require.Equal(t, reserveAccBalances, sdk.NewCoins(depositA, depositB))

	// send coin to manipulate reserve account
	err = simapp.BankKeeper.SendCoins(ctx, addrs[1], reserveAcc, sdk.NewCoins(manipulationReserveA1))
	require.NoError(t, err)
	metadata := simapp.LiquidityKeeper.GetPoolMetaData(ctx, pool)
	require.Equal(t, depositA.Add(manipulationReserveA1).Amount, metadata.ReserveCoins.AmountOf(denomA))

	_ = simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)
	withdrawMsg := types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, withdrawerPoolCoinBefore.Amount.QuoRaw(2)))
	_, _ = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)

	poolBatch, _ := simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	_ = simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	// send coin to manipulate reserve account
	_ = simapp.BankKeeper.SendCoins(ctx, addrs[1], reserveAcc, sdk.NewCoins(manipulationReserveA2))
	_ = simapp.BankKeeper.SendCoins(ctx, addrs[1], reserveAcc, sdk.NewCoins(manipulationReserveOther))
	_ = simapp.BankKeeper.GetAllBalances(ctx, reserveAcc)
	metadata = simapp.LiquidityKeeper.GetPoolMetaData(ctx, pool)
	require.NotEqual(t, manipulationReserveOther, metadata.ReserveCoins.AmountOf(sdk.DefaultBondDenom))

	// Case for withdrawing all reserve coins after manipulation
	poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	withdrawerPoolCoinBefore = simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.PoolCoinDenom)
	withdrawMsg = types.NewMsgWithdrawWithinBatch(addrs[0], pool.Id, sdk.NewCoin(pool.PoolCoinDenom, poolCoinBefore))

	_, _ = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
	require.NoError(t, err)

	poolBatch, _ = simapp.LiquidityKeeper.GetPoolBatch(ctx, withdrawMsg.PoolId)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)

	err = simapp.LiquidityKeeper.ExecuteWithdrawal(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	withdrawerDenomABalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[0])
	withdrawerDenomBBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pool.ReserveCoinDenoms[1])
	withdrawerDenomOtherBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], sdk.DefaultBondDenom)
	require.Equal(t, addedDepositA, withdrawerDenomABalance)
	require.Equal(t, deposit.AmountOf(pool.ReserveCoinDenoms[1]), withdrawerDenomBBalance.Amount)
	require.NotEqual(t, manipulationReserveOther, withdrawerDenomOtherBalance)
}

func TestGetLiquidityPoolMetadata(t *testing.T) {
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

func TestIsPoolCoinDenom(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "denomA"
	denomB := "denomB"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	deposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(100*1000000)), sdk.NewCoin(denomB, sdk.NewInt(2000*1000000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit)

	depositA := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomA)
	depositB := simapp.BankKeeper.GetBalance(ctx, addrs[0], denomB)
	depositBalance := sdk.NewCoins(depositA, depositB)

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, depositBalance)

	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)
	getPool, found := simapp.LiquidityKeeper.GetPool(ctx, pool.Id)
	require.True(t, found)
	require.Equal(t, pool, getPool)

	require.Equal(t, "denomA/denomB/1", pool.Name())
	poolCoinDenom := types.GetPoolCoinDenom(pool.Name())
	require.Equal(t, pool.PoolCoinDenom, poolCoinDenom)
	require.True(t, simapp.LiquidityKeeper.IsPoolCoinDenom(ctx, pool.PoolCoinDenom))
	require.False(t, simapp.LiquidityKeeper.IsPoolCoinDenom(ctx, pool.Name()))
}

func TestGetPoolByReserveAccIndex(t *testing.T) {
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
	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	fmt.Println(pool)
	poolStored, found := simapp.LiquidityKeeper.GetPool(ctx, pool.Id)
	require.True(t, found)
	require.Equal(t, pool, poolStored)
	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	require.Equal(t, pool, pools[0])

	poolByReserveAcc, found := simapp.LiquidityKeeper.GetPoolByReserveAccIndex(ctx, pool.GetReserveAccount())
	require.True(t, found)
	require.Equal(t, pool, poolByReserveAcc)

	poolCoinDenom := types.GetPoolCoinDenom(pool.Name())
	require.Equal(t, pool.PoolCoinDenom, poolCoinDenom)
	require.True(t, simapp.LiquidityKeeper.IsPoolCoinDenom(ctx, pool.PoolCoinDenom))
	require.False(t, simapp.LiquidityKeeper.IsPoolCoinDenom(ctx, pool.Name()))
	// SetPoolByReserveAccIndex
}

func TestDepositWithdrawEdgecase(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		r := rand.New(rand.NewSource(seed))

		simapp, ctx := createTestInput()
		params := simapp.LiquidityKeeper.GetParams(ctx)

		X := params.MinInitDepositAmount.Add(app.GetRandRange(r, 0, 1_000_000))
		Y := params.MinInitDepositAmount.Add(app.GetRandRange(r, 0, 1_000_000))

		creatorCoins := sdk.NewCoins(sdk.NewCoin(DenomX, X), sdk.NewCoin(DenomY, Y))
		creatorAddr := app.AddRandomTestAddr(simapp, ctx, creatorCoins.Add(params.PoolCreationFee...))

		pool, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creatorAddr, types.DefaultPoolTypeID, creatorCoins))
		require.NoError(t, err)

		for i := 0; i < 500; i++ {
			liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
			type action int
			const (
				deposit action = iota + 1
				withdraw
			)
			actions := []action{}
			balanceX := simapp.BankKeeper.GetBalance(ctx, creatorAddr, DenomX)
			balanceY := simapp.BankKeeper.GetBalance(ctx, creatorAddr, DenomY)
			balancePoolCoin := simapp.BankKeeper.GetBalance(ctx, creatorAddr, pool.PoolCoinDenom)
			if balanceX.IsPositive() || balanceY.IsPositive() {
				actions = append(actions, deposit)
			}
			if balancePoolCoin.Amount.GT(sdk.OneInt()) {
				actions = append(actions, withdraw)
			}
			require.Positive(t, len(actions))
			switch actions[r.Intn(len(actions))] {
			case deposit:
				depositAmtA := sdk.OneInt().Add(sdk.NewInt(r.Int63n(balanceX.Amount.Int64())))
				depositAmtB := sdk.OneInt().Add(sdk.NewInt(r.Int63n(balanceY.Amount.Int64())))
				depositCoins := sdk.NewCoins(sdk.NewCoin(DenomX, depositAmtA), sdk.NewCoin(DenomY, depositAmtB))
				_, err := simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(
					creatorAddr, pool.Id, depositCoins))
				require.NoError(t, err)
			case withdraw:
				totalPoolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
				withdrawAmt := sdk.OneInt().Add(sdk.NewInt(r.Int63n(balancePoolCoin.Amount.Int64())))
				withdrawCoin := sdk.NewCoin(pool.PoolCoinDenom, sdk.MinInt(totalPoolCoin.Sub(sdk.OneInt()), withdrawAmt))
				_, err := simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(
					creatorAddr, pool.Id, withdrawCoin))
				require.NoError(t, err)
			}

			liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
		}
	}
}

func TestWithdrawEdgecase(t *testing.T) {
	simapp, ctx := createTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)

	X, Y := sdk.NewInt(1_000_000), sdk.NewInt(10_000_000)

	depositCoins := sdk.NewCoins(sdk.NewCoin(DenomX, X), sdk.NewCoin(DenomY, Y))
	creatorAddr := app.AddRandomTestAddr(simapp, ctx, depositCoins.Add(params.PoolCreationFee...))

	pool, err := simapp.LiquidityKeeper.CreatePool(ctx, types.NewMsgCreatePool(creatorAddr, types.DefaultPoolTypeID, depositCoins))
	require.NoError(t, err)

	creatorBalance := simapp.BankKeeper.GetBalance(ctx, creatorAddr, pool.PoolCoinDenom).Sub(sdk.NewCoin(pool.PoolCoinDenom, sdk.NewInt(2)))

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, creatorBalance))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	fmt.Println(simapp.LiquidityKeeper.GetPoolCoinTotal(ctx, pool))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, creatorAddr))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, pool.GetReserveAccount()))

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewCoin(pool.PoolCoinDenom, sdk.OneInt())))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	fmt.Println(simapp.LiquidityKeeper.GetPoolCoinTotal(ctx, pool))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, creatorAddr))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, pool.GetReserveAccount()))

	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewCoin(pool.PoolCoinDenom, sdk.OneInt())))
	require.NoError(t, err)

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	fmt.Println(simapp.LiquidityKeeper.GetPoolCoinTotal(ctx, pool))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, creatorAddr))
	fmt.Println(simapp.BankKeeper.GetAllBalances(ctx, pool.GetReserveAccount()))
}

func TestWithdrawEdgecase2(t *testing.T) {
	simapp, ctx, pool, creatorAddr, err := createTestPool(sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 1500000))
	require.NoError(t, err)

	for i := 0; i < 1002; i++ {
		liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
		_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 998)))
		require.NoError(t, err)
		liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	}

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 1)))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
}

func TestWithdrawSmallAmount(t *testing.T) {
	simapp, ctx, pool, creatorAddr, err := createTestPool(sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 1500000))
	require.NoError(t, err)

	require.NotPanics(t, func() {
		liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
		_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 1)))
		require.NoError(t, err)
		liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	})
}

func TestGetReserveCoins(t *testing.T) {
	simapp, ctx, pool, creatorAddr, err := createTestPool(sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 1000000))
	require.NoError(t, err)

	reserveCoins := simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.Len(t, reserveCoins, 2)
	require.True(t, reserveCoins.AmountOf(DenomX).Equal(sdk.NewInt(1000000)))
	require.True(t, reserveCoins.AmountOf(DenomY).Equal(sdk.NewInt(1000000)))

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, sdk.NewInt64Coin(pool.PoolCoinDenom, 1000000)))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.Len(t, reserveCoins, 2)
	require.True(t, reserveCoins.AmountOf(DenomX).IsZero())
	require.True(t, reserveCoins.AmountOf(DenomY).IsZero())
}

func TestDepositToDepletedPool(t *testing.T) {
	simapp, ctx, pool, creatorAddr, err := createTestPool(sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 1000000))
	require.NoError(t, err)
	params := simapp.LiquidityKeeper.GetParams(ctx)

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	pc := simapp.BankKeeper.GetBalance(ctx, creatorAddr, pool.PoolCoinDenom)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(creatorAddr, pool.Id, pc))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	reserveCoins := simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.AmountOf(DenomX).IsZero())
	require.True(t, reserveCoins.AmountOf(DenomY).IsZero())

	require.True(t, simapp.LiquidityKeeper.IsDepletedPool(ctx, pool))

	err = simapp.BankKeeper.SendCoins(ctx, creatorAddr, pool.GetReserveAccount(), sdk.NewCoins(sdk.NewInt64Coin(DenomX, 10000)))
	require.NoError(t, err)

	// Deposit request must be rejected since the pool is depleted and
	// depositing coins amount is smaller than MinInitDepositAmount.
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	depositCoins := sdk.NewCoins(sdk.NewInt64Coin(DenomX, 10000), sdk.NewInt64Coin(DenomY, 10000))
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(creatorAddr, pool.Id, depositCoins))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.AmountOf(DenomX).Equal(sdk.NewInt(10000)))
	require.True(t, reserveCoins.AmountOf(DenomY).IsZero())
	creatorCoins := simapp.BankKeeper.GetAllBalances(ctx, creatorAddr)
	require.True(t, creatorCoins.AmountOf(DenomX).Equal(sdk.NewInt(990000)))
	require.True(t, creatorCoins.AmountOf(DenomY).Equal(sdk.NewInt(1000000)))

	// This time the request will be accepted since depositCoins + reserveCoins > MinInitDepositAmount.
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	depositCoins = sdk.NewCoins(sdk.NewInt64Coin(DenomX, 990000), sdk.NewInt64Coin(DenomY, 1000000))
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(creatorAddr, pool.Id, depositCoins))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(t, reserveCoins.AmountOf(DenomX).Equal(sdk.NewInt(1000000)))
	require.True(t, reserveCoins.AmountOf(DenomY).Equal(sdk.NewInt(1000000)))
	creatorCoins = simapp.BankKeeper.GetAllBalances(ctx, creatorAddr)
	require.True(t, creatorCoins.AmountOf(DenomX).Equal(sdk.NewInt(0)))
	require.True(t, creatorCoins.AmountOf(DenomY).Equal(sdk.NewInt(0)))
	require.True(t, creatorCoins.AmountOf(pool.PoolCoinDenom).Equal(params.InitPoolCoinMintAmount))
}

func TestDepositWithCoinsSent(t *testing.T) {
	simapp, ctx, pool, _, err := createTestPool(sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 1000000))
	require.NoError(t, err)

	// Send extra coins to the pool reserve account, which causes the pool price to change.
	// Any other coins(coins with denom "denomZ" here) than pool's reserve coins will not have any effect.
	extraCoins := sdk.NewCoins(
		sdk.NewInt64Coin(DenomX, 1000000), sdk.NewInt64Coin(DenomY, 2000000), sdk.NewInt64Coin("denomZ", 1000000))
	addr := app.AddRandomTestAddr(simapp, ctx, extraCoins)
	err = simapp.BankKeeper.SendCoins(ctx, addr, pool.GetReserveAccount(), extraCoins)
	require.NoError(t, err)
	reserveCoins := simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.Len(t, reserveCoins, 2) // denomZ coins are ignored
	require.True(sdk.IntEq(t, sdk.NewInt(2000000), reserveCoins.AmountOf(DenomX)))
	require.True(sdk.IntEq(t, sdk.NewInt(3000000), reserveCoins.AmountOf(DenomY)))

	// Add more coins to deposit.
	depositCoins := sdk.NewCoins(sdk.NewInt64Coin(DenomX, 3000000), sdk.NewInt64Coin(DenomY, 3000000))
	app.SaveAccount(simapp, ctx, addr, depositCoins)

	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, types.NewMsgDepositWithinBatch(addr, pool.Id, depositCoins))
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

	reserveCoins = simapp.LiquidityKeeper.GetReserveCoins(ctx, pool)
	require.True(sdk.IntEq(t, sdk.NewInt(4000000), reserveCoins.AmountOf(DenomX)))
	require.True(sdk.IntEq(t, sdk.NewInt(6000000), reserveCoins.AmountOf(DenomY)))
	balances := simapp.BankKeeper.GetAllBalances(ctx, addr)
	require.True(sdk.IntEq(t, sdk.NewInt(1000000), balances.AmountOf(DenomX)))
	require.True(sdk.IntEq(t, sdk.NewInt(0), balances.AmountOf(DenomY)))
	require.True(sdk.IntEq(t, sdk.NewInt(1000000), balances.AmountOf(pool.PoolCoinDenom)))
}

func TestCreatePoolEqualDenom(t *testing.T) {
	simapp, ctx := createTestInput()
	params := types.DefaultParams()
	simapp.LiquidityKeeper.SetParams(ctx, params)
	addrs := app.AddTestAddrs(simapp, ctx, 1, params.PoolCreationFee)

	msg := types.NewMsgCreatePool(addrs[0], types.DefaultPoolTypeID,
		sdk.Coins{
			sdk.NewCoin(DenomA, sdk.NewInt(1000000)),
			sdk.NewCoin(DenomA, sdk.NewInt(1000000)),
		})
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.ErrorIs(t, err, types.ErrEqualDenom)
}

func TestOverflowAndZeroCases(t *testing.T) {
	simapp, ctx := createTestInput()
	params := types.DefaultParams()
	simapp.LiquidityKeeper.SetParams(ctx, params)
	keeper.BatchLogicInvariantCheckFlag = false

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	// Check overflow case on deposit
	deposit := sdk.NewCoins(
		sdk.NewCoin(denomA, sdk.NewInt(1_000_000)),
		sdk.NewCoin(denomB, sdk.NewInt(2_000_000_000_000*1_000_000).MulRaw(1_000_000)))
	hugeCoins := sdk.NewCoins(
		sdk.NewCoin(denomA, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000)),
		sdk.NewCoin(denomB, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000)))
	hugeCoins2 := sdk.NewCoins(
		sdk.NewCoin(denomA, sdk.NewInt(1_000_000_000_000_000_000)),
		sdk.NewCoin(denomB, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000)))
	app.SaveAccount(simapp, ctx, addrs[0], deposit.Add(hugeCoins.Add(hugeCoins2...)...))

	msg := types.NewMsgCreatePool(addrs[0], poolTypeID, deposit)
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)
	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pools[0])

	depositorBalance := simapp.BankKeeper.GetAllBalances(ctx, addrs[0])
	depositMsg := types.NewMsgDepositWithinBatch(addrs[0], pools[0].Id, hugeCoins)
	depositMsg2 := types.NewMsgDepositWithinBatch(addrs[0], pools[0].Id, hugeCoins2)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg2)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 2, len(msgs))
	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[0], poolBatch)
	require.ErrorIs(t, err, types.ErrOverflowAmount)
	err = simapp.LiquidityKeeper.RefundDeposit(ctx, msgs[0], poolBatch)
	require.NoError(t, err)
	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[1], poolBatch)
	require.ErrorIs(t, err, types.ErrOverflowAmount)
	err = simapp.LiquidityKeeper.RefundDeposit(ctx, msgs[1], poolBatch)
	require.NoError(t, err)

	poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pools[0])
	depositorPoolCoinBalance := simapp.BankKeeper.GetBalance(ctx, addrs[0], pools[0].PoolCoinDenom)
	require.Equal(t, poolCoin, poolCoinAfter)
	require.Equal(t, poolCoinAfter, depositorPoolCoinBalance.Amount)
	require.Equal(t, depositorBalance.AmountOf(pools[0].PoolCoinDenom), depositorPoolCoinBalance.Amount)

	hugeCoins3 := sdk.NewCoins(
		sdk.NewCoin(denomA, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000)),
		sdk.NewCoin(denomB, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000)))
	depositMsg = types.NewMsgDepositWithinBatch(addrs[0], pools[0].Id, hugeCoins3)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)
	msgs = simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 3, len(msgs))
	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[2], poolBatch)
	require.NoError(t, err)

	// Check overflow case on withdraw
	depositorPoolCoinBalance = simapp.BankKeeper.GetBalance(ctx, addrs[0], pools[0].PoolCoinDenom)
	_, err = simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(addrs[0], pools[0].Id, depositorPoolCoinBalance.SubAmount(sdk.NewInt(1))))
	require.NoError(t, err)

	poolBatch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	withdrawMsgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 1, len(withdrawMsgs))
	err = simapp.LiquidityKeeper.ExecuteWithdrawal(ctx, withdrawMsgs[0], poolBatch)
	require.ErrorIs(t, err, types.ErrOverflowAmount)
	err = simapp.LiquidityKeeper.RefundWithdrawal(ctx, withdrawMsgs[0], poolBatch)
	require.NoError(t, err)

	// Check overflow, division by zero case on swap
	swapUserBalanceBefore := simapp.BankKeeper.GetAllBalances(ctx, addrs[0])
	offerCoinA := sdk.NewCoin(denomA, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000))
	orderPriceA := sdk.MustNewDecFromStr("110000000000000000000000000000000000000000000000000000000000.000000000000000001")
	offerCoinB := sdk.NewCoin(denomB, sdk.NewInt(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000_000_000).MulRaw(1_000_000_000_000))
	orderPriceB := sdk.MustNewDecFromStr("0.000000000000000001")
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	_, err = simapp.LiquidityKeeper.SwapWithinBatch(
		ctx,
		types.NewMsgSwapWithinBatch(addrs[0], pools[0].Id, types.DefaultSwapTypeID, offerCoinA, denomB, orderPriceA, params.SwapFeeRate),
		0)
	require.ErrorIs(t, err, types.ErrOverflowAmount)
	_, err = simapp.LiquidityKeeper.SwapWithinBatch(
		ctx,
		types.NewMsgSwapWithinBatch(addrs[0], pools[0].Id, types.DefaultSwapTypeID, offerCoinB, denomA, orderPriceB, params.SwapFeeRate),
		0)
	require.NoError(t, err)
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)
	swapUserBalanceAfter := simapp.BankKeeper.GetAllBalances(ctx, addrs[0])
	require.Equal(t, swapUserBalanceBefore, swapUserBalanceAfter)
	depositMsgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 0, len(depositMsgs))
	withdrawMsgs = simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, poolBatch)
	require.Equal(t, 0, len(withdrawMsgs))
	swapMsgs := simapp.LiquidityKeeper.GetAllPoolBatchSwapMsgStates(ctx, poolBatch)
	require.Equal(t, 0, len(swapMsgs))
}

func TestExecuteBigDeposit(t *testing.T) {
	simapp, ctx := createTestInput()
	simapp.LiquidityKeeper.SetParams(ctx, types.DefaultParams())
	params := simapp.LiquidityKeeper.GetParams(ctx)
	keeper.BatchLogicInvariantCheckFlag = false

	poolTypeID := types.DefaultPoolTypeID
	addrs := app.AddTestAddrs(simapp, ctx, 3, params.PoolCreationFee)

	denomA := "uETH"
	denomB := "uUSD"
	denomA, denomB = types.AlphabeticalDenomPair(denomA, denomB)

	// 2^63-1
	hugeInt := int64(9223372036854775807)
	initDeposit := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(hugeInt)), sdk.NewCoin(denomB, sdk.NewInt(hugeInt)))
	app.SaveAccount(simapp, ctx, addrs[0], initDeposit)
	app.SaveAccount(simapp, ctx, addrs[1], initDeposit)
	app.SaveAccount(simapp, ctx, addrs[2], initDeposit)

	createBalance := sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewInt(1*1000000)), sdk.NewCoin(denomB, sdk.NewInt(1*1000000)))

	createMsg := types.NewMsgCreatePool(addrs[0], poolTypeID, createBalance)

	_, err := simapp.LiquidityKeeper.CreatePool(ctx, createMsg)
	require.NoError(t, err)

	pools := simapp.LiquidityKeeper.GetAllPools(ctx)
	pool := pools[0]

	poolCoinInit := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	require.Equal(t, poolCoinInit, sdk.NewInt(1*1000000))

	depositMsg := types.NewMsgDepositWithinBatch(addrs[1], pool.Id, initDeposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	poolBatch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	require.Equal(t, poolCoin.Sub(poolCoinInit), simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom).Amount)

	simapp.LiquidityKeeper.DeleteAllReadyPoolBatchDepositMsgStates(ctx, poolBatch)

	depositMsg = types.NewMsgDepositWithinBatch(addrs[2], pool.Id, initDeposit)
	_, err = simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
	require.NoError(t, err)

	poolBatch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, depositMsg.PoolId)
	require.True(t, found)
	msgs = simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, poolBatch)
	require.Equal(t, 1, len(msgs))

	err = simapp.LiquidityKeeper.ExecuteDeposit(ctx, msgs[0], poolBatch)
	require.NoError(t, err)

	poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	require.Equal(t, poolCoinAfter.Sub(poolCoin), simapp.BankKeeper.GetBalance(ctx, addrs[2], pool.PoolCoinDenom).Amount)
	require.Equal(t, simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom).Amount, simapp.BankKeeper.GetBalance(ctx, addrs[2], pool.PoolCoinDenom).Amount)

	require.True(t, simapp.BankKeeper.GetBalance(ctx, addrs[1], denomA).IsZero())
	require.True(t, simapp.BankKeeper.GetBalance(ctx, addrs[1], denomB).IsZero())

	// Error due to decimal operation exceeding precision
	require.Equal(t, sdk.NewInt(8), simapp.BankKeeper.GetBalance(ctx, addrs[2], denomA).Amount)
	require.Equal(t, sdk.NewInt(8), simapp.BankKeeper.GetBalance(ctx, addrs[2], denomB).Amount)

	poolCoinAmt := simapp.BankKeeper.GetBalance(ctx, addrs[1], pool.PoolCoinDenom)
	state, err := simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, types.NewMsgWithdrawWithinBatch(addrs[1], pool.Id, poolCoinAmt))
	require.NoError(t, err)

	err = simapp.LiquidityKeeper.ExecuteWithdrawal(ctx, state, poolBatch)
	require.NoError(t, err)

	balanceAfter := simapp.BankKeeper.GetAllBalances(ctx, addrs[1])
	liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	// Error due to decimal operation exceeding precision
	require.Equal(t, sdk.ZeroInt(), balanceAfter.AmountOf(pool.PoolCoinDenom))
	require.Equal(t, sdk.NewInt(-4), balanceAfter.AmountOf(denomA).SubRaw(hugeInt))
	require.Equal(t, sdk.NewInt(-4), balanceAfter.AmountOf(denomB).SubRaw(hugeInt))
}
