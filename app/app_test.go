package gaia_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	db "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	gaia "github.com/cosmos/gaia/v18/app"
	gaiahelpers "github.com/cosmos/gaia/v18/app/helpers"
)

type EmptyAppOptions struct{}

func (ao EmptyAppOptions) Get(_ string) interface{} {
	return nil
}

func TestGaiaApp_BlockedModuleAccountAddrs(t *testing.T) {
	encConfig := gaia.RegisterEncodingConfig()
	app := gaia.NewGaiaApp(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaia.DefaultNodeHome,
		encConfig,
		EmptyAppOptions{},
	)

	moduleAccountAddresses := app.ModuleAccountAddrs()
	blockedAddrs := app.BlockedModuleAccountAddrs(moduleAccountAddresses)

	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestGaiaApp_Export(t *testing.T) {
	app := gaiahelpers.Setup(t)
	_, err := app.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
