package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamStoreKeyMinGasPrices store key
var ParamStoreKeyMinGasPrices = []byte("MinimumGasPricesParam")

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{MinimumGasPrices: sdk.DecCoins{}}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ValidateBasic performs basic validation.
func (p Params) ValidateBasic() error {
	return validateMinimumGasPrices(p.MinimumGasPrices)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			ParamStoreKeyMinGasPrices, &p.MinimumGasPrices, validateMinimumGasPrices,
		),
	}
}

// this requires the fee non-negative
func validateMinimumGasPrices(i interface{}) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected sdk.DecCoins", i)
	}

	dec := DecCoins(v)
	return dec.Validate()
}

type DecCoins sdk.DecCoins

// Validate checks that the DecCoins are sorted, have nonnegtive amount, with a valid and unique
// denomination (i.e no duplicates). Otherwise, it returns an error.
func (coins DecCoins) Validate() error {
	if len(coins) == 0 {
		return nil
	}

	lowDenom := ""
	seenDenoms := make(map[string]bool)

	for _, coin := range coins {
		if seenDenoms[coin.Denom] {
			return fmt.Errorf("duplicate denomination %s", coin.Denom)
		}
		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return err
		}
		if coin.Denom <= lowDenom {
			return fmt.Errorf("denomination %s is not sorted", coin.Denom)
		}
		if coin.IsNegative() {
			return fmt.Errorf("coin %s amount is negative", coin.Amount)
		}

		// we compare each coin against the last denom
		lowDenom = coin.Denom
		seenDenoms[coin.Denom] = true
	}

	return nil
}
