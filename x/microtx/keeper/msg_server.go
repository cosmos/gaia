package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/althea-net/althea-chain/x/microtx/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// Xfer delegates the msg server's call to the keeper
func (m msgServer) Xfer(c context.Context, msg *types.MsgXfer) (*types.MsgXferResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// The following validation logic has been copied from x/bank in the sdk
	if err := m.bankKeeper.IsSendEnabledCoins(ctx, msg.Amounts...); err != nil {
		return nil, err
	}

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	if m.bankKeeper.BlockedAddr(receiver) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", msg.Receiver)
	}

	// Call the actual transfer implementation
	err = m.Keeper.Xfer(ctx, sender, receiver, msg.Amounts)

	// Emit an event for the block's event_log
	ctx.EventManager().EmitTypedEvent(
		&types.EventXfer{
			Sender:   msg.Sender,
			Receiver: msg.Receiver,
			Amounts:  msg.Amounts,
			Fee:      []sdk.Coin{}, // TODO: Calculate and report fee here
		},
	)

	return &types.MsgXferResponse{}, err
}

// Xfer implements the transfer of funds from sender to receiver
func (k Keeper) Xfer(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, amounts sdk.Coins) error {
	// TODO: Collect a fee here

	return k.bankKeeper.SendCoins(ctx, sender, receiver, amounts)
}
