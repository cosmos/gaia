package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

const (
	DefaultPoolTypeId = uint32(1)
	DefaultPoolId     = uint64(1)
	DefaultSwapTypeId = uint32(1)
	DenomX            = "denomX"
	DenomY            = "denomY"
)

func TestMsgCreatePool(t *testing.T) {
	poolCreator := sdk.AccAddress(crypto.AddressHash([]byte("testAccount")))

	cases := []struct {
		expectedErr string // empty means no error expected
		msg         *types.MsgCreatePool
	}{
		{
			"",
			types.NewMsgCreatePool(poolCreator, DefaultPoolTypeId, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid index of the pool type",
			types.NewMsgCreatePool(poolCreator, 0, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid pool creator address",
			types.NewMsgCreatePool(sdk.AccAddress{}, DefaultPoolTypeId, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid number of reserve coin",
			types.NewMsgCreatePool(poolCreator, DefaultPoolTypeId, sdk.NewCoins(sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid number of reserve coin",
			types.NewMsgCreatePool(poolCreator, DefaultPoolTypeId, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)), sdk.NewCoin("denomZ", sdk.NewInt(1000)))),
		},
	}

	for _, tc := range cases {
		require.IsType(t, &types.MsgCreatePool{}, tc.msg)
		require.Equal(t, types.TypeMsgCreatePool, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetPoolCreator(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgDepositWithinBatch(t *testing.T) {
	depositor := sdk.AccAddress(crypto.AddressHash([]byte("testAccount")))

	cases := []struct {
		expectedErr string // empty means no error expected
		msg         *types.MsgDepositWithinBatch
	}{
		{
			"",
			types.NewMsgDepositWithinBatch(depositor, DefaultPoolId, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"",
			types.NewMsgDepositWithinBatch(depositor, 0, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid pool depositor address",
			types.NewMsgDepositWithinBatch(sdk.AccAddress{}, DefaultPoolId, sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(1000)), sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
		{
			"invalid number of reserve coin",
			types.NewMsgDepositWithinBatch(depositor, DefaultPoolId, sdk.NewCoins(sdk.NewCoin(DenomY, sdk.NewInt(1000)))),
		},
	}

	for _, tc := range cases {
		require.IsType(t, &types.MsgDepositWithinBatch{}, tc.msg)
		require.Equal(t, types.TypeMsgDepositWithinBatch, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetDepositor(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgWithdrawWithinBatch(t *testing.T) {
	withdrawer := sdk.AccAddress(crypto.AddressHash([]byte("testAccount")))
	poolCoinDenom := "poolCoinDenom"

	cases := []struct {
		expectedErr string // empty means no error expected
		msg         *types.MsgWithdrawWithinBatch
	}{
		{
			"",
			types.NewMsgWithdrawWithinBatch(withdrawer, DefaultPoolId, sdk.NewCoin(poolCoinDenom, sdk.NewInt(1000))),
		},
		{
			"invalid pool withdrawer address",
			types.NewMsgWithdrawWithinBatch(sdk.AccAddress{}, DefaultPoolId, sdk.NewCoin(poolCoinDenom, sdk.NewInt(1000))),
		},
		{
			"invalid pool coin amount",
			types.NewMsgWithdrawWithinBatch(withdrawer, DefaultPoolId, sdk.NewCoin(poolCoinDenom, sdk.NewInt(0))),
		},
	}

	for _, tc := range cases {
		require.IsType(t, &types.MsgWithdrawWithinBatch{}, tc.msg)
		require.Equal(t, types.TypeMsgWithdrawWithinBatch, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetWithdrawer(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgSwapWithinBatch(t *testing.T) {
	swapRequester := sdk.AccAddress(crypto.AddressHash([]byte("testAccount")))
	offerCoin := sdk.NewCoin(DenomX, sdk.NewInt(1000))
	orderPrice, err := sdk.NewDecFromStr("0.1")
	require.NoError(t, err)

	cases := []struct {
		expectedErr string // empty means no error expected
		msg         *types.MsgSwapWithinBatch
	}{
		{
			"",
			types.NewMsgSwapWithinBatch(swapRequester, DefaultPoolId, DefaultSwapTypeId, offerCoin, DenomY, orderPrice, types.DefaultSwapFeeRate),
		},
		{
			"invalid pool swap requester address",
			types.NewMsgSwapWithinBatch(sdk.AccAddress{}, DefaultPoolId, DefaultSwapTypeId, offerCoin, DenomY, orderPrice, types.DefaultSwapFeeRate),
		},
		{
			"invalid offer coin amount",
			types.NewMsgSwapWithinBatch(swapRequester, DefaultPoolId, DefaultSwapTypeId, sdk.NewCoin(DenomX, sdk.NewInt(0)), DenomY, orderPrice, types.DefaultSwapFeeRate),
		},
		{
			"invalid order price",
			types.NewMsgSwapWithinBatch(swapRequester, DefaultPoolId, DefaultSwapTypeId, offerCoin, DenomY, sdk.ZeroDec(), types.DefaultSwapFeeRate),
		},
		{
			"offer amount should be over 100 micro",
			types.NewMsgSwapWithinBatch(swapRequester, DefaultPoolId, DefaultSwapTypeId, sdk.NewCoin(DenomX, sdk.NewInt(1)), DenomY, orderPrice, types.DefaultSwapFeeRate),
		},
	}

	for _, tc := range cases {
		require.IsType(t, &types.MsgSwapWithinBatch{}, tc.msg)
		require.Equal(t, types.TypeMsgSwapWithinBatch, tc.msg.Type())
		require.Equal(t, types.RouterKey, tc.msg.Route())
		require.Equal(t, sdk.MustSortJSON(types.ModuleCdc.MustMarshalJSON(tc.msg)), tc.msg.GetSignBytes())

		err := tc.msg.ValidateBasic()
		if tc.expectedErr == "" {
			require.Nil(t, err)
			signers := tc.msg.GetSigners()
			require.Len(t, signers, 1)
			require.Equal(t, tc.msg.GetSwapRequester(), signers[0])
		} else {
			require.EqualError(t, err, tc.expectedErr)
		}
	}
}

func TestMsgPanics(t *testing.T) {
	emptyMsgCreatePool := types.MsgCreatePool{}
	emptyMsgDeposit := types.MsgDepositWithinBatch{}
	emptyMsgWithdraw := types.MsgWithdrawWithinBatch{}
	emptyMsgSwap := types.MsgSwapWithinBatch{}
	for _, msg := range []sdk.Msg{&emptyMsgCreatePool, &emptyMsgDeposit, &emptyMsgWithdraw, &emptyMsgSwap} {
		require.PanicsWithError(t, "empty address string is not allowed", func() { msg.GetSigners() })
	}
	for _, tc := range []func() sdk.AccAddress{
		emptyMsgCreatePool.GetPoolCreator,
		emptyMsgDeposit.GetDepositor,
		emptyMsgWithdraw.GetWithdrawer,
		emptyMsgSwap.GetSwapRequester,
	} {
		require.PanicsWithError(t, "empty address string is not allowed", func() { tc() })
	}
}

func TestMsgValidateBasic(t *testing.T) {
	validPoolTypeId := DefaultPoolTypeId
	validAddr := sdk.AccAddress(crypto.AddressHash([]byte("testAccount"))).String()
	validCoin := sdk.NewCoin(DenomY, sdk.NewInt(10000))

	invalidDenomCoin := sdk.Coin{Denom: "-", Amount: sdk.NewInt(10000)}
	negativeCoin := sdk.Coin{Denom: DenomX, Amount: sdk.NewInt(-1)}
	zeroCoin := sdk.Coin{Denom: DenomX, Amount: sdk.ZeroInt()}

	coinsWithInvalidDenom := sdk.Coins{invalidDenomCoin, validCoin}
	coinsWithNegative := sdk.Coins{negativeCoin, validCoin}
	coinsWithZero := sdk.Coins{zeroCoin, validCoin}

	invalidDenomErrMsg := "invalid denom: -"
	negativeCoinErrMsg := "coin -1denomX amount is not positive"
	negativeAmountErrMsg := "negative coin amount: -1"
	zeroCoinErrMsg := "coin 0denomX amount is not positive"

	t.Run("MsgCreatePool", func(t *testing.T) {
		for _, tc := range []struct {
			msg    types.MsgCreatePool
			errMsg string
		}{
			{
				types.MsgCreatePool{},
				types.ErrBadPoolTypeID.Error(),
			},
			{
				types.MsgCreatePool{PoolTypeId: validPoolTypeId},
				types.ErrInvalidPoolCreatorAddr.Error(),
			},
			{
				types.MsgCreatePool{PoolCreatorAddress: validAddr, PoolTypeId: validPoolTypeId},
				types.ErrNumOfReserveCoin.Error(),
			},
			{
				types.MsgCreatePool{
					PoolCreatorAddress: validAddr,
					PoolTypeId:         validPoolTypeId,
					DepositCoins:       coinsWithInvalidDenom,
				},
				invalidDenomErrMsg,
			},
			{
				types.MsgCreatePool{
					PoolCreatorAddress: validAddr,
					PoolTypeId:         validPoolTypeId,
					DepositCoins:       coinsWithNegative,
				},
				negativeCoinErrMsg,
			},
			{
				types.MsgCreatePool{
					PoolCreatorAddress: validAddr,
					PoolTypeId:         validPoolTypeId,
					DepositCoins:       coinsWithZero,
				},
				zeroCoinErrMsg,
			},
			{
				types.MsgCreatePool{
					PoolCreatorAddress: validAddr,
					PoolTypeId:         validPoolTypeId,
					DepositCoins:       sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(int64(types.MinReserveCoinNum)-1))),
				},
				types.ErrNumOfReserveCoin.Error(),
			},
			{
				types.MsgCreatePool{
					PoolCreatorAddress: validAddr,
					PoolTypeId:         validPoolTypeId,
					DepositCoins:       sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(int64(types.MaxReserveCoinNum)+1))),
				},
				types.ErrNumOfReserveCoin.Error(),
			},
		} {
			err := tc.msg.ValidateBasic()
			require.EqualError(t, err, tc.errMsg)
		}
	})
	t.Run("MsgDepositWithinBatch", func(t *testing.T) {
		for _, tc := range []struct {
			msg    types.MsgDepositWithinBatch
			errMsg string
		}{
			{
				types.MsgDepositWithinBatch{},
				types.ErrInvalidDepositorAddr.Error(),
			},
			{
				types.MsgDepositWithinBatch{DepositorAddress: validAddr},
				types.ErrBadDepositCoinsAmount.Error(),
			},
			{
				types.MsgDepositWithinBatch{DepositorAddress: validAddr, DepositCoins: coinsWithInvalidDenom},
				invalidDenomErrMsg,
			},
			{
				types.MsgDepositWithinBatch{DepositorAddress: validAddr, DepositCoins: coinsWithNegative},
				negativeCoinErrMsg,
			},
			{
				types.MsgDepositWithinBatch{DepositorAddress: validAddr, DepositCoins: coinsWithZero},
				zeroCoinErrMsg,
			},
			{
				types.MsgDepositWithinBatch{
					DepositorAddress: validAddr,
					DepositCoins:     sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(int64(types.MinReserveCoinNum)-1))),
				},
				types.ErrNumOfReserveCoin.Error(),
			},
			{
				types.MsgDepositWithinBatch{
					DepositorAddress: validAddr,
					DepositCoins:     sdk.NewCoins(sdk.NewCoin(DenomX, sdk.NewInt(int64(types.MaxReserveCoinNum)+1))),
				},
				types.ErrNumOfReserveCoin.Error(),
			},
		} {
			err := tc.msg.ValidateBasic()
			require.EqualError(t, err, tc.errMsg)
		}
	})
	t.Run("MsgWithdrawWithinBatch", func(t *testing.T) {
		for _, tc := range []struct {
			msg    types.MsgWithdrawWithinBatch
			errMsg string
		}{
			{
				types.MsgWithdrawWithinBatch{},
				types.ErrInvalidWithdrawerAddr.Error(),
			},
			{
				types.MsgWithdrawWithinBatch{WithdrawerAddress: validAddr, PoolCoin: invalidDenomCoin},
				invalidDenomErrMsg,
			},
			{
				types.MsgWithdrawWithinBatch{WithdrawerAddress: validAddr, PoolCoin: negativeCoin},
				negativeAmountErrMsg,
			},
			{
				types.MsgWithdrawWithinBatch{WithdrawerAddress: validAddr, PoolCoin: zeroCoin},
				types.ErrBadPoolCoinAmount.Error(),
			},
		} {
			err := tc.msg.ValidateBasic()
			require.EqualError(t, err, tc.errMsg)
		}
	})
	t.Run("MsgSwap", func(t *testing.T) {
		offerCoin := sdk.NewCoin(DenomX, sdk.NewInt(10000))
		orderPrice := sdk.MustNewDecFromStr("1.0")

		for _, tc := range []struct {
			msg    types.MsgSwapWithinBatch
			errMsg string
		}{
			{
				types.MsgSwapWithinBatch{},
				types.ErrInvalidSwapRequesterAddr.Error(),
			},
			{
				types.MsgSwapWithinBatch{SwapRequesterAddress: validAddr, OfferCoin: invalidDenomCoin, OrderPrice: orderPrice},
				invalidDenomErrMsg,
			},
			{
				types.MsgSwapWithinBatch{SwapRequesterAddress: validAddr, OfferCoin: zeroCoin},
				types.ErrBadOfferCoinAmount.Error(),
			},
			{
				types.MsgSwapWithinBatch{SwapRequesterAddress: validAddr, OfferCoin: negativeCoin},
				negativeAmountErrMsg,
			},
			{
				types.MsgSwapWithinBatch{SwapRequesterAddress: validAddr, OfferCoin: offerCoin, OrderPrice: sdk.ZeroDec()},
				types.ErrBadOrderPrice.Error(),
			},
			{
				types.MsgSwapWithinBatch{SwapRequesterAddress: validAddr, OfferCoin: sdk.NewCoin(DenomX, sdk.OneInt()), OrderPrice: orderPrice},
				types.ErrLessThanMinOfferAmount.Error(),
			},
		} {
			err := tc.msg.ValidateBasic()
			require.EqualError(t, err, tc.errMsg)
		}
	})
}
