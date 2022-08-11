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

// Validate checks that the DecCoins are sorted, have nonnegtive amount, with a valid and unique
// denomination (i.e no duplicates). Otherwise, it returns an error.
type DecCoins sdk.DecCoins

func (coins DecCoins) Validate() error {
	switch len(coins) {
	case 0:
		return nil

	case 1:
		// match the denom reg expr
		if err := sdk.ValidateDenom(coins[0].Denom); err != nil {
			return err
		}
		if coins[0].IsNegative() {
			return fmt.Errorf("coin %s amount is negtive", coins[0])
		}
		return nil
	default:
		// check single coin case
		if err := (DecCoins{coins[0]}).Validate(); err != nil {
			return err
		}

		lowDenom := coins[0].Denom
		seenDenoms := make(map[string]bool)
		seenDenoms[lowDenom] = true

		for _, coin := range coins[1:] {
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
				return fmt.Errorf("coin %s amount is negtive", coin.Denom)
			}

			// we compare each coin against the last denom
			lowDenom = coin.Denom
			seenDenoms[coin.Denom] = true
		}

		return nil
	}
}
