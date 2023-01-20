package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	microtxtypes "github.com/althea-net/althea-chain/x/microtx/types"
)

// InitGenesis starts a chain from a genesis state
func InitGenesis(ctx sdk.Context, k Keeper, data microtxtypes.GenesisState) {
	k.SetParams(ctx, *data.Params)
}

// ExportGenesis exports all the state needed to restart the chain
// from the current state of the chain
func ExportGenesis(ctx sdk.Context, k Keeper) microtxtypes.GenesisState {
	var (
		p = k.GetParams(ctx)
	)

	return microtxtypes.GenesisState{
		Params: &p,
	}
}
