package handlers_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gaia/v26/app/handlers"
	"github.com/stretchr/testify/require"
)

// MockMsgServer mocks basic MsgServer functionality
type MockMsgServer struct {
	banktypes.MsgServer
}

func (m MockMsgServer) MultiSend(ctx context.Context, msg *banktypes.MsgMultiSend) (*banktypes.MsgMultiSendResponse, error) {
	return &banktypes.MsgMultiSendResponse{}, nil
}

func TestMsgServerWrapper(t *testing.T) {
	mockHelper := MockMsgServer{}

	// Helper to create context with gas meter
	setupCtx := func() sdk.Context {
		return sdk.Context{}.WithGasMeter(storetypes.NewGasMeter(1000000000))
	}

	t.Run("MaxRecipients", func(t *testing.T) {
		config := handlers.MultiSendConfig{MaxRecipients: 2, GasFactor: 1000}
		wrapper := handlers.NewMsgServerWrapper(mockHelper, config)
		ctx := setupCtx()

		// 3 recipients > 2
		msg := &banktypes.MsgMultiSend{
			Outputs: make([]banktypes.Output, 3),
		}
		_, err := wrapper.MultiSend(sdk.WrapSDKContext(ctx), msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "too many recipients")
	})

	t.Run("GasSurcharge_DefaultFactor", func(t *testing.T) {
		config := handlers.MultiSendConfig{MaxRecipients: 100, GasFactor: 1000}
		wrapper := handlers.NewMsgServerWrapper(mockHelper, config)
		ctx := setupCtx()

		// 10 recipients -> 1000 * 10^2 = 100,000 gas
		msg := &banktypes.MsgMultiSend{
			Outputs: make([]banktypes.Output, 10),
		}

		gasBefore := ctx.GasMeter().GasConsumed()
		_, err := wrapper.MultiSend(sdk.WrapSDKContext(ctx), msg)
		require.NoError(t, err)
		gasAfter := ctx.GasMeter().GasConsumed()

		expectedSurcharge := uint64(1000 * 10 * 10)
		require.Equal(t, expectedSurcharge, gasAfter-gasBefore, "Should consume quadratic gas")
	})

	t.Run("GasSurcharge_CustomFactor", func(t *testing.T) {
		config := handlers.MultiSendConfig{MaxRecipients: 100, GasFactor: 500}
		wrapper := handlers.NewMsgServerWrapper(mockHelper, config)
		ctx := setupCtx()

		// 10 recipients -> 500 * 10^2 = 50,000 gas
		msg := &banktypes.MsgMultiSend{
			Outputs: make([]banktypes.Output, 10),
		}

		gasBefore := ctx.GasMeter().GasConsumed()
		_, err := wrapper.MultiSend(sdk.WrapSDKContext(ctx), msg)
		require.NoError(t, err)
		gasAfter := ctx.GasMeter().GasConsumed()

		expectedSurcharge := uint64(500 * 10 * 10)
		require.Equal(t, expectedSurcharge, gasAfter-gasBefore, "Should consume custom quadratic gas")
	})

	t.Run("PassThrough", func(t *testing.T) {
		config := handlers.MultiSendConfig{MaxRecipients: 50, GasFactor: 1000}
		wrapper := handlers.NewMsgServerWrapper(mockHelper, config)
		ctx := setupCtx()

		msg := &banktypes.MsgMultiSend{
			Outputs: make([]banktypes.Output, 1),
		}
		_, err := wrapper.MultiSend(sdk.WrapSDKContext(ctx), msg)
		require.NoError(t, err)
	})
}
