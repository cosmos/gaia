package types

import (
	"fmt"

	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// ParamStoreKeyMinGasPrices store key
	ParamStoreKeyMinGasPrices                    = []byte("MinimumGasPricesParam")
	ParamStoreKeyBypassMinFeeMsgTypes            = []byte("BypassMinFeeMsgTypes")
	ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage = []byte("MaxTotalBypassMinFeeMsgGasUsage")

	DefaultMinGasPrices         = sdk.DecCoins{}
	DefaultBypassMinFeeMsgTypes = []string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeoutOnClose{}),
	}

	// maxTotalBypassMinFeeMsgGasUsage is the allowed maximum gas usage
	// for all the bypass msgs in a transactions.
	// A transaction that contains only bypass message types and the gas usage does not
	// exceed maxTotalBypassMinFeeMsgGasUsage can be accepted with a zero fee.
	// For details, see gaiafeeante.NewFeeDecorator()
	DefaultmaxTotalBypassMinFeeMsgGasUsage uint64 = 1_000_000
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinimumGasPrices:                DefaultMinGasPrices,
		BypassMinFeeMsgTypes:            DefaultBypassMinFeeMsgTypes,
		MaxTotalBypassMinFeeMsgGasUsage: DefaultmaxTotalBypassMinFeeMsgGasUsage,
	}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ValidateBasic performs basic validation.
func (p Params) ValidateBasic() error {
	if err := validateMinimumGasPrices(p.MinimumGasPrices); err != nil {
		return err
	}

	if err := validateBypassMinFeeMsgTypes(p.BypassMinFeeMsgTypes); err != nil {
		return err
	}

	if err := validateMaxTotalBypassMinFeeMsgGasUsage(p.MaxTotalBypassMinFeeMsgGasUsage); err != nil {
		return err
	}

	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			ParamStoreKeyMinGasPrices, &p.MinimumGasPrices, validateMinimumGasPrices,
		),
		paramtypes.NewParamSetPair(
			ParamStoreKeyBypassMinFeeMsgTypes, &p.BypassMinFeeMsgTypes, validateBypassMinFeeMsgTypes,
		),
		paramtypes.NewParamSetPair(
			ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &p.MaxTotalBypassMinFeeMsgGasUsage, validateMaxTotalBypassMinFeeMsgGasUsage,
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

// todo check if correct?
func validateBypassMinFeeMsgTypes(i interface{}) error {
	_, ok := i.([]sdk.Msg)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected []sdk.Msg", i)
	}

	return nil
}

func validateMaxTotalBypassMinFeeMsgGasUsage(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected uint64", i)
	}

	if v < 0 {
		return fmt.Errorf("gas usage %s is negtive", v)
	}

	return nil
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
