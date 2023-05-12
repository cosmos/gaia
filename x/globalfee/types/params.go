package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
)

var (
	// ParamStoreKeyMinGasPrices store key
	ParamStoreKeyMinGasPrices                    = []byte("MinimumGasPricesParam")
	ParamStoreKeyBypassMinFeeMsgTypes            = []byte("BypassMinFeeMsgTypes")
	ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage = []byte("MaxTotalBypassMinFeeMsgGasUsage")

	// DefaultMinGasPrices is set at runtime to the staking token with zero amount i.e. "0uatom"
	// see DefaultZeroGlobalFee method in gaia/x/globalfee/ante/fee.go.
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

	return validateMaxTotalBypassMinFeeMsgGasUsage(p.MaxTotalBypassMinFeeMsgGasUsage)
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

func validateMinimumGasPrices(i interface{}) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected sdk.DecCoins", i)
	}

	dec := DecCoins(v)
	return dec.Validate()
}

type BypassMinFeeMsgTypes []string

// validateBypassMinFeeMsgTypes checks that bypass msg types aren't empty
func validateBypassMinFeeMsgTypes(i interface{}) error {
	bypassMinFeeMsgTypes, ok := i.([]string)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected []sdk.Msg", i)
	}

	for _, msgType := range bypassMinFeeMsgTypes {
		if msgType == "" {
			return fmt.Errorf("invalid empty bypass msg type")
		}

		if !strings.HasPrefix(msgType, sdk.MsgTypeURL(nil)) {
			return fmt.Errorf("invalid bypass msg type name %s", msgType)
		}
	}

	return nil
}

func validateMaxTotalBypassMinFeeMsgGasUsage(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type: %T, expected uint64", i)
	}

	return nil
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

	for i, coin := range coins {
		if seenDenoms[coin.Denom] {
			return fmt.Errorf("duplicate denomination %s", coin.Denom)
		}
		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return err
		}
		// skip the denom order check for the first denom in the coins list
		if i != 0 && coin.Denom <= lowDenom {
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
