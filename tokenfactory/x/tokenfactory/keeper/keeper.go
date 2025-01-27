package keeper

import (
	"context"
	"fmt"

	"github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	store "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type (
	// IsAdmin is a function signature that checks if an address is an admin.
	IsSudoAdmin func(ctx context.Context, addr string) bool

	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  store.StoreKey
		permAddrs map[string]authtypes.PermissionsForAddress

		accountKeeper       types.AccountKeeper
		bankKeeper          types.BankKeeper
		communityPoolKeeper types.CommunityPoolKeeper

		enabledCapabilities []string

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		IsSudoAdminFunc IsSudoAdmin
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey store.StoreKey,
	maccPerms map[string][]string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	communityPoolKeeper types.CommunityPoolKeeper,
	enabledCapabilities []string,
	// use DefaultIsSudoAdminFunc if you don't have a custom one
	isSudoAdminFunc IsSudoAdmin,
	authority string,
) Keeper {
	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = authtypes.NewPermissionsForAddress(name, perms)
	}

	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		permAddrs: permAddrs,

		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,

		authority: authority,

		enabledCapabilities: enabledCapabilities,

		IsSudoAdminFunc: isSudoAdminFunc,
	}
}

// DefaultIsSudoAdminFunc returns false for all addresses.
func DefaultIsSudoAdminFunc(_ context.Context, _ string) bool {
	return false
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) GetEnabledCapabilities() []string {
	return k.enabledCapabilities
}

func (k *Keeper) SetEnabledCapabilities(_ sdk.Context, newCapabilities []string) {
	k.enabledCapabilities = newCapabilities
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDenomPrefixStore returns the substore for a specific denom
func (k Keeper) GetDenomPrefixStore(ctx sdk.Context, denom string) store.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

// GetCreatorPrefixStore returns the substore for a specific creator address
func (k Keeper) GetCreatorPrefixStore(ctx sdk.Context, creator string) store.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

// GetCreatorsPrefixStore returns the substore that contains a list of creators
func (k Keeper) GetCreatorsPrefixStore(ctx sdk.Context) store.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}
