package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v24/x/liquid/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Create new liquid hooks
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
	return h.k.SetLiquidValidator(ctx, lVal)
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RemoveLiquidValidator(ctx, valAddr)
}

func (h Hooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterDelegationModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	// fraction = tokens_to_burn / validator.Tokens
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

func (h Hooks) BeforeDelegationRemoved(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

func (h Hooks) BeforeTokenizeShareRecordRemoved(_ context.Context, _ uint64) error {
	return nil
}
