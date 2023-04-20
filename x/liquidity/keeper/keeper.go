package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// Keeper of the liquidity store
type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      sdk.StoreKey
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	distrKeeper   types.DistributionKeeper
	paramSpace    paramstypes.Subspace
}

// NewKeeper returns a liquidity keeper. It handles:
// - creating new ModuleAccounts for each pool ReserveAccount
// - sending to and from ModuleAccounts
// - minting, burning PoolCoins
func NewKeeper(cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramstypes.Subspace, bankKeeper types.BankKeeper, accountKeeper types.AccountKeeper, distrKeeper types.DistributionKeeper) Keeper {
	// ensure liquidity module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      key,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		distrKeeper:   distrKeeper,
		cdc:           cdc,
		paramSpace:    paramSpace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// GetParams gets the parameters for the liquidity module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters for the liquidity module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetCircuitBreakerEnabled returns circuit breaker enabled param from the paramspace.
func (k Keeper) GetCircuitBreakerEnabled(ctx sdk.Context) (enabled bool) {
	k.paramSpace.Get(ctx, types.KeyCircuitBreakerEnabled, &enabled)
	return
}
