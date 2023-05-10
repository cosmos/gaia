package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMinInitDepositAmount),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMinInitDepositAmount(r).Int64())
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyInitPoolCoinMintAmount),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenInitPoolCoinMintAmount(r).Int64())
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxReserveCoinAmount),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%d\"", GenMaxReserveCoinAmount(r).Int64())
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySwapFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSwapFeeRate(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyWithdrawFeeRate),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenWithdrawFeeRate(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMaxOrderAmountRatio),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMaxOrderAmountRatio(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyUnitBatchHeight),
			func(r *rand.Rand) string {
				return fmt.Sprintf("%d", GenUnitBatchHeight(r))
			},
		),
	}
}
