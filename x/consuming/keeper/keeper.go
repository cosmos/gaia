package keeper

import (
	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryMarshaler
	ChannelKeeper types.ChannelKeeper
	scopedKeeper  capabilitykeeper.ScopedKeeper
}

// NewKeeper creates a new oracle Keeper instance.
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, channelKeeper types.ChannelKeeper, scopedKeeper capabilitykeeper.ScopedKeeper) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		ChannelKeeper: channelKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

func (k Keeper) SetResult(ctx sdk.Context, requestID bandoracle.RequestID, result []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ResultStoreKey(requestID), result)
}

func (k Keeper) GetResult(ctx sdk.Context, requestID bandoracle.RequestID) ([]byte, error) {
	if !k.HasResult(ctx, requestID) {
		return nil, sdkerrors.Wrapf(types.ErrItemNotFound,
			"GetResult: Result for request ID %d is not available.", requestID,
		)
	}
	store := ctx.KVStore(k.storeKey)
	return store.Get(types.ResultStoreKey(requestID)), nil
}

func (k Keeper) HasResult(ctx sdk.Context, requestID bandoracle.RequestID) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ResultStoreKey(requestID))
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
