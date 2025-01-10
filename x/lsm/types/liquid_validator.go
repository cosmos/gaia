package types

import "github.com/cosmos/cosmos-sdk/codec"

func MustMarshalValidator(cdc codec.BinaryCodec, validator *LiquidValidator) []byte {
	return cdc.MustMarshal(validator)
}

// unmarshal from a store value
func UnmarshalValidator(cdc codec.BinaryCodec, value []byte) (v LiquidValidator, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
