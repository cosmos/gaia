package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v26/ante"
	"github.com/cosmos/gaia/v26/app/helpers"
)

func TestProviderDecorator_BlockMsgCreateConsumer(t *testing.T) {
	app := helpers.Setup(t)
	ctx := app.NewContext(false)
	cdc := app.AppCodec()

	decorator := ante.NewProviderDecorator(cdc)

	// Create a mock MsgCreateConsumer
	msgCreateConsumer := &providertypes.MsgCreateConsumer{
		Submitter: "cosmos1abcdef1234567890abcdef1234567890abcdef",
	}

	// Create a transaction with the MsgCreateConsumer
	txBuilder := app.GetTxConfig().NewTxBuilder()
	err := txBuilder.SetMsgs(msgCreateConsumer)
	require.NoError(t, err)
	tx := txBuilder.GetTx()

	// Test that the decorator blocks the MsgCreateConsumer
	_, err = decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		require.Fail(t, "next handler should not be called")
		return ctx, nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "MsgCreateConsumer is disabled")
}

func TestProviderDecorator_AllowOtherMessages(t *testing.T) {
	app := helpers.Setup(t)
	ctx := app.NewContext(false)
	cdc := app.AppCodec()

	decorator := ante.NewProviderDecorator(cdc)

	// Create a normal bank message (should be allowed)
	msgSend := &banktypes.MsgSend{
		FromAddress: "cosmos1abcdef1234567890abcdef1234567890abcdef",
		ToAddress:   "cosmos1fedcba0987654321fedcba0987654321fedcba",
		Amount:      sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(100))),
	}

	// Create a transaction with the MsgSend
	txBuilder := app.GetTxConfig().NewTxBuilder()
	err := txBuilder.SetMsgs(msgSend)
	require.NoError(t, err)
	tx := txBuilder.GetTx()

	// Test that the decorator allows other messages and calls next handler
	nextCalled := false
	_, err = decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	})

	require.NoError(t, err)
	require.True(t, nextCalled, "next handler should have been called")
}

func TestProviderDecorator_BlockAuthzWrappedMsgCreateConsumer(t *testing.T) {
	app := helpers.Setup(t)
	ctx := app.NewContext(false)
	cdc := app.AppCodec()

	decorator := ante.NewProviderDecorator(cdc)

	// Create a mock MsgCreateConsumer
	msgCreateConsumer := &providertypes.MsgCreateConsumer{
		Submitter: "cosmos1abcdef1234567890abcdef1234567890abcdef",
	}

	// Wrap it in an authz MsgExec
	anyMsg, err := codectypes.NewAnyWithValue(msgCreateConsumer)
	require.NoError(t, err)

	msgExec := &authz.MsgExec{
		Grantee: "cosmos1fedcba0987654321fedcba0987654321fedcba",
		Msgs:    []*codectypes.Any{anyMsg},
	}

	// Create a transaction with the wrapped message
	txBuilder := app.GetTxConfig().NewTxBuilder()
	err = txBuilder.SetMsgs(msgExec)
	require.NoError(t, err)
	tx := txBuilder.GetTx()

	// Test that the decorator blocks the wrapped MsgCreateConsumer
	_, err = decorator.AnteHandle(ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		require.Fail(t, "next handler should not be called")
		return ctx, nil
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "MsgCreateConsumer is disabled")
}
