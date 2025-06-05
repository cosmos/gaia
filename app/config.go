package gaia

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	evmhd "github.com/cosmos/evm/crypto/hd"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gaiatypes "github.com/cosmos/gaia/v25/types"
)

var (
	SupportedKeyAlgorithms = keyring.SigningAlgoList{hd.Secp256k1, evmhd.EthSecp256k1}
	sealed                 = false
)

func KeyringOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = SupportedKeyAlgorithms
		options.SupportedAlgosLedger = SupportedKeyAlgorithms
	}
}

// EVMOptionsFn defines a function type for setting app options specifically for
// the app. The function should receive the chainID and return an error if
// any.
type EVMOptionsFn func(uint64) error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEVMOptions(uint64) error {
	return nil
}

// EVMAppOptions performs setup of the global configuration
// for the chain.
func EVMAppOptions(chainID uint64) error {
	if sealed {
		return nil
	}
	sealed = true
	// set the denom info for the chain
	if err := setBaseDenom(gaiatypes.UAtomCoinInfo); err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	return evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		// NOTE: we're using the 18 decimals
		WithEVMCoinInfo(gaiatypes.UAtomCoinInfo).
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
