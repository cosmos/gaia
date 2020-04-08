package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const port = "port_id"

// RadomEnabled randomized send or receive enabled param with 75% prob of being true.
func RadomEnabled(r *rand.Rand) bool {
	return r.Int63n(101) <= 75
}

// RandomizedGenState generates a random GenesisState for transfer.
func RandomizedGenState(simState *module.SimulationState) {
	consumingGenesis := types.GenesisState{}

	bz, err := json.MarshalIndent(consumingGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated %s parameters:\n%s\n", types.ModuleName, bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&consumingGenesis)
}
