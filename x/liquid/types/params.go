package types

import (
	"fmt"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
)

var (

	// DefaultGlobalLiquidStakingCap is set to 100%
	DefaultGlobalLiquidStakingCap = math.LegacyOneDec()
	// DefaultValidatorLiquidStakingCap is set to 100%
	DefaultValidatorLiquidStakingCap = math.LegacyOneDec()
)

// NewParams creates a new Params instance
func NewParams(
	globalLiquidStakingCap math.LegacyDec,
	validatorLiquidStakingCap math.LegacyDec,
) Params {
	return Params{
		GlobalLiquidStakingCap:    globalLiquidStakingCap,
		ValidatorLiquidStakingCap: validatorLiquidStakingCap,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultGlobalLiquidStakingCap,
		DefaultValidatorLiquidStakingCap,
	)
}

// unmarshal the current liquid params value from store key or panic
func MustUnmarshalParams(cdc *codec.LegacyAmino, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}

	return params
}

// unmarshal the current liquid params value from store key
func UnmarshalParams(cdc *codec.LegacyAmino, value []byte) (params Params, err error) {
	err = cdc.Unmarshal(value, &params)
	if err != nil {
		return
	}

	return
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateGlobalLiquidStakingCap(p.GlobalLiquidStakingCap); err != nil {
		return err
	}

	return validateValidatorLiquidStakingCap(p.ValidatorLiquidStakingCap)
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
