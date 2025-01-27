package bindings_test

import (
	"os"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/strangelove-ventures/tokenfactory/app"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func CreateTestInput(t *testing.T) (*app.TokenFactoryApp, sdk.Context) {
	ctx, chain := app.Setup(t)
	return chain, sdk.UnwrapSDKContext(ctx)
}

func FundAccount(t *testing.T, ctx sdk.Context, app *app.TokenFactoryApp, acct sdk.AccAddress) {
	err := banktestutil.FundAccount(ctx, app.BankKeeper, acct, sdk.NewCoins(
		sdk.NewCoin("uosmo", sdkmath.NewInt(10000000000)),
	))
	require.NoError(t, err)
}

// we need to make this deterministic (same every test run), as content might affect gas costs
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress() string {
	return RandomAccountAddress().String()
}

func storeReflectCode(t *testing.T, ctx sdk.Context, app *app.TokenFactoryApp, addr sdk.AccAddress) uint64 {
	wasmCode, err := os.ReadFile("./testdata/token_reflect.wasm")
	require.NoError(t, err)

	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	codeID, _, err := contractKeeper.Create(ctx, addr, wasmCode, nil)
	require.NoError(t, err)

	return codeID
}

func instantiateReflectContract(t *testing.T, ctx sdk.Context, app *app.TokenFactoryApp, funder sdk.AccAddress) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	codeID := uint64(1)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	require.NoError(t, err)

	return addr
}

func fundAccount(t *testing.T, ctx sdk.Context, app *app.TokenFactoryApp, addr sdk.AccAddress, coins sdk.Coins) {
	err := banktestutil.FundAccount(
		ctx,
		app.BankKeeper,
		addr,
		coins,
	)
	require.NoError(t, err)
}

func SetupCustomApp(t *testing.T, addr sdk.AccAddress) (*app.TokenFactoryApp, sdk.Context) {
	app, ctx := CreateTestInput(t)
	wasmKeeper := app.WasmKeeper

	storeReflectCode(t, ctx, app, addr)

	cInfo := wasmKeeper.GetCodeInfo(ctx, 1)
	require.NotNil(t, cInfo)

	return app, ctx
}
