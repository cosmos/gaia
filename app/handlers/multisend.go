package handlers

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MultiSendConfig defines configuration for the MultiSend decorator
type MultiSendConfig struct {
	MaxRecipients int
	GasFactor     uint64 // Quadratic gas factor "A" in: Gas = A * N^2
}

// DefaultMultiSendConfig returns the default configuration
func DefaultMultiSendConfig() MultiSendConfig {
	return MultiSendConfig{
		MaxRecipients: 500,  // Limit to 500 recipients per transaction
		GasFactor:     1000, // Default quadratic factor
	}
}

// BankAppModuleWrapper wraps the standard bank module to intercept RegisterServices
type BankAppModuleWrapper struct {
	bank.AppModule
	keeper bankkeeper.Keeper
	config MultiSendConfig
}

// NewBankAppModuleWrapper creates a new wrapper for the bank module
func NewBankAppModuleWrapper(am bank.AppModule, keeper bankkeeper.Keeper, config MultiSendConfig) BankAppModuleWrapper {
	return BankAppModuleWrapper{
		AppModule: am,
		keeper:    keeper,
		config:    config,
	}
}

// RegisterServices overrides the standard bank module's RegisterServices to register our custom MsgServer
func (am BankAppModuleWrapper) RegisterServices(cfg module.Configurator) {
	// Register the QueryServer normally (delegating to the keeper)
	banktypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	// Create the standard MsgServer implementation from the keeper
	standardMsgServer := bankkeeper.NewMsgServerImpl(am.keeper)

	// Wrap the standard MsgServer with our custom logic
	wrappedMsgServer := NewMsgServerWrapper(standardMsgServer, am.config)

	// Register our wrapped MsgServer
	banktypes.RegisterMsgServer(cfg.MsgServer(), wrappedMsgServer)
}

// MsgServerWrapper wraps the standard bank MsgServer
type MsgServerWrapper struct {
	banktypes.MsgServer
	config MultiSendConfig
}

// NewMsgServerWrapper creates a new MsgServer wrapper
func NewMsgServerWrapper(keeper banktypes.MsgServer, config MultiSendConfig) MsgServerWrapper {
	return MsgServerWrapper{
		MsgServer: keeper,
		config:    config,
	}
}

// MultiSend intercepts and enforces controls on MsgMultiSend
func (s MsgServerWrapper) MultiSend(goCtx context.Context, msg *banktypes.MsgMultiSend) (*banktypes.MsgMultiSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check max recipients
	if len(msg.Outputs) > s.config.MaxRecipients {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "too many recipients in MultiSend: max %d, got %d", s.config.MaxRecipients, len(msg.Outputs))
	}

	// Apply Quadratic Gas Surcharge: A * N^2
	n := uint64(len(msg.Outputs))
	if n > 0 {
		surcharge := s.config.GasFactor * n * n
		ctx.GasMeter().ConsumeGas(surcharge, "MultiSend quadratic surcharge")
	}

	// Forward to the original implementation
	return s.MsgServer.MultiSend(goCtx, msg)
}
