# Gaia Telemetry

Starting with v26.0.0, Gaia validator nodes within the bonded validator set will automatically emit telemetry data to help Interchain Labs monitor network health and debug potential network issues. 
Telemetry is **opt-out** and can be disabled by editing your `app.toml`:

```markdown app.toml
[opentelemetry]
disable = true
```

To change the node's frequency of pushing telemetry data, add the following to the `opentelemetry` configuration in `app.toml`:

```markdown app.toml
[opentelemetry]
push-interval = "25s"
```


## What Data Is Collected?

Gaia nodes will emit telemetry data collected from Cosmos SDK, CometBFT, and the Go Runtime. The following information about the validator node is attached to this data: moniker and binary version. 
Follow the links below to see what telemetry data is emitted from these services.

- Cosmos SDK: https://docs.cosmos.network/main/learn/advanced/telemetry
- CometBFT: https://docs.cometbft.com/main/explanation/core/metrics
- Go Runtime (via OpenTelemetry): https://opentelemetry.io/docs/specs/semconv/runtime/go-metrics/
