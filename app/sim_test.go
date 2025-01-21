package gaia_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulation2 "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	"github.com/cosmos/gaia/v23/ante"
	gaia "github.com/cosmos/gaia/v23/app"
	// "github.com/cosmos/gaia/v11/app/helpers"
	// "github.com/cosmos/gaia/v11/app/params"
	"github.com/cosmos/gaia/v23/app/sim"
)

// AppChainID hardcoded chainID for simulation
const AppChainID = "gaia-app"

func init() {
	sim.GetSimulatorFlags()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !sim.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	// since we can't provide tx fees to SimulateFromSeed(), we must switch off the feemarket
	ante.UseFeeMarketDecorator = false

	config := sim.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = AppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5

	// We will be overriding the random seed and just run a single simulation on the provided seed value
	if config.Seed != simcli.DefaultSeedValue {
		numSeeds = 1
	}

	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[server.FlagInvCheckPeriod] = sim.FlagPeriodValue

	for i := 0; i < numSeeds; i++ {
		if config.Seed == simcli.DefaultSeedValue {
			config.Seed = rand.Int63()
		}

		fmt.Println("config.Seed: ", config.Seed)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if sim.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			dir, err := os.MkdirTemp("", "gaia-simulation")
			require.NoError(t, err)
			appOptions[flags.FlagHome] = dir
			app := gaia.NewGaiaApp(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				dir,
				appOptions,
				emptyWasmOption,
				interBlockCacheOpt(),
				baseapp.SetChainID(AppChainID),
			)

			// NOTE: setting to zero to avoid failing the simulation
			// due to the minimum staked tokens required to submit a vote
			ante.SetMinStakedTokens(math.LegacyZeroDec())

			// NOTE: setting to zero to avoid failing the simulation
			// gaia ante allows only certain proposals to be expedited - the simulation doesn't know about this
			ante.SetExpeditedProposalsEnabled(false)

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			blockedAddresses := app.BlockedModuleAccountAddrs(app.ModuleAccountAddrs())

			_, _, err = simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.ModuleBasics.DefaultGenesis(app.AppCodec())),
				simulation2.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
				simtestutil.SimulationOperations(app, app.AppCodec(), config),
				blockedAddresses,
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				sim.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
