package gaia_test

import (
	"os"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	gaia "github.com/cosmos/gaia/v8/app"
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

	// TODO: Blocked on updating to v0.46.x
	// require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(grouptypes.ModuleName).String())
	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestGaiaApp_Export(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	app := gaia.NewGaiaApp(
		logger,
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaia.DefaultNodeHome,
		0,
		gaia.MakeTestEncodingConfig(),
		EmptyAppOptions{})
	_, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
