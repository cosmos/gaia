package types

import (
	"bytes"
	"sort"
	"strings"

	"cosmossdk.io/core/address"
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

// LiquidValidators is a collection of LiquidValidator
type LiquidValidators struct {
	LiquidValidators []LiquidValidator
	ValidatorCodec   address.Codec
}

func (v LiquidValidators) String() (out string) {
	for _, val := range v.LiquidValidators {
		out += val.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// Sort Validators sorts validator array in ascending operator address order
func (v LiquidValidators) Sort() {
	sort.Sort(v)
}

// Len implements sort interface
func (v LiquidValidators) Len() int {
	return len(v.LiquidValidators)
}

// Less implements sort interface
func (v LiquidValidators) Less(i, j int) bool {
	vi, err := v.ValidatorCodec.StringToBytes(v.LiquidValidators[i].OperatorAddress)
	if err != nil {
		panic(err)
	}
	vj, err := v.ValidatorCodec.StringToBytes(v.LiquidValidators[j].OperatorAddress)
	if err != nil {
		panic(err)
	}

	return bytes.Compare(vi, vj) == -1
}

// Swap implements sort interface
func (v LiquidValidators) Swap(i, j int) {
	v.LiquidValidators[i], v.LiquidValidators[j] = v.LiquidValidators[j], v.LiquidValidators[i]
}

func MustMarshalValidator(cdc codec.BinaryCodec, validator *LiquidValidator) []byte {
	return cdc.MustMarshal(validator)
}

// unmarshal from a store value
func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v LiquidValidator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
