package bank

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		MaxRecipients: 500, // Limit to 500 recipients per transaction
		GasFactor:     300, // Default quadratic factor
	}
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
