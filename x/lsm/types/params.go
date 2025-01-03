package types

import (
	"fmt"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	// ValidatorBondFactor of -1 indicates that it's disabled
	ValidatorBondCapDisabled = math.LegacyNewDecFromInt(math.NewInt(-1))

	// DefaultValidatorBondFactor is set to -1 (disabled)
	DefaultValidatorBondFactor = ValidatorBondCapDisabled
	// DefaultGlobalLiquidStakingCap is set to 100%
	DefaultGlobalLiquidStakingCap = math.LegacyOneDec()
	// DefaultValidatorLiquidStakingCap is set to 100%
	DefaultValidatorLiquidStakingCap = math.LegacyOneDec()
)

// NewParams creates a new Params instance
func NewParams(
	validatorBondFactor math.LegacyDec,
	globalLiquidStakingCap math.LegacyDec,
	validatorLiquidStakingCap math.LegacyDec,
) Params {
	return Params{
		ValidatorBondFactor:       validatorBondFactor,
		GlobalLiquidStakingCap:    globalLiquidStakingCap,
		ValidatorLiquidStakingCap: validatorLiquidStakingCap,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultValidatorBondFactor,
		DefaultGlobalLiquidStakingCap,
		DefaultValidatorLiquidStakingCap,
	)
}

// unmarshal the current lsm params value from store key or panic
func MustUnmarshalParams(cdc *codec.LegacyAmino, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// unmarshal the current lsm params value from store key
func UnmarshalParams(cdc *codec.LegacyAmino, value []byte) (params Params, err error) {
	err = cdc.Unmarshal(value, &params)
	if err != nil {
		return
	}

	return
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateValidatorBondFactor(p.ValidatorBondFactor); err != nil {
		return err
	}

	if err := validateGlobalLiquidStakingCap(p.GlobalLiquidStakingCap); err != nil {
		return err
	}

	return validateValidatorLiquidStakingCap(p.ValidatorLiquidStakingCap)
}

func validateValidatorBondFactor(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() && !v.Equal(math.LegacyNewDec(-1)) {
		return fmt.Errorf("invalid validator bond factor: %s", v)
	}

	return nil
}

func validateGlobalLiquidStakingCap(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("global liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("global liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}

func validateValidatorLiquidStakingCap(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("validator liquid staking cap cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("validator liquid staking cap cannot be greater than 100%%: %s", v)
	}

	return nil
}
