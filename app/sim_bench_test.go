package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	config, db, dir, _, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		fmt.Println(err, "simulation setup failed")
		b.Fail()
	}

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err.Error())
			b.Fail()
		}
	}()

	logger := log.NewNopLogger()
	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	if err = simapp.CheckExportSimulation(app, config, simParams); err != nil {
		fmt.Println(err)
		b.Fail()
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

	if config.Commit {
		simapp.PrintStats(db)
	}
}

func BenchmarkInvariants(b *testing.B) {
	config, db, dir, _, _, err := simapp.SetupSimulation("leveldb-app-invariant-bench", "Simulation")
	if err != nil {
		fmt.Println(err, "simulation setup failed")
		b.Fail()
	}

	logger := log.NewNopLogger()
	config.AllInvariants = false

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err.Error())
			b.Fail()
		}
	}()

	app := NewGaiaApp(logger, db, nil, true, simapp.FlagPeriodValue, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b, os.Stdout, app.BaseApp, simapp.AppStateFn(app.Codec(), app.SimulationManager()),
		simapp.SimulationOperations(app, app.Codec(), config),
		app.ModuleAccountAddrs(), config,
	)

	// export state and simParams before the simulation error is checked
	if err = simapp.CheckExportSimulation(app, config, simParams); err != nil {
		fmt.Println(err)
		b.Fail()
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

	if config.Commit {
		simapp.PrintStats(db)
	}

	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.crisisKeeper.Routes() {
		cr := cr
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				fmt.Printf("broken invariant at block %d of %d\n%s", ctx.BlockHeight()-1, config.NumBlocks, res)
				b.FailNow()
			}
		})
	}
}
