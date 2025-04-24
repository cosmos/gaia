package gaia

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	gaiatypes "github.com/cosmos/gaia/v23/types"
)

// TODO eric fixme
var chainID = "foobar"

// EVMOptionsFn defines a function type for setting app options specifically for
// the app. The function should receive the chainID and return an error if
// any.
type EVMOptionsFn func() error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEVMOptions() error {
	return nil
}

// EVMAppOptions performs setup of the global configuration
// for the chain.
func EVMAppOptions() error {
	// set the denom info for the chain
	if err := setBaseDenom(gaiatypes.UAtomCoinInfo); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	// TODO eric -- pull the chain ID from state somewhere
	ethCfg := evmtypes.DefaultChainConfig(chainID)

	return evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		// NOTE: we're using the 18 decimals
		WithEVMCoinInfo(baseDenom, uint8(gaiatypes.UAtomCoinInfo.Decimals)).
		Configure()
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain.
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}

	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
