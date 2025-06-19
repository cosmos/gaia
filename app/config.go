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

func OpenTelemetryTemplate() string {
	return `
###############################################################################
###                        OpenTelemetry Configuration                     ###
###############################################################################

[opentelemetry]
# OTLP collector endpoint
otlp-collector-endpoint = "{{ .OpenTelemetry.OtlpCollectorEndpoint }}"

# OTLP collector metrics URL path
otlp-collector-metrics-url-path = "{{ .OpenTelemetry.OtlpCollectorMetricsURLPath }}"

# OTLP user for authentication
otlp-user = "{{ .OpenTelemetry.OtlpUser }}"

# OTLP token for authentication
otlp-token = "{{ .OpenTelemetry.OtlpToken }}"

# OTLP service name
otlp-service-name = "{{ .OpenTelemetry.OtlpServiceName }}"

# OTLP push interval
otlp-push-interval = "{{ .OpenTelemetry.OtlpPushInterval }}"

`
}
