package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
)

var _ types.QueryServer = Keeper{}

func (q Keeper) Result(c context.Context, req *types.QueryResultRequest) (*types.QueryResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	bz, err := q.GetResult(ctx, bandoracle.RequestID(req.RequestId))
	if err != nil {
		return nil, err
	}
	return &types.QueryResultResponse{Result: bz}, nil
}
