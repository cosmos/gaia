package gaia_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgrade "github.com/cosmos/gaia/v23/app/upgrades/v23_1_1"
	"github.com/stretchr/testify/require"

	db "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	gaia "github.com/cosmos/gaia/v23/app"
	gaiahelpers "github.com/cosmos/gaia/v23/app/helpers"
)

type EmptyAppOptions struct{}

var emptyWasmOption []wasmkeeper.Option

func (ao EmptyAppOptions) Get(_ string) interface{} {
	return nil
}

func TestGaiaApp_BlockedModuleAccountAddrs(t *testing.T) {
	app := gaia.NewGaiaApp(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaia.DefaultNodeHome,
		EmptyAppOptions{},
		emptyWasmOption,
	)

	moduleAccountAddresses := app.ModuleAccountAddrs()
	blockedAddrs := app.BlockedModuleAccountAddrs(moduleAccountAddresses)

	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestGaiaApp_Preblock(t *testing.T) {
	app := gaia.NewGaiaApp(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaia.DefaultNodeHome,
		EmptyAppOptions{},
		emptyWasmOption,
	)
	var height int64 = 25283301
	granteeAddr, err := sdk.AccAddressFromBech32(upgrade.GranteeAddress)
	require.NoError(t, err)
	granterAddr, err := sdk.AccAddressFromBech32(app.GovKeeper.GetAuthority())
	require.NoError(t, err)
	ctx := app.NewUncachedContext(true, tmproto.Header{
		Height: height,
	})

	auth, _ := app.AuthzKeeper.GetAuthorization(
		ctx, granteeAddr,
		granterAddr,
		upgrade.IBCWasmMigrateTypeURL)
	require.Nil(t, auth)

	_, err = app.PreBlocker(ctx, &abci.RequestFinalizeBlock{Height: height})
	require.NoError(t, err)

	auth, _ = app.AuthzKeeper.GetAuthorization(
		ctx, granteeAddr,
		granterAddr,
		upgrade.IBCWasmMigrateTypeURL)
	require.NotNil(t, auth)

	resp, err := app.AuthzKeeper.GranteeGrants(ctx, &authz.QueryGranteeGrantsRequest{
		Grantee: upgrade.GranteeAddress,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Grants))
	require.Equal(t, granterAddr.String(), resp.Grants[0].Granter)
	require.Equal(t, granteeAddr.String(), resp.Grants[0].Grantee)
	require.Equal(t, upgrade.GrantExpiration, *resp.Grants[0].Expiration)

}

func TestGaiaApp_Export(t *testing.T) {
	app := gaiahelpers.Setup(t)
	_, err := app.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
