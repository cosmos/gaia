package keeper

import (
	"github.com/althea-net/althea-chain/x/microtx/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	// NOTE: If you add anything to this struct, add a nil check to ValidateMembers below!
	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	paramSpace paramtypes.Subspace

	// NOTE: If you add anything to this struct, add a nil check to ValidateMembers below!
	cdc           codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper    *bankkeeper.BaseKeeper
	accountKeeper *authkeeper.AccountKeeper
}

// Check for nil members
func (k Keeper) ValidateMembers() {
	if k.bankKeeper == nil {
		panic("Nil bankKeeper!")
	}
	if k.accountKeeper == nil {
		panic("Nil accountKeeper!")
	}
}

// NewKeeper returns a new instance of the gravity keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	cdc codec.BinaryCodec,
	bankKeeper *bankkeeper.BaseKeeper,
	accKeeper *authkeeper.AccountKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,

		cdc:           cdc,
		bankKeeper:    bankKeeper,
		accountKeeper: accKeeper,
	}

	k.ValidateMembers()

	return k
}

// ReadData is an example store read function
func (k Keeper) ReadData(ctx sdk.Context, key string) (string, error) {
	// store := ctx.KVStore(k.storeKey)

	// TODO: Implement data fetching
	return "", nil
}

// SetValue is an example store write function
func (k Keeper) SetValue(ctx sdk.Context, key string, value string) (string, error) {
	// store := ctx.KVStore(k.storeKey)
	// if store.Has([]byte(key)) {
	// 	panic("Oh no what do I do?")
	// }
	// store.Set([]byte(key), []byte(value))
	// return "Job's done", nil

	// TODO: Implement data storage
	return "", nil
}

func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "unable to store params with failing ValidateBasic()")
	}
	k.paramSpace.SetParamSet(ctx, &params)
	return nil
}
