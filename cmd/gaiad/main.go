package main

import (
	"os"

	"github.com/cosmos/gaia/cmd/gaiad/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
