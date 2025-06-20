package gaia

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/gaia/v25/telemetry"
)

type GaiaAppConfig struct {
	serverconfig.Config

	Wasm wasmtypes.NodeConfig `mapstructure:"wasm"`

	OpenTelemetry telemetry.OtelConfig `mapstructure:"opentelemetry"`
}
