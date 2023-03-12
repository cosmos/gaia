package gaia_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gaia "github.com/cosmos/gaia/v9/app"
	gaiahelpers "github.com/cosmos/gaia/v9/app/helpers"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"
)

type EmptyAppOptions struct{}

func (ao EmptyAppOptions) Get(o string) interface{} {
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
		0,
		gaia.MakeTestEncodingConfig(),
		EmptyAppOptions{},
	)
	moduleAccountAddresses := app.ModuleAccountAddrs()
	blockedAddrs := app.BlockedModuleAccountAddrs(moduleAccountAddresses)

	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestGaiaApp_Export(t *testing.T) {
	app := gaiahelpers.Setup(t)
	_, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
