package app

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulationOperations retrieves the simulation params from the provided file path
// and returns all the modules weighted operations
func SimulationOperations(app *GaiaApp, cdc *codec.Codec, config simulation.Config) []simulation.WeightedOperation {
	simState := module.SimulationState{
		AppParams: make(simulation.AppParams),
		Cdc:       cdc,
	}

	if config.ParamsFile != "" {
		bz, err := ioutil.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		app.cdc.MustUnmarshalJSON(bz, &simState.AppParams)
	}

	simState.ParamChanges = app.sm.GenerateParamChanges(config.Seed)
	simState.Contents = app.sm.GetProposalContents(simState)
	return app.sm.WeightedOperations(simState)
}
