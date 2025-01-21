package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/cosmos/gaia/v23/app"
	"github.com/cosmos/gaia/v23/cmd/gaiad/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
