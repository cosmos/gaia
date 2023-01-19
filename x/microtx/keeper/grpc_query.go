package keeper

import (
	"context"

	"github.com/althea-net/althea-chain/x/microtx/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ types.QueryServer = Keeper{
	storeKey:   nil,
	paramSpace: paramstypes.Subspace{},
	cdc:        nil,
	bankKeeper: &bankkeeper.BaseKeeper{},
}

// Data is an example query endpoint
func (k Keeper) Data(c context.Context, req *types.QueryData) (*types.QueryDataResponse, error) {
	// TODO: fetch the data from the keeper, format, return
	return nil, nil
}
