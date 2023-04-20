package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = (*MsgCreatePool)(nil)
	_ sdk.Msg = (*MsgDepositWithinBatch)(nil)
	_ sdk.Msg = (*MsgWithdrawWithinBatch)(nil)
	_ sdk.Msg = (*MsgSwapWithinBatch)(nil)
)

// Message types for the liquidity module
//
//nolint:gosec
const (
	TypeMsgCreatePool          = "create_pool"
	TypeMsgDepositWithinBatch  = "deposit_within_batch"
	TypeMsgWithdrawWithinBatch = "withdraw_within_batch"
	TypeMsgSwapWithinBatch     = "swap_within_batch"
)

// NewMsgCreatePool creates a new MsgCreatePool.
func NewMsgCreatePool(poolCreator sdk.AccAddress, poolTypeID uint32, depositCoins sdk.Coins) *MsgCreatePool {
	return &MsgCreatePool{
		PoolCreatorAddress: poolCreator.String(),
		PoolTypeId:         poolTypeID,
		DepositCoins:       depositCoins,
	}
}

func (msg MsgCreatePool) Route() string { return RouterKey }

func (msg MsgCreatePool) Type() string { return TypeMsgCreatePool }

func (msg MsgCreatePool) ValidateBasic() error {
	if 1 > msg.PoolTypeId {
		return ErrBadPoolTypeID
	}
	if _, err := sdk.AccAddressFromBech32(msg.PoolCreatorAddress); err != nil {
		return ErrInvalidPoolCreatorAddr
	}
	if err := msg.DepositCoins.Validate(); err != nil {
		return err
	}
	if n := uint32(len(msg.DepositCoins)); n > MaxReserveCoinNum || n < MinReserveCoinNum {
		return ErrNumOfReserveCoin
	}
	return nil
}

func (msg MsgCreatePool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgCreatePool) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.PoolCreatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgCreatePool) GetPoolCreator() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.PoolCreatorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgDepositWithinBatch creates a new MsgDepositWithinBatch.
func NewMsgDepositWithinBatch(depositor sdk.AccAddress, poolID uint64, depositCoins sdk.Coins) *MsgDepositWithinBatch {
	return &MsgDepositWithinBatch{
		DepositorAddress: depositor.String(),
		PoolId:           poolID,
		DepositCoins:     depositCoins,
	}
}

func (msg MsgDepositWithinBatch) Route() string { return RouterKey }

func (msg MsgDepositWithinBatch) Type() string { return TypeMsgDepositWithinBatch }

func (msg MsgDepositWithinBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DepositorAddress); err != nil {
		return ErrInvalidDepositorAddr
	}
	if err := msg.DepositCoins.Validate(); err != nil {
		return err
	}
	if !msg.DepositCoins.IsAllPositive() {
		return ErrBadDepositCoinsAmount
	}
	if n := uint32(len(msg.DepositCoins)); n > MaxReserveCoinNum || n < MinReserveCoinNum {
		return ErrNumOfReserveCoin
	}
	return nil
}

func (msg MsgDepositWithinBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgDepositWithinBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.DepositorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgDepositWithinBatch) GetDepositor() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.DepositorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgWithdrawWithinBatch creates a new MsgWithdrawWithinBatch.
func NewMsgWithdrawWithinBatch(withdrawer sdk.AccAddress, poolID uint64, poolCoin sdk.Coin) *MsgWithdrawWithinBatch {
	return &MsgWithdrawWithinBatch{
		WithdrawerAddress: withdrawer.String(),
		PoolId:            poolID,
		PoolCoin:          poolCoin,
	}
}

func (msg MsgWithdrawWithinBatch) Route() string { return RouterKey }

func (msg MsgWithdrawWithinBatch) Type() string { return TypeMsgWithdrawWithinBatch }

func (msg MsgWithdrawWithinBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress); err != nil {
		return ErrInvalidWithdrawerAddr
	}
	if err := msg.PoolCoin.Validate(); err != nil {
		return err
	}
	if !msg.PoolCoin.IsPositive() {
		return ErrBadPoolCoinAmount
	}
	return nil
}

func (msg MsgWithdrawWithinBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgWithdrawWithinBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgWithdrawWithinBatch) GetWithdrawer() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.WithdrawerAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

// NewMsgSwapWithinBatch creates a new MsgSwapWithinBatch.
func NewMsgSwapWithinBatch(
	swapRequester sdk.AccAddress,
	poolID uint64,
	swapTypeID uint32,
	offerCoin sdk.Coin,
	demandCoinDenom string,
	orderPrice sdk.Dec,
	swapFeeRate sdk.Dec,
) *MsgSwapWithinBatch {
	return &MsgSwapWithinBatch{
		SwapRequesterAddress: swapRequester.String(),
		PoolId:               poolID,
		SwapTypeId:           swapTypeID,
		OfferCoin:            offerCoin,
		OfferCoinFee:         GetOfferCoinFee(offerCoin, swapFeeRate),
		DemandCoinDenom:      demandCoinDenom,
		OrderPrice:           orderPrice,
	}
}

func (msg MsgSwapWithinBatch) Route() string { return RouterKey }

func (msg MsgSwapWithinBatch) Type() string { return TypeMsgSwapWithinBatch }

func (msg MsgSwapWithinBatch) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.SwapRequesterAddress); err != nil {
		return ErrInvalidSwapRequesterAddr
	}
	if err := msg.OfferCoin.Validate(); err != nil {
		return err
	}
	if !msg.OfferCoin.IsPositive() {
		return ErrBadOfferCoinAmount
	}
	if !msg.OrderPrice.IsPositive() {
		return ErrBadOrderPrice
	}
	if !msg.OfferCoin.Amount.GTE(MinOfferCoinAmount) {
		return ErrLessThanMinOfferAmount
	}
	return nil
}

func (msg MsgSwapWithinBatch) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgSwapWithinBatch) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.SwapRequesterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgSwapWithinBatch) GetSwapRequester() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.SwapRequesterAddress)
	if err != nil {
		panic(err)
	}
	return addr
}
