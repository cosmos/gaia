package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Get flags every time the simulator is run
func init() {
	simapp.GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
	Prefixes [][]byte
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	appState, _, consensusParams, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, newApp.Name())

	var genesisState GenesisState
	err = app.Codec().UnmarshalJSON(appState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, app.cdc, genesisState)
	newApp.StoreConsensusParams(ctxB, consensusParams)

	fmt.Printf("comparing stores...\n")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[auth.StoreKey], newApp.keys[auth.StoreKey], [][]byte{}},
		{app.keys[staking.StoreKey], newApp.keys[staking.StoreKey],
			[][]byte{
				staking.UnbondingQueueKey, staking.RedelegationQueueKey, staking.ValidatorQueueKey,
				staking.HistoricalInfoKey,
			}}, // ordering may change but it doesn't matter
		{app.keys[slashing.StoreKey], newApp.keys[slashing.StoreKey], [][]byte{}},
		{app.keys[mint.StoreKey], newApp.keys[mint.StoreKey], [][]byte{}},
		{app.keys[distr.StoreKey], newApp.keys[distr.StoreKey], [][]byte{}},
		{app.keys[bank.StoreKey], newApp.keys[bank.StoreKey], [][]byte{bank.BalancesPrefix}},
		{app.keys[paramtypes.StoreKey], newApp.keys[paramtypes.StoreKey], [][]byte{}},
		{app.keys[gov.StoreKey], newApp.keys[gov.StoreKey], [][]byte{}},
		{app.keys[evidence.StoreKey], newApp.keys[evidence.StoreKey], [][]byte{}},
		{app.keys[capability.StoreKey], newApp.keys[capability.StoreKey], [][]byte{}},
	}

	for _, skp := range storeKeysPrefixes {
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		require.Equal(t, len(failedKVAs), 0, simapp.GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	appState, _, _, err := app.ExportAppStateAndValidators(true, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, newApp.Name())

	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: appState,
	})

	_, _, err = simulation.SimulateFromSeed(
		t, os.Stdout, newApp.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(newApp, newApp.Codec(), config),
		newApp.ModuleAccountAddrs(), config,
	)
	require.NoError(t, err)
}

func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = helpers.SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = log.TestingLogger()
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()

			app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", interBlockCacheOpt())

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
				simapp.SimulationOperations(app, app.Codec(), config),
				app.ModuleAccountAddrs(), config,
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
