package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/cosmos/gaia/v12/app"
	"github.com/cosmos/gaia/v12/cmd/gaiad/cmd"
)

func TestRootCmdConfig(t *testing.T) {
	rootCmd, _ := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"config",          // Test the config cmd
		"keyring-backend", // key
		"test",            // value
	})

	require.NoError(t, svrcmd.Execute(rootCmd, app.DefaultNodeHome))
}
