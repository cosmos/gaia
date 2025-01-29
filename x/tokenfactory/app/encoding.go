package app

import (
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/gaia/v23/x/tokenfactory/app/params"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

// MakeEncodingConfig creates a new EncodingConfig with all modules registered. For testing only
func MakeEncodingConfig(tb testing.TB) params.EncodingConfig {
	tb.Helper()
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := NewApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(tb.TempDir()), []wasmkeeper.Option{})
	return makeEncodingConfig(tempApp)
}

func makeEncodingConfig(tempApp *TokenFactoryApp) params.EncodingConfig {
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}
	return encodingConfig
}
