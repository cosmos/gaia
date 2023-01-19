package keeper

import (
	"context"

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
func (m msgServer) Name(c context.Context, msg *types.MsgName) (*types.MsgNameResponse, error) {
	// ctx := sdk.UnwrapSDKContext(c)

	// TODO: Implement the Name state changes
	return nil, nil
}
