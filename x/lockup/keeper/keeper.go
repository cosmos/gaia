package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/althea-net/althea-chain/x/lockup/types"
)

type Keeper struct {
	storeKey   sdk.StoreKey
	paramSpace paramstypes.Subspace
	cdc        codec.BinaryMarshaler
}

func NewKeeper(cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramstypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		cdc:        cdc,
		paramSpace: paramSpace,
		storeKey:   storeKey,
	}

	return k
}

// TODO: Doc all these methods
func (k Keeper) GetChainLocked(ctx sdk.Context) bool {
	locked := types.DefaultParams().Locked
	k.paramSpace.GetIfExists(ctx, types.LockedKey, &locked)
	return locked
}

func (k Keeper) SetChainLocked(ctx sdk.Context, locked bool) {
	k.paramSpace.Set(ctx, types.LockedKey, &locked)
}

func (k Keeper) GetLockExemptAddresses(ctx sdk.Context) []string {
	lockExempt := types.DefaultParams().LockExempt
	k.paramSpace.GetIfExists(ctx, types.LockExemptKey, &lockExempt)
	return lockExempt
}

func (k Keeper) GetLockExemptAddressesSet(ctx sdk.Context) map[string]struct{} {
	return createSet(k.GetLockExemptAddresses(ctx))
}

// TODO: It would be nice to just store the pseudo-set instead of the string array
// so that we get better efficiency on each read (happens each transaction in antehandler)
// however we would need to make a custom param change proposal handler to construct the
// set upon governance proposal before storage in keeper
func (k Keeper) SetLockExemptAddresses(ctx sdk.Context, lockExempt []string) {
	k.paramSpace.Set(ctx, types.LockExemptKey, &lockExempt)
}

func (k Keeper) GetLockedMessageTypes(ctx sdk.Context) []string {
	lockedMessageTypes := types.DefaultParams().LockedMessageTypes
	k.paramSpace.GetIfExists(ctx, types.LockedMessageTypesKey, &lockedMessageTypes)
	return lockedMessageTypes
}

func (k Keeper) GetLockedMessageTypesSet(ctx sdk.Context) map[string]struct{} {
	return createSet(k.GetLockedMessageTypes(ctx))
}

func (k Keeper) SetLockedMessageTypes(ctx sdk.Context, lockedMessageTypes []string) {
	k.paramSpace.Set(ctx, types.LockedMessageTypesKey, &lockedMessageTypes)
}

func createSet(strings []string) map[string]struct{} {
	type void struct{}
	var member void
	set := make(map[string]struct{})

	for _, str := range strings {
		if _, present := set[str]; present {
			continue
		}
		set[str] = member
	}

	return set
}
