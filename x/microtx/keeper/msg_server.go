package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/althea-net/althea-chain/x/microtx/types"
)

// BasisPointDivisor used in calculating the MsgXfer fee amount to deduct
const BasisPointDivisor uint64 = 10000

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
	if err := m.Keeper.Xfer(ctx, sender, receiver, msg.Amounts); err != nil {
		return nil, sdkerrors.Wrap(err, "unable to complete the transfer")
	}

	return &types.MsgXferResponse{}, err
}

// Xfer implements the transfer of funds from sender to receiver
func (k Keeper) Xfer(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, amounts sdk.Coins) error {
	feesCollected, err := k.DeductXferFee(ctx, sender, amounts)

	if err != nil {
		return sdkerrors.Wrap(err, "unable to collect fees")
	}

	err = k.bankKeeper.SendCoins(ctx, sender, receiver, amounts)

	if err != nil {
		return sdkerrors.Wrap(err, "unable to send tokens via the bank module")
	}

	// Emit an event for the block's event log
	ctx.EventManager().EmitTypedEvent(
		&types.EventXfer{
			Sender:   sender.String(),
			Receiver: receiver.String(),
			Amounts:  amounts,
			Fee:      feesCollected,
		},
	)

	return nil
}

// checkAndDeductSendToEthFees asserts that the minimum chainFee has been met for the given sendAmount
func (k Keeper) DeductXferFee(ctx sdk.Context, sender sdk.AccAddress, sendAmounts sdk.Coins) (feeCollected sdk.Coins, err error) {
	// Compute the minimum fees which must be paid
	xferFeeBasisPoints := int64(0)
	params, err := k.GetParamsIfSet(ctx)
	if err == nil {
		// The params have been set, get the min send to eth fee
		xferFeeBasisPoints = int64(params.XferFeeBasisPoints)
	}
	var xferFees sdk.Coins
	for _, sendAmount := range sendAmounts {
		xferFee := sdk.NewDecFromInt(sendAmount.Amount).
			QuoInt64(int64(BasisPointDivisor)).
			MulInt64(xferFeeBasisPoints).
			TruncateInt()
		xferFeeCoin := sdk.NewCoin(sendAmount.Denom, xferFee)
		xferFees = xferFees.Add(xferFeeCoin)
	}

	// Require that the minimum has been met
	if !xferFees.IsZero() { // Ignore fees too low to collect
		balances := k.bankKeeper.GetAllBalances(ctx, sender)
		if xferFees.IsAnyGT(balances) {
			err := sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFee,
				"balances are insufficient, one of the needed fees are larger (%v > %v)",
				xferFees,
				balances,
			)
			return nil, err
		}

		// Finally, collect the necessary fee
		senderAcc := k.accountKeeper.GetAccount(ctx, sender)

		err = sdkante.DeductFees(k.bankKeeper, ctx, senderAcc, xferFees)
		if err != nil {
			ctx.Logger().Error("Could not deduct MsgXfer fee!", "error", err, "account", senderAcc, "fees", xferFees)
			return nil, err
		}
	}

	return xferFees, nil
}
