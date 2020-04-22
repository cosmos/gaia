package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/std"
)

// GenesisState defines a type alias for the Gaia genesis application state.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	cdc := std.MakeCodec(ModuleBasics)
	return ModuleBasics.DefaultGenesis(cdc)
}
