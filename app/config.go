package gaia

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v26/telemetry"
)

type AppConfig struct {
	serverconfig.Config

	Wasm wasmtypes.NodeConfig `mapstructure:"wasm"`

	OpenTelemetry telemetry.OtelConfig `mapstructure:"opentelemetry"`
}
