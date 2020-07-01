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
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsim "github.com/cosmos/cosmos-sdk/x/params/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Get flags every time the simulator is run
func init() {
	flag.StringVar(&genesisFile, "Genesis", "", "custom simulation genesis file; cannot be used with params file")
	flag.StringVar(&paramsFile, "Params", "", "custom simulation params file which overrides any random params; cannot be used with genesis")
	flag.StringVar(&exportParamsPath, "ExportParamsPath", "", "custom file path to save the exported params JSON")
	flag.IntVar(&exportParamsHeight, "ExportParamsHeight", 0, "height to which export the randomly generated params")
	flag.StringVar(&exportStatePath, "ExportStatePath", "", "custom file path to save the exported app state JSON")
	flag.StringVar(&exportStatsPath, "ExportStatsPath", "", "custom file path to save the exported simulation statistics JSON")
	flag.Int64Var(&seed, "Seed", 42, "simulation random seed")
	flag.IntVar(&initialBlockHeight, "InitialBlockHeight", 1, "initial block to start the simulation")
	flag.IntVar(&numBlocks, "NumBlocks", 500, "number of new blocks to simulate from the initial block height")
	flag.IntVar(&blockSize, "BlockSize", 200, "operations per block")
	flag.BoolVar(&enabled, "Enabled", false, "enable the simulation")
	flag.BoolVar(&verbose, "Verbose", false, "verbose log output")
	flag.BoolVar(&lean, "Lean", false, "lean simulation log output")
	flag.BoolVar(&commit, "Commit", false, "have the simulation commit")
	flag.IntVar(&period, "Period", 1, "run slow invariants only once every period assertions")
	flag.BoolVar(&onOperation, "SimulateEveryOperation", false, "run slow invariants every operation")
	flag.BoolVar(&allInvariants, "PrintAllInvariants", false, "print all invariants if a broken invariant is found")
	flag.Int64Var(&genesisTime, "GenesisTime", 0, "override genesis UNIX time instead of using a random UNIX time")
}

// helper function for populating input for SimulateFromSeed
func getSimulateFromSeedInput(tb testing.TB, w io.Writer, app *GaiaApp) (
	testing.TB, io.Writer, *baseapp.BaseApp, simulation.AppStateFn, int64,
	simulation.WeightedOperations, sdk.Invariants, int, int, int, int, string,
	bool, bool, bool, bool, bool, map[string]bool) {

	exportParams := exportParamsPath != ""

	return tb, w, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app), invariants(app),
		initialBlockHeight, numBlocks, exportParamsHeight, blockSize,
		exportStatsPath, exportParams, commit, lean, onOperation, allInvariants, app.ModuleAccountAddrs()
}

func appStateFn(
	r *rand.Rand, accs []simulation.Account,
) (appState json.RawMessage, simAccs []simulation.Account, chainID string, genesisTimestamp time.Time) {

	cdc := MakeCodec()

	if genesisTime == 0 {
		genesisTimestamp = simulation.RandTimestamp(r)
	} else {
		genesisTimestamp = time.Unix(genesisTime, 0)
	}

	switch {
	case paramsFile != "" && genesisFile != "":
		panic("cannot provide both a genesis file and a params file")

	case genesisFile != "":
		genesisDoc, accounts := simapp.AppStateFromGenesisFileFn(r)

		if genesisTime == 0 {
			// use genesis timestamp if no custom timestamp is provided (i.e no random timestamp)
			genesisTimestamp = genesisDoc.GenesisTime
		}

		appState = genesisDoc.AppState
		chainID = genesisDoc.ChainID
		simAccs = accounts

	case paramsFile != "":
		appParams := make(simulation.AppParams)
		bz, err := ioutil.ReadFile(paramsFile)
		if err != nil {
			panic(err)
		}

		cdc.MustUnmarshalJSON(bz, &appParams)
		appState, simAccs, chainID = appStateRandomizedFn(r, accs, genesisTimestamp, appParams)

	default:
		appParams := make(simulation.AppParams)
		appState, simAccs, chainID = appStateRandomizedFn(r, accs, genesisTimestamp, appParams)
	}

	return appState, simAccs, chainID, genesisTimestamp
}

