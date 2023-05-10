package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	app "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity/keeper"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

const custom = "custom"

func getQueriedLiquidityPool(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, poolID uint64) (types.Pool, error) {
	t.Helper()
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPool}, "/"),
		Data: cdc.MustMarshalJSON(types.QueryLiquidityPoolParams{PoolId: poolID}),
	}

	pool := types.Pool{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPool}, query)
	if err != nil {
		return pool, err
	}
	require.Nil(t, cdc.UnmarshalJSON(bz, &pool))
	return pool, nil
}

func getQueriedLiquidityPools(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) (types.Pools, error) {
	t.Helper()
	queryDelParams := types.NewQueryLiquidityPoolsParams(1, 100)
	bz, errRes := cdc.MarshalJSON(queryDelParams)
	fmt.Println(bz, errRes)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPools}, "/"),
		Data: bz,
	}

	pools := types.Pools{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPools}, query)
	if err != nil {
		return pools, err
	}
	require.Nil(t, cdc.UnmarshalJSON(bz, &pools))
	return pools, nil
}

func TestNewQuerier(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	simapp := app.Setup(false)
	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})
	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))

	querier := keeper.NewQuerier(simapp.LiquidityKeeper, cdc)

	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, DenomX, DenomY, addrs[0])
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryLiquidityPool}, "/"),
		Data: cdc.MustMarshalJSON(types.QueryLiquidityPoolParams{PoolId: poolID}),
	}
	queryFailCase := abci.RequestQuery{
		Path: strings.Join([]string{"failCustom", "failRoute", "failQuery"}, "/"),
		Data: cdc.MustMarshalJSON(types.Pool{}),
	}
	pool := types.Pool{}
	bz, err := querier(ctx, []string{types.QueryLiquidityPool}, query)
	require.NoError(t, err)
	require.Nil(t, cdc.UnmarshalJSON(bz, &pool))

	bz, err = querier(ctx, []string{"fail"}, queryFailCase)
	require.EqualError(t, err, "unknown query path of liquidity module: fail: unknown request")
	err = cdc.UnmarshalJSON(bz, &pool)
	require.EqualError(t, err, "UnmarshalJSON cannot decode empty bytes")
}

func TestQueries(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)

	simapp := app.Setup(false)
	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair(DenomX, DenomY)
	// denoms := []string{denomX, denomY}

	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))

	poolID := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])
	poolID2 := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, "testDenom", addrs[0])
	require.Equal(t, uint64(1), poolID)
	require.Equal(t, uint64(2), poolID2)

	// begin block, init
	app.TestDepositPool(t, simapp, ctx, X, Y, addrs[1:10], poolID, true)

	querier := keeper.NewQuerier(simapp.LiquidityKeeper, cdc)

	require.Equal(t, uint64(1), poolID)
	poolRes, err := getQueriedLiquidityPool(t, ctx, cdc, querier, poolID)
	require.NoError(t, err)
	require.Equal(t, poolID, poolRes.Id)
	require.Equal(t, types.DefaultPoolTypeID, poolRes.TypeId)
	require.Equal(t, []string{DenomX, DenomY}, poolRes.ReserveCoinDenoms)
	require.NotNil(t, poolRes.PoolCoinDenom)
	require.NotNil(t, poolRes.ReserveAccountAddress)

	poolResEmpty, err := getQueriedLiquidityPool(t, ctx, cdc, querier, uint64(3))
	require.ErrorIs(t, err, types.ErrPoolNotExists)
	require.Equal(t, uint64(0), poolResEmpty.Id)

	poolsRes, err := getQueriedLiquidityPools(t, ctx, cdc, querier)
	require.NoError(t, err)
	require.Equal(t, 2, len(poolsRes))
}
