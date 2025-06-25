package telemetry

import "time"

type OtelConfig struct {
	Disable                 bool          `mapstructure:"disable"`
	CollectorEndpoint       string        `mapstructure:"collector-endpoint"`
	CollectorMetricsURLPath string        `mapstructure:"collector-metrics-url-path"`
	User                    string        `mapstructure:"user"`
	Token                   string        `mapstructure:"token"`
	PushInterval            time.Duration `mapstructure:"push-interval"`
}

var DefaultOtelConfig = OtelConfig{
	Disable:                 false,
	CollectorEndpoint:       "mimir.overseer.skip.build",
	CollectorMetricsURLPath: "/otlp/v1/metrics",
	User:                    "",
	Token:                   "",
	PushInterval:            10 * time.Second,
}

// LocalOtelConfig is the config needed to interact with the telemetry infrastructure in contrib/telemetry.
// You can use `make start-telemetry-server` to launch the local telemetry infra.
var LocalOtelConfig = OtelConfig{
	Disable:                 false,
	CollectorEndpoint:       "localhost:4318",
	CollectorMetricsURLPath: "/v1/metrics",
	User:                    "", // empty for local testing
	Token:                   "", // empty for local testing
	PushInterval:            10 * time.Second,
}

func OpenTelemetryTemplate() string {
	return `
###############################################################################
###                        OpenTelemetry Configuration                     ###
###############################################################################

[opentelemetry]
disable = "{{ .OpenTelemetry.Disable }}"

# OTLP collector endpoint
collector-endpoint = "{{ .OpenTelemetry.CollectorEndpoint }}"

# OTLP collector metrics URL path
collector-metrics-url-path = "{{ .OpenTelemetry.CollectorMetricsURLPath }}"

# OTLP user for authentication
user = "{{ .OpenTelemetry.User }}"

# OTLP token for authentication
token = "{{ .OpenTelemetry.Token }}"

# OTLP push interval
push-interval = "{{ .OpenTelemetry.PushInterval }}"

`
}
