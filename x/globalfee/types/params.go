package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKeyMinGasPrices store key
var ParamStoreKeyMinGasPrices = []byte("MinimumGasPricesParam")

// DefaultParams returns default wasm parameters
func DefaultParams() Params {
	return Params{MinimumGasPrices: sdk.DecCoins{}}
}

func ParamKeyTable() types.KeyTable {
	return types.NewKeyTable().RegisterParamSet(&Params{})
}

// ValidateBasic performs basic validation.
func (p Params) ValidateBasic() error {
	return validateMinimumGasPrices(p.MinimumGasPrices)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		types.NewParamSetPair(
			ParamStoreKeyMinGasPrices, &p.MinimumGasPrices, validateMinimumGasPrices,
		),
	}
}

func validateMinimumGasPrices(i interface{}) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalid, "type: %T", i)
	}
	return v.Validate()
}
