package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v28/ante"
	"github.com/cosmos/gaia/v28/app/helpers"
)

// TestWasmDisableDecoratorValidateMsgs tests that the decorator rejects
// exactly the five blocked wasm messages, and leaves everything else
// (including MsgMigrateContract, which is deliberately not blocked) alone.
func TestWasmDisableDecoratorValidateMsgs(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	decorator := ante.NewWasmDisableDecorator(gaiaApp.AppCodec())

	tests := []struct {
		name       string
		msg        sdk.Msg
		expectPass bool
		errContain string
	}{
		{
			name:       "MsgStoreCode is blocked",
			msg:        &wasmtypes.MsgStoreCode{},
			expectPass: false,
			errContain: "MsgStoreCode is temporarily disabled",
		},
		{
			name:       "MsgInstantiateContract is blocked",
			msg:        &wasmtypes.MsgInstantiateContract{},
			expectPass: false,
			errContain: "MsgInstantiateContract is temporarily disabled",
		},
		{
			name:       "MsgInstantiateContract2 is blocked",
			msg:        &wasmtypes.MsgInstantiateContract2{},
			expectPass: false,
			errContain: "MsgInstantiateContract2 is temporarily disabled",
		},
		{
			name:       "MsgStoreAndInstantiateContract is blocked",
			msg:        &wasmtypes.MsgStoreAndInstantiateContract{},
			expectPass: false,
			errContain: "MsgStoreAndInstantiateContract is temporarily disabled",
		},
		{
			name:       "MsgStoreAndMigrateContract is blocked",
			msg:        &wasmtypes.MsgStoreAndMigrateContract{},
			expectPass: false,
			errContain: "MsgStoreAndMigrateContract is temporarily disabled",
		},
		{
			name:       "MsgMigrateContract is not blocked",
			msg:        &wasmtypes.MsgMigrateContract{},
			expectPass: true,
		},
		{
			name:       "MsgExecuteContract is not blocked",
			msg:        &wasmtypes.MsgExecuteContract{},
			expectPass: true,
		},
		{
			name:       "unrelated messages are not blocked",
			msg:        &banktypes.MsgSend{},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := decorator.ValidateMsgs([]sdk.Msg{tc.msg})
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContain)
			}
		})
	}
}

// TestWasmDisableDecoratorMsgExec tests that a blocked message is still
// caught when wrapped inside one or more authz.MsgExec, and that nesting it
// deeper than the decorator's recursion limit fails closed (rejected for a
// different reason — too many wrapped messages — rather than being let
// through unchecked).
func TestWasmDisableDecoratorMsgExec(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	decorator := ante.NewWasmDisableDecorator(gaiaApp.AppCodec())

	wrap := func(t *testing.T, msg sdk.Msg, depth int) sdk.Msg {
		t.Helper()

		wrapped := msg
		for i := 0; i < depth; i++ {
			anyMsg, err := codectypes.NewAnyWithValue(wrapped)
			require.NoError(t, err)
			wrapped = &authz.MsgExec{
				Grantee: sdk.AccAddress{}.String(),
				Msgs:    []*codectypes.Any{anyMsg},
			}
		}
		return wrapped
	}

	t.Run("blocked message wrapped once is still caught", func(t *testing.T) {
		err := decorator.ValidateMsgs([]sdk.Msg{wrap(t, &wasmtypes.MsgStoreCode{}, 1)})
		require.Error(t, err)
		require.ErrorContains(t, err, "MsgStoreCode is temporarily disabled")
	})

	t.Run("blocked message wrapped well under the depth limit is still caught", func(t *testing.T) {
		err := decorator.ValidateMsgs([]sdk.Msg{wrap(t, &wasmtypes.MsgInstantiateContract{}, 5)})
		require.Error(t, err)
		require.ErrorContains(t, err, "MsgInstantiateContract is temporarily disabled")
	})

	t.Run("unrelated message wrapped is not blocked", func(t *testing.T) {
		err := decorator.ValidateMsgs([]sdk.Msg{wrap(t, &banktypes.MsgSend{}, 3)})
		require.NoError(t, err)
	})

	t.Run("nesting deeper than the recursion limit fails closed", func(t *testing.T) {
		// Exceeds the ante package's internal maxWrappedMessageDepth (20):
		// rejected before the wrapped message is ever unwrapped/inspected,
		// so this is never a way to smuggle a blocked message through.
		err := decorator.ValidateMsgs([]sdk.Msg{wrap(t, &wasmtypes.MsgStoreCode{}, 25)})
		require.Error(t, err)
		require.ErrorContains(t, err, "too many wrapped sdk messages")
	})
}

// TestWasmDisableDecoratorAnteHandleSimulate tests that the check is
// skipped during simulation (e.g. gas estimation), so only the real
// broadcast is rejected.
func TestWasmDisableDecoratorAnteHandleSimulate(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	decorator := ante.NewWasmDisableDecorator(gaiaApp.AppCodec())

	txBuilder := gaiaApp.GetTxConfig().NewTxBuilder()
	require.NoError(t, txBuilder.SetMsgs(&wasmtypes.MsgStoreCode{}))
	tx := txBuilder.GetTx()

	nextCalled := false
	next := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { //nolint:unparam // sdk.AnteHandler requires an error return, and this test stub always succeeds.
		nextCalled = true
		return ctx, nil
	}

	t.Run("simulate=true skips the check", func(t *testing.T) {
		nextCalled = false
		_, err := decorator.AnteHandle(ctx, tx, true, next)
		require.NoError(t, err)
		require.True(t, nextCalled, "next should be called during simulation even for a blocked message")
	})

	t.Run("simulate=false rejects the blocked message", func(t *testing.T) {
		nextCalled = false
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		require.Error(t, err)
		require.ErrorContains(t, err, "MsgStoreCode is temporarily disabled")
		require.False(t, nextCalled, "next should not be called for a blocked message")
	})
}
