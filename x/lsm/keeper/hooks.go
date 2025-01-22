package keeper

import (
	"context"
	"errors"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/x/lsm/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Create new lsm hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// initialize liquid validator record
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	val, err := h.k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
	lVal := types.NewLiquidValidator(val.GetOperator())
	del, err := h.k.stakingKeeper.GetDelegation(ctx, sdk.AccAddress(val.GetOperator()), valAddr)
	if err != nil && !errors.Is(err, stakingtypes.ErrNoDelegation) {
		return err
	} else if err == nil {
		lVal.ValidatorBondShares = del.Shares
	}
	return h.k.SetLiquidValidator(ctx, lVal)
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RemoveLiquidValidator(ctx, valAddr)
}

func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if delAddr.Equals(sdk.AccAddress(valAddr)) {
		del, err := h.k.stakingKeeper.GetDelegation(ctx, sdk.AccAddress(valAddr), valAddr)
		if err != nil {
			return err
		}
		lVal, err := h.k.GetLiquidValidator(ctx, valAddr)
		if err != nil {
			return err
		}
		lVal.ValidatorBondShares = del.Shares
		return h.k.SetLiquidValidator(ctx, lVal)
	}
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	validator, err := h.k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
	liquidVal, err := h.k.GetLiquidValidator(ctx, valAddr)
	if err != nil {
		return err
	}
	initialLiquidTokens := validator.TokensFromShares(liquidVal.LiquidShares).TruncateInt()
	slashedLiquidTokens := fraction.Mul(sdkmath.LegacyNewDecFromInt(initialLiquidTokens))

	decrease := slashedLiquidTokens.TruncateInt()
	if err := h.k.DecreaseTotalLiquidStakedTokens(ctx, decrease); err != nil {
		// This only error's if the total liquid staked tokens underflows
		// which would indicate there's a corrupted state where the validator has
		// liquid tokens that are not accounted for in the global total
		panic(err)
	}
	return nil
}

func (h Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if delAddr.Equals(sdk.AccAddress(valAddr)) {
		lVal, err := h.k.GetLiquidValidator(ctx, valAddr)
		if err != nil {
			return err
		}
		lVal.ValidatorBondShares = sdkmath.LegacyZeroDec()
		return h.k.SetLiquidValidator(ctx, lVal)
	}
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
