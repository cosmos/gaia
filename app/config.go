package gaia

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

type AppConfig struct {
	serverconfig.Config

	Wasm wasmtypes.NodeConfig `mapstructure:"wasm"`
}
