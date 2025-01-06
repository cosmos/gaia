package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/gaia/v22/x/lsm/types"
)

// Simulation parameter constants
const (
	ValidatorBondFactor       = "validator_bond_factor"
	GlobalLiquidStakingCap    = "global_liquid_staking_cap"
	ValidatorLiquidStakingCap = "validator_liquid_staking_cap"
)

// getGlobalLiquidStakingCap returns randomized GlobalLiquidStakingCap between 0-1.
func getGlobalLiquidStakingCap(r *rand.Rand) sdkmath.LegacyDec {
	return simulation.RandomDecAmount(r, sdkmath.LegacyOneDec())
}

// getValidatorLiquidStakingCap returns randomized ValidatorLiquidStakingCap between 0-1.
func getValidatorLiquidStakingCap(r *rand.Rand) sdkmath.LegacyDec {
	return simulation.RandomDecAmount(r, sdkmath.LegacyOneDec())
}

// getValidatorBondFactor returns randomized ValidatorBondCap between -1 and 300.
func getValidatorBondFactor(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(simulation.RandIntBetween(r, -1, 300)))
}

// RandomizedGenState generates a random GenesisState for lsm
func RandomizedGenState(simState *module.SimulationState) {
	// params
	var (
		validatorBondFactor       sdkmath.LegacyDec
		globalLiquidStakingCap    sdkmath.LegacyDec
		validatorLiquidStakingCap sdkmath.LegacyDec
	)

	simState.AppParams.GetOrGenerate(ValidatorBondFactor, &validatorBondFactor, simState.Rand, func(r *rand.Rand) { validatorBondFactor = getValidatorBondFactor(r) })

	simState.AppParams.GetOrGenerate(GlobalLiquidStakingCap, &globalLiquidStakingCap, simState.Rand, func(r *rand.Rand) { globalLiquidStakingCap = getGlobalLiquidStakingCap(r) })

	simState.AppParams.GetOrGenerate(ValidatorLiquidStakingCap, &validatorLiquidStakingCap, simState.Rand, func(r *rand.Rand) { validatorLiquidStakingCap = getValidatorLiquidStakingCap(r) })

	params := types.NewParams(
		validatorBondFactor,
		globalLiquidStakingCap,
		validatorLiquidStakingCap,
	)

	lsmGenesis := types.NewGenesisState(params, nil, 0, sdkmath.ZeroInt(), nil)

	bz, err := json.MarshalIndent(&lsmGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated lsm parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(lsmGenesis)
}
