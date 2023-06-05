package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/althea-net/ibc-test-chain/v9/x/icaauth/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl creates and returns a new types.MsgServer, fulfilling the icaauth Msg service interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterAccount implements the Msg/RegisterAccount interface
func (k msgServer) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	acc, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := k.RegisterInterchainAccount(ctx, acc, msg.ConnectionId, msg.Version); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAccountResponse{}, nil
}

// SubmitTx implements the Msg/SubmitTx interface
func (k msgServer) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}

	err = k.DoSubmitTx(ctx, msg.ConnectionId, msg.Owner, msgs)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitTypedEvent(
		&types.EventSubmitTx{
			Owner:        msg.Owner,
			ConnectionId: msg.ConnectionId,
			Msgs:         msg.Msgs,
		},
	)

	return &types.MsgSubmitTxResponse{}, nil
}
