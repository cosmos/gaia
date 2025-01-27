package keeper

import (
	"context"

	"cosmossdk.io/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) addDenomFromCreator(ctx context.Context, creator, denom string) {
	store := k.GetCreatorPrefixStore(sdk.UnwrapSDKContext(ctx), creator)
	store.Set([]byte(denom), []byte(denom))
}

func (k Keeper) GetDenomsFromCreator(ctx context.Context, creator string) []string {
	store := k.GetCreatorPrefixStore(sdk.UnwrapSDKContext(ctx), creator)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	denoms := []string{}
	for ; iterator.Valid(); iterator.Next() {
		denoms = append(denoms, string(iterator.Key()))
	}
	return denoms
}

func (k Keeper) GetAllDenomsIterator(ctx context.Context) store.Iterator {
	return k.GetCreatorsPrefixStore(sdk.UnwrapSDKContext(ctx)).Iterator(nil, nil)
}
