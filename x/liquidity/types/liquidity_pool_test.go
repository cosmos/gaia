package types_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestUnmarshalerPanics(t *testing.T) {
	t.Run("MustUnmarshalPool", func(t *testing.T) {
		require.Panics(t, func() {
			types.MustUnmarshalPool(types.ModuleCdc, []byte{0x00})
		})
	})
	t.Run("MustUnmarshalPoolBatch", func(t *testing.T) {
		require.Panics(t, func() {
			types.MustUnmarshalPoolBatch(types.ModuleCdc, []byte{0x00})
		})
	})
	t.Run("MustUnmarshalDepositMsgState", func(t *testing.T) {
		require.Panics(t, func() {
			types.MustUnmarshalDepositMsgState(types.ModuleCdc, []byte{0x00})
		})
	})
	t.Run("MustUnmarshalWithdrawMsgState", func(t *testing.T) {
		require.Panics(t, func() {
			types.MustUnmarshalWithdrawMsgState(types.ModuleCdc, []byte{0x00})
		})
	})
	t.Run("MustUnmarshalSwapMsgState", func(t *testing.T) {
		require.Panics(t, func() {
			types.MustUnmarshalSwapMsgState(types.ModuleCdc, []byte{0x00})
		})
	})
}

func TestLiquidityPoolBatch(t *testing.T) {
	simapp, ctx := app.CreateTestInput()
	params := simapp.LiquidityKeeper.GetParams(ctx)

	pool := types.Pool{}
	require.Equal(t, types.ErrPoolNotExists, pool.Validate())

	pool.Id = 1
	require.Equal(t, types.ErrPoolTypeNotExists, pool.Validate())

	pool.TypeId = 1
	require.Equal(t, types.ErrNumOfReserveCoinDenoms, pool.Validate())

	pool.ReserveCoinDenoms = []string{DenomY, DenomX, DenomX}
	require.Equal(t, types.ErrNumOfReserveCoinDenoms, pool.Validate())

	pool.ReserveCoinDenoms = []string{DenomY, DenomX}
	require.Equal(t, types.ErrBadOrderingReserveCoinDenoms, pool.Validate())

	pool.ReserveCoinDenoms = []string{DenomX, DenomY}
	require.Equal(t, types.ErrEmptyReserveAccountAddress, pool.Validate())

	pool.ReserveAccountAddress = "badaddress"
	require.Equal(t, types.ErrBadReserveAccountAddress, pool.Validate())

	pool.ReserveAccountAddress = types.GetPoolReserveAcc(pool.Name(), false).String()
	add2, err := sdk.AccAddressFromBech32(pool.ReserveAccountAddress)
	require.NoError(t, err)
	require.Equal(t, add2, pool.GetReserveAccount())
	require.Equal(t, types.ErrEmptyPoolCoinDenom, pool.Validate())

	pool.PoolCoinDenom = "badPoolCoinDenom"
	require.Equal(t, types.ErrBadPoolCoinDenom, pool.Validate())

	pool.PoolCoinDenom = pool.Name()
	require.NoError(t, pool.Validate())
	require.Equal(t, pool.Name(), types.PoolName(pool.ReserveCoinDenoms, pool.TypeId))
	require.Equal(t, pool.Id, pool.GetId())
	require.Equal(t, pool.PoolCoinDenom, pool.GetPoolCoinDenom())

	cdc := simapp.AppCodec()
	poolByte := types.MustMarshalPool(cdc, pool)
	require.Equal(t, pool, types.MustUnmarshalPool(cdc, poolByte))

	poolByte = types.MustMarshalPool(cdc, pool)
	poolMarshaled, err := types.UnmarshalPool(cdc, poolByte)
	require.NoError(t, err)
	require.Equal(t, pool, poolMarshaled)

	addr, err := sdk.AccAddressFromBech32(pool.ReserveAccountAddress)
	require.NoError(t, err)
	require.True(t, pool.GetReserveAccount().Equals(addr))
	require.Equal(t, strings.TrimSpace(pool.String()+"\n"+pool.String()), types.Pools{pool, pool}.String())

	simapp.LiquidityKeeper.SetPool(ctx, pool)
	batch := types.NewPoolBatch(pool.Id, 1)
	simapp.LiquidityKeeper.SetPoolBatch(ctx, batch)
	batchByte := types.MustMarshalPoolBatch(cdc, batch)
	require.Equal(t, batch, types.MustUnmarshalPoolBatch(cdc, batchByte))

	batchMarshaled, err := types.UnmarshalPoolBatch(cdc, batchByte)
	require.NoError(t, err)
	require.Equal(t, batch, batchMarshaled)

	batchDepositMsg := types.DepositMsgState{}
	batchWithdrawMsg := types.WithdrawMsgState{}
	batchSwapMsg := types.SwapMsgState{
		ExchangedOfferCoin:   sdk.NewCoin("test", sdk.NewInt(1000)),
		RemainingOfferCoin:   sdk.NewCoin("test", sdk.NewInt(1000)),
		ReservedOfferCoinFee: types.GetOfferCoinFee(sdk.NewCoin("test", sdk.NewInt(2000)), params.SwapFeeRate),
	}
	b := types.MustMarshalDepositMsgState(cdc, batchDepositMsg)
	require.Equal(t, batchDepositMsg, types.MustUnmarshalDepositMsgState(cdc, b))

	marshaled, err := types.UnmarshalDepositMsgState(cdc, b)
	require.NoError(t, err)
	require.Equal(t, batchDepositMsg, marshaled)

	b = types.MustMarshalWithdrawMsgState(cdc, batchWithdrawMsg)
	require.Equal(t, batchWithdrawMsg, types.MustUnmarshalWithdrawMsgState(cdc, b))

	withdrawMsgMarshaled, err := types.UnmarshalWithdrawMsgState(cdc, b)
	require.NoError(t, err)
	require.Equal(t, batchWithdrawMsg, withdrawMsgMarshaled)

	b = types.MustMarshalSwapMsgState(cdc, batchSwapMsg)
	require.Equal(t, batchSwapMsg, types.MustUnmarshalSwapMsgState(cdc, b))

	SwapMsgMarshaled, err := types.UnmarshalSwapMsgState(cdc, b)
	require.NoError(t, err)
	require.Equal(t, batchSwapMsg, SwapMsgMarshaled)
}
