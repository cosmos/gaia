package wasm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiawasm "github.com/cosmos/gaia/v28/x/wasm"
)

// mockMsgServer mocks the underlying wasmd MsgServer for passthrough testing.
type mockMsgServer struct {
	wasmtypes.MsgServer
	executeContractCalled bool
}

func (m *mockMsgServer) ExecuteContract(_ context.Context, _ *wasmtypes.MsgExecuteContract) (*wasmtypes.MsgExecuteContractResponse, error) {
	m.executeContractCalled = true
	return &wasmtypes.MsgExecuteContractResponse{}, nil
}

// TestMsgServerBlocksStoreAndInstantiate tests that the decorated MsgServer
// rejects the five blocked messages outright, regardless of message
// contents — the underlying (mocked) MsgServer is never called for these.
func TestMsgServerBlocksStoreAndInstantiate(t *testing.T) {
	msgServer := gaiawasm.NewMsgServerImpl(&mockMsgServer{})
	ctx := context.Background()

	t.Run("StoreCode", func(t *testing.T) {
		_, err := msgServer.StoreCode(ctx, &wasmtypes.MsgStoreCode{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "MsgStoreCode is temporarily disabled")
	})

	t.Run("InstantiateContract", func(t *testing.T) {
		_, err := msgServer.InstantiateContract(ctx, &wasmtypes.MsgInstantiateContract{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "MsgInstantiateContract is temporarily disabled")
	})

	t.Run("InstantiateContract2", func(t *testing.T) {
		_, err := msgServer.InstantiateContract2(ctx, &wasmtypes.MsgInstantiateContract2{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "MsgInstantiateContract2 is temporarily disabled")
	})

	t.Run("StoreAndInstantiateContract", func(t *testing.T) {
		_, err := msgServer.StoreAndInstantiateContract(ctx, &wasmtypes.MsgStoreAndInstantiateContract{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "MsgStoreAndInstantiateContract is temporarily disabled")
	})

	t.Run("StoreAndMigrateContract", func(t *testing.T) {
		_, err := msgServer.StoreAndMigrateContract(ctx, &wasmtypes.MsgStoreAndMigrateContract{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "MsgStoreAndMigrateContract is temporarily disabled")
	})
}

// TestMsgServerPassthroughMethods tests that non-blocked methods are
// properly passed through to the underlying MsgServer.
func TestMsgServerPassthroughMethods(t *testing.T) {
	mock := &mockMsgServer{}
	msgServer := gaiawasm.NewMsgServerImpl(mock)
	ctx := context.Background()

	_, err := msgServer.ExecuteContract(ctx, &wasmtypes.MsgExecuteContract{})
	require.NoError(t, err)
	require.True(t, mock.executeContractCalled, "ExecuteContract should be passed through to the underlying MsgServer")
}
