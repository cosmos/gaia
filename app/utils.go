//nolint
package app

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
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
	flagPeriodValue      uint
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

// NewGaiaAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewGaiaAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *GaiaApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewGaiaApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.GetKey(baseapp.MainStoreKey), gapp.GetKey(staking.StoreKey), gapp.stakingKeeper
}
