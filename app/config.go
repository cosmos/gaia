package gaia

import (
	evmserverconfig "github.com/cosmos/evm/server/config"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v25/telemetry"
)

type AppConfig struct {
	serverconfig.Config

	EVM     evmserverconfig.EVMConfig
	JSONRPC evmserverconfig.JSONRPCConfig
	TLS     evmserverconfig.TLSConfig

	Wasm wasmtypes.NodeConfig `mapstructure:"wasm"`

	OpenTelemetry telemetry.OtelConfig `mapstructure:"opentelemetry"`
}
