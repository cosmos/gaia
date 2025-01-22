package keeper

import (
	"context"

	"cosmossdk.io/math"

	"github.com/cosmos/gaia/v23/x/lsm/types"
)

// SetParams sets the x/lsm module parameters.
// CONTRACT: This method performs no validation of the parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// GetParams gets the x/lsm module parameters.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return params, err
	}

	if bz == nil {
		return params, nil
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// Validator bond factor for all validators
func (k Keeper) ValidatorBondFactor(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	return params.ValidatorBondFactor, err
}

// Global liquid staking cap across all liquid staking providers
func (k Keeper) GlobalLiquidStakingCap(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	return params.GlobalLiquidStakingCap, err
}

// Liquid staking cap for each validator
func (k Keeper) ValidatorLiquidStakingCap(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	return params.ValidatorLiquidStakingCap, err
}
