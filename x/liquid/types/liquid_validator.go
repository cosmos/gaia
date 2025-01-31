package types

import (
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewLiquidValidator constructs a new LiquidValidator
func NewLiquidValidator(operator string) LiquidValidator {
	return LiquidValidator{
		OperatorAddress: operator,
		LiquidShares:    math.LegacyZeroDec(),
	}
}

func MustMarshalValidator(cdc codec.BinaryCodec, validator *LiquidValidator) []byte {
	return cdc.MustMarshal(validator)
}

// unmarshal from a store value
func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v LiquidValidator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
