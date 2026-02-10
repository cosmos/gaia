// Package bank provides a custom wrapper around the SDK bank module
// that adds a quadratic gas surcharge and recipient limit for
// MsgMultiSend at the MsgServer level. This ensures all multi-send
// messages are subject to anti-spam controls regardless of their
// origin (user tx, ICA, wasm, authz, or any future mechanism).
package bank

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AppModuleWrapper wraps the standard bank module to intercept RegisterServices
type AppModuleWrapper struct {
	bank.AppModule
	keeper bankkeeper.Keeper
	config MultiSendConfig
}

// NewAppModuleWrapper creates a new wrapper for the bank module
func NewAppModuleWrapper(am bank.AppModule, keeper bankkeeper.Keeper, config MultiSendConfig) AppModuleWrapper {
	return AppModuleWrapper{
		AppModule: am,
		keeper:    keeper,
		config:    config,
	}
}

// RegisterServices overrides the standard bank module's RegisterServices to register our custom MsgServer
func (am AppModuleWrapper) RegisterServices(cfg module.Configurator) {
	// Register the QueryServer normally (delegating to the keeper)
	banktypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	// Create the standard MsgServer implementation from the keeper
	standardMsgServer := bankkeeper.NewMsgServerImpl(am.keeper)

	// Wrap the standard MsgServer with our custom logic
	wrappedMsgServer := NewMsgServerWrapper(standardMsgServer, am.config)

	// Register our wrapped MsgServer
	banktypes.RegisterMsgServer(cfg.MsgServer(), wrappedMsgServer)
}
