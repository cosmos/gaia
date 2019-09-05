//nolint
package app

import (
	"fmt"
	"io/ioutil"
)

//---------------------------------------------------------------------
// Flags

// List of available flags for the simulator
var (
	flagGenesisFileValue        string
	flagParamsFileValue         string
	flagExportParamsPathValue   string
	flagExportParamsHeightValue int
	flagExportStatePathValue    string
	flagExportStatsPathValue    string
	flagSeedValue               int64
	flagInitialBlockHeightValue int
	flagNumBlocksValue          int
	flagBlockSizeValue          int
	flagLeanValue               bool
	flagCommitValue             bool
	flagOnOperationValue        bool // TODO: Remove in favor of binary search for invariant violation
	flagAllInvariantsValue      bool

	flagEnabledValue     bool
	flagVerboseValue     bool
	flagPeriodValue      int
	flagGenesisTimeValue int64
)

// ExportStateToJSON util function to export the app state to JSON
func ExportStateToJSON(app *GaiaApp, path string) error {
	fmt.Println("exporting app state...")
	appState, _, err := app.ExportAppStateAndValidators(false, nil)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(appState), 0644)
}
