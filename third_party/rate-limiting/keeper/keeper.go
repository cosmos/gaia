package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v11/types"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		paramstore   paramtypes.Subspace
		authority    string

		bankKeeper    types.BankKeeper
		channelKeeper types.ChannelKeeper
		clientKeeper  types.ClientKeeper
		ics4Wrapper   types.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ps paramtypes.Subspace,
	authority string,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	clientKeeper types.ClientKeeper,
	ics4Wrapper types.ICS4Wrapper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeService:  storeService,
		paramstore:    ps,
		authority:     authority,
		bankKeeper:    bankKeeper,
		channelKeeper: channelKeeper,
		clientKeeper:  clientKeeper,
		ics4Wrapper:   ics4Wrapper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetIBCKeepers allows us to set the relevant IBC keepers post dependency
// injection, as IBC doesn't support dependency injection yet.
func (k *Keeper) SetIBCKeepers(channelKeeper types.ChannelKeeper, clientKeeper types.ClientKeeper, ics4Wrapper types.ICS4Wrapper) {
	k.channelKeeper = channelKeeper
	k.clientKeeper = clientKeeper
	k.ics4Wrapper = ics4Wrapper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
