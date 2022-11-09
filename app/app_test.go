package gaia_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	gaia "github.com/cosmos/gaia/v8/app"
	gaiahelpers "github.com/cosmos/gaia/v8/app/helpers"
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
	blockedAddrs := app.BlockedModuleAccountAddrs()

	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestGaiaApp_Export(t *testing.T) {
	app := gaiahelpers.Setup(t)

	_, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
