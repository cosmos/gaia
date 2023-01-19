package keeper

import (
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
