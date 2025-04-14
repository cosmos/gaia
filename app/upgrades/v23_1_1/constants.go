package v23_1_1 //nolint:revive
import "time"

const (
	IBCWasmMigrateTypeURL = "/ibc.lightclients.wasm.v1.MsgMigrateContract"
	GranteeAddress        = "cosmos1raa4kyx5ypz75qqk3566c6slx2mw3qzs5ps5du"
)

// GrantExpiration on Apr 15th 2025, 15:00:00+00:00
var GrantExpiration = time.Date(2025, time.April, 15, 15, 0, 0, 0, time.UTC)
