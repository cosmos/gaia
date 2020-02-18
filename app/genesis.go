package app

import (
	"encoding/json"

	appcodec "github.com/cosmos/gaia/app/codec"
)

// GenesisState defines a type alias for the Gaia genesis application state.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() GenesisState {
	cdc := appcodec.MakeCodec(ModuleBasics)
	return ModuleBasics.DefaultGenesis(cdc)
}
