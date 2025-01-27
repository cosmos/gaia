package keeper

import (
	"context"

	"github.com/cosmos/gaia/v23/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx context.Context) (p types.Params) {
	store := sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := sdk.UnwrapSDKContext(ctx).KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)

	return nil
}
