package app

import (
	"fmt"
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

// ExportStateToJSON util function to export the app state to JSON
func ExportStateToJSON(app *GaiaApp, path string) error {
	fmt.Println("exporting app state...")
	appState, _, err := app.ExportAppStateAndValidators(false, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(appState), 0644)
}