// TODO refactor out random initialization code to the modules
func appStateRandomizedFn(
	r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time, appParams simulation.AppParams,
) (json.RawMessage, []simulation.Account, string) {

	cdc := MakeCodec()
	genesisState := simapp.NewDefaultGenesisState()

	var (
		amount             int64
		numInitiallyBonded int64
	)

	appParams.GetOrGenerate(cdc, StakePerAccount, &amount, r,
		func(r *rand.Rand) { amount = int64(r.Intn(1e12)) })
	appParams.GetOrGenerate(cdc, InitiallyBondedValidators, &amount, r,
		func(r *rand.Rand) { numInitiallyBonded = int64(r.Intn(250)) })

	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%v",
  initially_bonded_validators: "%v"
}
`, amount, numInitiallyBonded,
	)

	simapp.GenGenesisAccounts(cdc, r, accs, genesisTimestamp, amount, numInitiallyBonded, genesisState)
	simapp.GenAuthGenesisState(cdc, r, appParams, genesisState)
	simapp.GenBankGenesisState(cdc, r, appParams, genesisState)
	simapp.GenSupplyGenesisState(cdc, amount, numInitiallyBonded, int64(len(accs)), genesisState)
	simapp.GenGovGenesisState(cdc, r, appParams, genesisState)
	simapp.GenMintGenesisState(cdc, r, appParams, genesisState)
	simapp.GenDistrGenesisState(cdc, r, appParams, genesisState)
	stakingGen := simapp.GenStakingGenesisState(cdc, r, accs, amount, numAccs, numInitiallyBonded, appParams, genesisState)
	simapp.GenSlashingGenesisState(cdc, r, stakingGen, appParams, genesisState)

	appState, err := MakeCodec().MarshalJSON(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs, "simulation"
}

func testAndRunTxs(app *GaiaApp) []simulation.WeightedOperation {
	cdc := MakeCodec()
	ap := make(simulation.AppParams)

	if paramsFile != "" {
		bz, err := ioutil.ReadFile(paramsFile)
		if err != nil {
			panic(err)
		}

		cdc.MustUnmarshalJSON(bz, &ap)
	}

	return []simulation.WeightedOperation{
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightDeductFee, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			authsim.SimulateDeductFee(app.accountKeeper, app.supplyKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgSend, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			bank.SimulateMsgSend(app.accountKeeper, app.bankKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSingleInputMsgMultiSend, &v, nil,
					func(_ *rand.Rand) {
						v = 10
					})
				return v
			}(nil),
			bank.SimulateSingleInputMsgMultiSend(app.accountKeeper, app.bankKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgSetWithdrawAddress, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsim.SimulateMsgSetWithdrawAddress(app.accountKeeper, app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgWithdrawDelegationReward, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsim.SimulateMsgWithdrawDelegatorReward(app.accountKeeper, app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgWithdrawValidatorCommission, &v, nil,
					func(_ *rand.Rand) {
						v = 50
					})
				return v
			}(nil),
			distrsim.SimulateMsgWithdrawValidatorCommission(app.accountKeeper, app.distrKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingTextProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, govsim.SimulateTextProposalContent),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingCommunitySpendProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, distrsim.SimulateCommunityPoolSpendProposalContent(app.distrKeeper)),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightSubmitVotingSlashingParamChangeProposal, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, paramsim.SimulateParamChangeProposalContent),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgDeposit, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			govsim.SimulateMsgDeposit(app.govKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsim.SimulateMsgCreateValidator(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgEditValidator, &v, nil,
					func(_ *rand.Rand) {
						v = 5
					})
				return v
			}(nil),
			stakingsim.SimulateMsgEditValidator(app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgDelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsim.SimulateMsgDelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgUndelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsim.SimulateMsgUndelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgBeginRedelegate, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			stakingsim.SimulateMsgBeginRedelegate(app.accountKeeper, app.stakingKeeper),
		},
		{
			func(_ *rand.Rand) int {
				var v int
				ap.GetOrGenerate(cdc, OpWeightMsgUnjail, &v, nil,
					func(_ *rand.Rand) {
						v = 100
					})
				return v
			}(nil),
			slashingsim.SimulateMsgUnjail(app.slashingKeeper),
		},
	}
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

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/GaiaApp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	logger := log.NewNopLogger()

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ = sdk.NewLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		_ = os.RemoveAll(dir)
	}()
	app := NewGaiaApp(logger, db, nil, true, 0)

	// Run randomized simulation
	// TODO: parameterize numbers, save for a later PR
	_, params, simErr := simulation.SimulateFromSeed(getSimulateFromSeedInput(b, os.Stdout, app))

	// export state and params before the simulation error is checked
	if exportStatePath != "" {
		fmt.Println("Exporting app state...")
		appState, _, err := app.ExportAppStateAndValidators(false, nil)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
		err = ioutil.WriteFile(exportStatePath, []byte(appState), 0644)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if exportParamsPath != "" {
		fmt.Println("Exporting simulation params...")
		paramsBz, err := json.MarshalIndent(params, "", " ")
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}

		err = ioutil.WriteFile(exportParamsPath, paramsBz, 0644)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

	if commit {
		fmt.Println("\nGoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}
}

func TestFullAppSimulation(t *testing.T) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		_ = os.RemoveAll(dir)
	}()

	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, map[int64]bool{}, "", fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())

	// Run randomized simulation
	_, params, simErr := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))

	// export state and params before the simulation error is checked
	if exportStatePath != "" {
		fmt.Println("Exporting app state...")
		appState, _, err := app.ExportAppStateAndValidators(false, nil)
		require.NoError(t, err)

		err = ioutil.WriteFile(exportStatePath, []byte(appState), 0644)
		require.NoError(t, err)
	}

	if exportParamsPath != "" {
		fmt.Println("Exporting simulation params...")
		fmt.Println(params)
		paramsBz, err := json.MarshalIndent(params, "", " ")
		require.NoError(t, err)

		err = ioutil.WriteFile(exportParamsPath, paramsBz, 0644)
		require.NoError(t, err)
	}

	require.NoError(t, simErr)

	if commit {
		// for memdb:
		// fmt.Println("Database Size", db.Stats()["database.size"])
		fmt.Println("\nGoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}
}

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
		_ = os.RemoveAll(dir)
	}()

	app := NewGaiaApp(logger, db, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))

	// export state and simParams before the simulation error is checked
	if exportStatePath != "" {
		fmt.Println("Exporting app state...")
		appState, _, err := app.ExportAppStateAndValidators(false, nil)
		require.NoError(t, err)

		err = ioutil.WriteFile(exportStatePath, []byte(appState), 0644)
		require.NoError(t, err)
	}

	if exportParamsPath != "" {
		fmt.Println("Exporting simulation params...")
		simParamsBz, err := json.MarshalIndent(simParams, "", " ")
		require.NoError(t, err)

		err = ioutil.WriteFile(exportParamsPath, simParamsBz, 0644)
		require.NoError(t, err)
	}

	require.NoError(t, simErr)

	if commit {
		// for memdb:
		// fmt.Println("Database Size", db.Stats()["database.size"])
		fmt.Println("\nGoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}

	fmt.Printf("Exporting genesis...\n")

	appState, _, consensusParams, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	_, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", newApp.Name())

	var genesisState GenesisState
	err = app.Codec().UnmarshalJSON(appState, &genesisState)
	require.NoError(t, err)

	ctxB := newApp.NewContext(true, abci.Header{Height: app.LastBlockHeight()})
	newApp.mm.InitGenesis(ctxB, genesisState)

	fmt.Printf("Comparing stores...\n")
	ctxA := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	type StoreKeysPrefixes struct {
		A        sdk.StoreKey
		B        sdk.StoreKey
		Prefixes [][]byte
	}

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[baseapp.MainStoreKey], newApp.keys[baseapp.MainStoreKey], [][]byte{}},
		{app.keys[auth.StoreKey], newApp.keys[auth.StoreKey], [][]byte{}},
		{app.keys[staking.StoreKey], newApp.keys[staking.StoreKey],
			[][]byte{
				staking.UnbondingQueueKey, staking.RedelegationQueueKey, staking.ValidatorQueueKey,
			}}, // ordering may change but it doesn't matter
		{app.keys[slashing.StoreKey], newApp.keys[slashing.StoreKey], [][]byte{}},
		{app.keys[mint.StoreKey], newApp.keys[mint.StoreKey], [][]byte{}},
		{app.keys[distr.StoreKey], newApp.keys[distr.StoreKey], [][]byte{}},
		{app.keys[supply.StoreKey], newApp.keys[supply.StoreKey], [][]byte{}},
		{app.keys[params.StoreKey], newApp.keys[params.StoreKey], [][]byte{}},
		{app.keys[gov.StoreKey], newApp.keys[gov.StoreKey], [][]byte{}},
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
	stopEarly, params, simErr := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))

	// export state and params before the simulation error is checked
	if exportStatePath != "" {
		fmt.Println("Exporting app state...")
		appState, _, err := app.ExportAppStateAndValidators(false, nil)
		require.NoError(t, err)

		err = ioutil.WriteFile(exportStatePath, []byte(appState), 0644)
		require.NoError(t, err)
	}

	if exportParamsPath != "" {
		fmt.Println("Exporting simulation params...")
		paramsBz, err := json.MarshalIndent(params, "", " ")
		require.NoError(t, err)

		err = ioutil.WriteFile(exportParamsPath, paramsBz, 0644)
		require.NoError(t, err)
	}

	require.NoError(t, simErr)

	if commit {
		// for memdb:
		// fmt.Println("Database Size", db.Stats()["database.size"])
		fmt.Println("\nGoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
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

	// Run randomized simulation on imported app
	_, _, err = simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, newApp))
	require.Nil(t, err)
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
			app := NewGaiaApp(logger, db, nil, true, 0)

			// Run randomized simulation
			simulation.SimulateFromSeed(
				t, os.Stdout, app.BaseApp, appStateFn, seed, testAndRunTxs(app),
				[]sdk.Invariant{}, 1, numBlocks, exportParamsHeight,
				blockSize, "", false, commit, lean,
				false, false, app.ModuleAccountAddrs(),
			)
			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash
		}
		for k := 1; k < numTimesToRunPerSeed; k++ {
			require.Equal(t, appHashList[0], appHashList[k], "appHash list: %v", appHashList)
		}
	}
}

func BenchmarkInvariants(b *testing.B) {
	logger := log.NewNopLogger()
	dir, _ := ioutil.TempDir("", "goleveldb-app-invariant-bench")
	db, _ := sdk.NewLevelDB("simulation", dir)

			_, _, err := simulation.SimulateFromSeed(
				t, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
				simapp.SimulationOperations(app, app.Codec(), config),
				app.ModuleAccountAddrs(), config,
			)
			require.NoError(t, err)

	app := NewGaiaApp(logger, db, nil, true, 0)
	exportParams := exportParamsPath != ""

	// 2. Run parameterized simulation (w/o invariants)
	_, params, simErr := simulation.SimulateFromSeed(
		b, ioutil.Discard, app.BaseApp, appStateFn, seed, testAndRunTxs(app),
		[]sdk.Invariant{}, initialBlockHeight, numBlocks, exportParamsHeight, blockSize,
		exportStatsPath, exportParams, commit, lean, onOperation, false, app.ModuleAccountAddrs(),
	)

	// export state and params before the simulation error is checked
	if exportStatePath != "" {
		fmt.Println("Exporting app state...")
		appState, _, err := app.ExportAppStateAndValidators(false, nil)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
		err = ioutil.WriteFile(exportStatePath, []byte(appState), 0644)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if exportParamsPath != "" {
		fmt.Println("Exporting simulation params...")
		paramsBz, err := json.MarshalIndent(params, "", " ")
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}

		err = ioutil.WriteFile(exportParamsPath, paramsBz, 0644)
		if err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

			if j != 0 {
				require.Equal(
					t, appHashList[0], appHashList[j],
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
